#!/bin/sh
# Create SQS Queue
awslocal sqs create-queue --queue-name tx-delivery-queue

# Create DynamoDB table for nonces and transaction caching
awslocal dynamodb create-table \
    --table-name relayer-state-cache \
    --attribute-definitions AttributeName=PK,AttributeType=S \
    --key-schema AttributeName=PK,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST
