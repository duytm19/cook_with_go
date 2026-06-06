package config

import (
	"os"
)

type Config struct {
	Port         string
	DatabaseURL  string
	AWSRegion    string
	AWSEndpoint  string
	AnvilRPCURL  string
	QueueURL     string
	DynamoDBTable string
}

func LoadConfig() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "host=localhost user=postgres password=1234567 dbname=relayer_db port=5432 sslmode=disable"),
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
		AWSEndpoint:  getEnv("AWS_ENDPOINT", "http://localhost:4566"),
		AnvilRPCURL:  getEnv("ANVIL_RPC_URL", "http://localhost:8545"),
		QueueURL:     getEnv("SQS_QUEUE_URL", "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/tx-delivery-queue"),
		DynamoDBTable: getEnv("DYNAMODB_TABLE", "relayer-state-cache"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
