package repository

import (
	"context"

	"cook_with_go/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepository interface {
	GetByAPIKeyHash(ctx context.Context, hash string) (*model.Tenant, error)
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *model.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]model.Wallet, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	Update(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
}

// tenantRepository implementation
type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) GetByAPIKeyHash(ctx context.Context, hash string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.WithContext(ctx).Where("api_key_hash = ?", hash).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// walletRepository implementation
type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *model.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]model.Wallet, error) {
	var wallets []model.Wallet
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&wallets).Error
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

// transactionRepository implementation
type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepository) Update(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
