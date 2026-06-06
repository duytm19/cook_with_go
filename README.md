# Web3 Custodial Wallet & Transaction Relayer Engine

An enterprise-grade, high-performance B2B middleware designed to abstract blockchain complexity, manage HSM-backed custodial wallets, and guarantee atomic transaction delivery. 

This engine is built with **Go (Gin)**, **GORM (Postgres)**, **AWS SDK v2 (KMS, SQS, DynamoDB)**, and **Kong Gateway**.

---

## Key Features

*   **Asymmetric KMS Wallets**: Cryptographic key pairs generated directly inside AWS KMS (ECC secp256k1 curve). Private keys never leave the HSM. Addresses are derived on-the-fly.
*   **Asynchronous Relaying**: Fast API ingestion. Client requests are validated, saved in PostgreSQL as `QUEUED`, and pushed to AWS SQS in $< 150$ms.
*   **API Gateway Ingress**: Kong Gateway manages rate-limiting, WAF policies, and handles client authentication via `X-API-Key` headers.
*   **Reliability & DLQ**: Transactions that fail to broadcast are automatically retried up to 3 times before being routed to the SQS Dead Letter Queue (DLQ) for alerting.
*   **Distributed Nonce Locks**: Atomic incrementing and tracking of EVM nonces inside AWS DynamoDB to prevent race conditions during high-concurrency broadcasts.

---

## Tech Stack

*   **Backend**: Go (Gin framework), GORM
*   **Blockchain**: `go-ethereum` client library
*   **API Ingress**: Kong API Gateway (DB-less mode)
*   **Database**: PostgreSQL 16 (Relational Metadata)
*   **Local Cloud Infrastructure**: LocalStack (mocking SQS, KMS, and DynamoDB)
*   **Local Blockchain Emulator**: Anvil (Foundry EVM node)

---

## Directory Layout

```plaintext
cook_with_go/
├── cmd/
│   └── server/
│       └── main.go         # Application bootstrap & DI setup
├── api/
│   ├── openapi.yaml        # API contract (OpenAPI 3.0)
│   └── api.gen.go          # Boilerplate routing (generated via oapi-codegen)
├── internal/
│   ├── config/             # Config loader (ENV parsing)
│   ├── model/              # GORM database schemas
│   ├── repository/         # Database operations
│   ├── service/            # Core business logic (KMS, SQS, Relayer)
│   └── handler/            # API endpoints & Cognito/API-Key Auth middleware
├── db-init/
│   └── init.sql            # PostgreSQL schema & Seed merchant data
├── localstack-init/
│   └── init.sh             # SQS DLQ & DynamoDB schema initializer
├── docker-compose.yml      # Sandbox configuration
└── kong.yml                # API Gateway routing rules
```

---

## Getting Started

### Prerequisites
Make sure you have the following installed locally:
*   [Go 1.21+](https://go.dev/doc/install)
*   [Docker & Docker Compose](https://docs.docker.com/engine/install/)
*   [Foundry (Anvil)](https://book.getfoundry.sh/getting-started/installation) (Optional: runs inside Docker as well)

---

### Step 1: Start the Local Sandbox
Spin up LocalStack, Postgres, Anvil, and Kong. Provide your `LOCALSTACK_AUTH_TOKEN` (required for LocalStack v4+):

```bash
# Option A: Via environment variable
export LOCALSTACK_AUTH_TOKEN="your_token_here"
docker compose up -d

# Option B: Or create a .env file in the cook_with_go directory
LOCALSTACK_AUTH_TOKEN=your_token_here
docker compose up -d
```

On startup:
1.  **Postgres** automatically runs `db-init/init.sql` to build the schemas and seed a test merchant.
2.  **LocalStack** runs `localstack-init/init.sh` to initialize the DynamoDB state cache table, SQS message queues, and configure the Dead Letter Queue redrive policy.
3.  **Kong** configures its proxy to route traffic from `http://localhost:8000/v1` to your local Go API.

---

### Step 2: Build & Start the Go Relayer API
Run the Go Gin API server on your host machine:

```bash
# Install package dependencies
go mod tidy

# Compile the server
go build -o bin/server cmd/server/main.go

# Start the server (defaults to port 8080)
./bin/server
```

---

## API Verification & Usage Guide

We seeded a test merchant in Postgres for local testing:
*   **Merchant Name**: `Test Merchant LLC`
*   **Merchant API Key**: `test_merchant_secret_key`

You must pass the API Key in the `X-API-Key` header. Requests are routed through the **Kong Gateway on port 8000**.

### 1. Create a Custodial Wallet
Request the relayer to create a new custodial wallet. The API requests KMS to generate a key pair and returns the public address.

```bash
curl -X POST http://localhost:8000/v1/wallets \
  -H "X-API-Key: test_merchant_secret_key"
```

**Example Response:**
```json
{
  "wallet_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3d4b1d",
  "address": "0x51E2812A27a56475854b7B1dD45C4b8b26cd2dF3"
}
```

### 2. Send a Relayed/Gasless Transaction
Request the relayer to sign and send a transaction asynchronously. Replace `wallet_id` with the UUID returned from the previous step.

```bash
curl -X POST http://localhost:8000/v1/transactions/send \
  -H "X-API-Key: test_merchant_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3d4b1d",
    "to_address": "0xdecafbad00000000000000000000000000000000",
    "data": "0x"
  }'
```

**Example Response:**
```json
{
  "tx_id": "e3e26cb3-91b5-4b13-8cfb-ebad10101010",
  "status": "QUEUED"
}
```

*The transaction request is now stored in PostgreSQL as `QUEUED` and a processing payload has been pushed to the SQS queue.*

---

## Non-Functional Requirements & Security Checklist

- [x] **HSM Isolation**: Keys are kept strictly within AWS KMS. Signing is done inside KMS.
- [x] **Network Security**: Private subnets utilized with VPC endpoints/PrivateLink.
- [x] **Gateway Rate-limiting**: Rate limiting is configured at Kong Gateway.
- [x] **DLQ Trigger Alerting**: Redrive policies automatically isolate stuck broadcasts.
