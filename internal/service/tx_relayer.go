package service

import (
	"context"
	"encoding/json"
	"fmt"

	"cook_with_go/internal/model"
	"cook_with_go/internal/repository"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

type TxRelayerService struct {
	txRepo    repository.TransactionRepository
	wRepo     repository.WalletRepository
	sqsClient *sqs.Client
	queueURL  string
}

type SQSMessagePayload struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	WalletID      uuid.UUID `json:"wallet_id"`
	ToAddress     string    `json:"to_address"`
	Data          string    `json:"data"`
}

func NewTxRelayerService(txRepo repository.TransactionRepository, wRepo repository.WalletRepository, sqsClient *sqs.Client, queueURL string) *TxRelayerService {
	return &TxRelayerService{
		txRepo:    txRepo,
		wRepo:     wRepo,
		sqsClient: sqsClient,
		queueURL:  queueURL,
	}
}

func (s *TxRelayerService) QueueTransaction(ctx context.Context, walletID uuid.UUID, toAddress string, hexData string) (*model.Transaction, error) {
	// 1. Verify that the wallet exists
	wallet, err := s.wRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	txID := uuid.New()

	// 2. Create the transaction record in the database as QUEUED
	// Note: Nonce will be allocated atomically when the worker processes the transaction.
	tx := &model.Transaction{
		ID:        txID,
		WalletID:  wallet.ID,
		ToAddress: toAddress,
		Data:      hexData,
		Status:    "QUEUED",
		Nonce:     0, // Will be set by the background worker
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction in db: %w", err)
	}

	// 3. Serialize transaction payload for SQS
	payload := SQSMessagePayload{
		TransactionID: tx.ID,
		WalletID:      wallet.ID,
		ToAddress:     toAddress,
		Data:          hexData,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction message: %w", err)
	}

	// 4. Send the message to SQS
	_, err = s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(bodyBytes)),
	})
	if err != nil {
		// If SQS fails, mark the transaction as failed in the database
		failMsg := "SQS dispatch failed"
		tx.Status = "FAILED"
		tx.ErrorMessage = &failMsg
		_ = s.txRepo.Update(ctx, tx)
		return nil, fmt.Errorf("failed to queue transaction to SQS: %w", err)
	}

	return tx, nil
}

func (s *TxRelayerService) GetTransactionByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}
