package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"fmt"

	"cook_with_go/internal/model"
	"cook_with_go/internal/repository"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

type WalletService struct {
	repo      repository.WalletRepository
	kmsClient *kms.Client
}

func NewWalletService(repo repository.WalletRepository, kmsClient *kms.Client) *WalletService {
	return &WalletService{
		repo:      repo,
		kmsClient: kmsClient,
	}
}

func (s *WalletService) CreateCustodialWallet(ctx context.Context, tenantID uuid.UUID) (*model.Wallet, error) {
	// 1. Create a key inside AWS KMS using ECC secp256k1 curve
	out, err := s.kmsClient.CreateKey(ctx, &kms.CreateKeyInput{
		KeySpec:     types.KeySpecEccSecgP256k1,
		KeyUsage:    types.KeyUsageTypeSignVerify,
		Description: aws.String(fmt.Sprintf("Custodial EVM key for Tenant: %s", tenantID.String())),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create key in AWS KMS: %w", err)
	}
	keyID := *out.KeyMetadata.KeyId

	// 2. Fetch the public key from KMS to compute the Ethereum address
	pubOut, err := s.kmsClient.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: out.KeyMetadata.KeyId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key from KMS: %w", err)
	}

	// 3. Decode the PKIX DER bytes
	parsedPubKey, err := x509.ParsePKIXPublicKey(pubOut.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}

	ecdsaPubKey, ok := parsedPubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("kms public key is not an ECDSA key")
	}

	// 4. Derive the Ethereum address
	address := crypto.PubkeyToAddress(*ecdsaPubKey).Hex()

	// 5. Save the wallet details to Postgres
	wallet := &model.Wallet{
		ID:       uuid.New(),
		TenantID: tenantID,
		Address:  address,
		KMSKeyID: keyID,
	}

	if err := s.repo.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to save wallet metadata to db: %w", err)
	}

	return wallet, nil
}

func (s *WalletService) GetWalletByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	return s.repo.GetByID(ctx, id)
}
