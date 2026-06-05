#!/bin/sh

echo "Initializing SQS Queues..."

# 1. Create Dead Letter Queue (DLQ)
awslocal sqs create-queue --queue-name tx-delivery-dlq

# Get the DLQ ARN
DLQ_ARN=$(awslocal sqs get-queue-attributes \
    --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/tx-delivery-dlq \
    --attribute-names QueueArn | grep QueueArn | awk -F '"' '{print $4}')

echo "DLQ ARN found: $DLQ_ARN"

# 2. Create Main Transaction Delivery Queue with Redrive Policy (3 retries)
awslocal sqs create-queue \
    --queue-name tx-delivery-queue \
    --attributes '{
        "RedrivePolicy": "{\"deadLetterTargetArn\":\"'"$DLQ_ARN"'\",\"maxReceiveCount\":3}"
    }'

echo "SQS Queues initialized successfully!"

# 3. Create DynamoDB Table for Nonces and Transaction Cache
echo "Initializing DynamoDB Tables..."
awslocal dynamodb create-table \
    --table-name relayer-state-cache \
    --attribute-definitions AttributeName=PK,AttributeType=S \
    --key-schema AttributeName=PK,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST

echo "DynamoDB Tables initialized successfully!"
