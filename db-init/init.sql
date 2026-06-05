-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tenants/B2B Clients Table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    api_key_hash VARCHAR(64) NOT NULL UNIQUE, -- For authentication through gateway/backend
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Custodial Wallets Table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL UNIQUE, -- EVM 42-char address (0x...)
    kms_key_id VARCHAR(255) NOT NULL UNIQUE, -- AWS KMS Key ID/ARN reference
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Transactions Log Table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    tx_hash VARCHAR(66) UNIQUE, -- 66-char transaction hash (0x...)
    to_address VARCHAR(42) NOT NULL,
    data TEXT NOT NULL, -- Hex payload
    status VARCHAR(20) NOT NULL DEFAULT 'QUEUED', -- QUEUED, BROADCASTED, MINED, FAILED
    nonce BIGINT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed a test merchant client for local testing
-- API Key is: "test_merchant_secret_key" (SHA-256 hash is seeded below)
INSERT INTO tenants (id, name, api_key_hash)
VALUES (
    'a3e26cb3-91b5-4b13-8cfb-ebad10101010', 
    'Test Merchant LLC', 
    'f35e806c9e0ff5b0d01d4a0f443813ff38c645b206cd2df3b018bdf3b0222ebd'
) ON CONFLICT DO NOTHING;
