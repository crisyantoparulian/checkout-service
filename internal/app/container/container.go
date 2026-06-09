package container

import (
	"github.com/crisyantoparulian/checkout-service/config"
	"github.com/crisyantoparulian/checkout-service/internal/app/driver"
	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout"
	"github.com/crisyantoparulian/checkout-service/internal/app/usecase/health_check"
	"github.com/crisyantoparulian/checkout-service/internal/app/usecase/product"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/transaction"
	"github.com/insaneadinesia/gobang/gotel"
	"github.com/insaneadinesia/gobang/logger"
)

type Container struct {
	Config             config.Config
	HealthCheckUsecase health_check.HealthCheckUsecase
	ProductUsecase     product.ProductUsecase
	CheckoutUsecase    checkout.CheckoutUsecase
}

func Setup() *Container {
	// Load Config
	cfg := config.Load()

	// Setup Driver
	db, _ := driver.NewPostgresDatabase(cfg)

	// Setup Tools
	logger.NewLogger(logger.Option{
		IsEnable:            cfg.LoggerEnable,
		EnableStackTrace:    cfg.LoggerEnableStackTrace,
		EnableMaskingFields: cfg.LoggerEnableMasking,
		MaskingFields:       cfg.LoggerMaskingFields,
	})

	gotel.NewOtelWithJaegerExporter(cfg.AppName, gotel.OtelWithJaegerOption{
		Endpoint: cfg.JaegerEndpoint,
	})

	// Setup Repository
	healthCheckRepository := repository.NewHealthCheckRepository(db)
	productRepository := repository.NewProductRepository(db)
	promotionRepository := repository.NewPromotionRepository(db)
	checkoutRepository := repository.NewCheckoutRepository(db)
	inventoryRepository := repository.NewInventoryRepository(db)

	// Setup Transaction Manager
	txManager := transaction.NewManager(db)

	// Setup Usecase
	healthCheckUsecase := health_check.NewUsecase().SetHealthCheckRepository(healthCheckRepository).Validate()
	productUsecase := product.NewUsecase().SetProductRepository(productRepository).Validate()
	checkoutUsecase := checkout.NewUsecase().
		SetTxManager(txManager).
		SetCheckoutRepository(checkoutRepository).
		SetProductRepository(productRepository).
		SetPromotionRepository(promotionRepository).
		SetInventoryRepository(inventoryRepository).
		Validate()

	return &Container{
		Config:             cfg,
		HealthCheckUsecase: healthCheckUsecase,
		ProductUsecase:     productUsecase,
		CheckoutUsecase:    checkoutUsecase,
	}
}
