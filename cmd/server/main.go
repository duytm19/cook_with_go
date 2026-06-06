package main

import (
	"context"
	"log"
	"net/http"

	"cook_with_go/api"
	"cook_with_go/internal/config"
	"cook_with_go/internal/handler"
	"cook_with_go/internal/repository"
	"cook_with_go/internal/service"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Web3 Relayer Engine API...")

	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Database Connection (GORM)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("Database connection established successfully.")

	// 3. Initialize AWS SDK v2 Configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		log.Fatalf("failed to load AWS default config: %v", err)
	}

	// Configure KMS & SQS clients to point to custom LocalStack endpoints if running locally
	kmsClient := kms.NewFromConfig(awsCfg, func(o *kms.Options) {
		if cfg.AWSEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AWSEndpoint)
		}
	})

	sqsClient := sqs.NewFromConfig(awsCfg, func(o *sqs.Options) {
		if cfg.AWSEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AWSEndpoint)
		}
	})

	log.Println("AWS service clients initialized.")

	// 4. Initialize Repositories and Services
	tenantRepo := repository.NewTenantRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	walletService := service.NewWalletService(walletRepo, kmsClient)
	relayerService := service.NewTxRelayerService(txRepo, walletRepo, sqsClient, cfg.QueueURL)

	// 5. Initialize API Handler
	apiHandler := handler.NewAPIHandler(walletService, relayerService, tenantRepo)

	// 6. Setup Gin Router
	router := gin.Default()

	// Register OpenAPI health/ping checks
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Register generated handlers from OpenAPI spec
	api.RegisterHandlers(router, apiHandler)

	log.Printf("Listening and serving HTTP on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run Gin server: %v", err)
	}
}
