package model

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string    `gorm:"not null;unique"`
	APIKeyHash string    `gorm:"not null;unique"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Wallets    []Wallet  `gorm:"foreignKey:TenantID"`
}

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null"`
	Address   string    `gorm:"not null;unique"`
	KMSKeyID  string    `gorm:"column:kms_key_id;not null;unique"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Transaction struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletID     uuid.UUID `gorm:"type:uuid;not null"`
	TxHash       *string   `gorm:"column:tx_hash;unique"`
	ToAddress    string    `gorm:"column:to_address;not null"`
	Data         string    `gorm:"not null"`
	Status       string    `gorm:"not null;default:'QUEUED'"`
	Nonce        int64     `gorm:"not null"`
	ErrorMessage *string   `gorm:"column:error_message"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
