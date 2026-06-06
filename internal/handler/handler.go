package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"cook_with_go/api"
	"cook_with_go/internal/repository"
	"cook_with_go/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIHandler struct {
	walletService *service.WalletService
	relayerService *service.TxRelayerService
	tenantRepo     repository.TenantRepository
}

func NewAPIHandler(wService *service.WalletService, rService *service.TxRelayerService, tRepo repository.TenantRepository) *APIHandler {
	return &APIHandler{
		walletService:  wService,
		relayerService: rService,
		tenantRepo:     tRepo,
	}
}

// Helper: Authenticate tenant using X-API-Key header
func (h *APIHandler) authenticateTenant(c *gin.Context) string {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header missing"})
		c.Abort()
		return ""
	}

	// Compute SHA-256 hash of API Key to check against database
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	tenant, err := h.tenantRepo.GetByAPIKeyHash(c.Request.Context(), keyHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
		c.Abort()
		return ""
	}

	return tenant.ID.String()
}

// POST /v1/wallets
func (h *APIHandler) CreateWallet(c *gin.Context) {
	tenantIDStr := h.authenticateTenant(c)
	if tenantIDStr == "" {
		return // context aborted in authenticator
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal tenant ID mapping error"})
		return
	}

	wallet, err := h.walletService.CreateCustodialWallet(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, api.Wallet{
		WalletId: wallet.ID,
		Address:  wallet.Address,
	})
}

// POST /v1/transactions/send
func (h *APIHandler) SendTransaction(c *gin.Context) {
	tenantIDStr := h.authenticateTenant(c)
	if tenantIDStr == "" {
		return // context aborted in authenticator
	}

	var req api.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Queue transaction in Postgres + SQS
	tx, err := h.relayerService.QueueTransaction(c.Request.Context(), req.WalletId, req.ToAddress, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, api.TransactionResponse{
		TxId:   tx.ID,
		Status: api.QUEUED,
	})
}
