.PHONY: build build-alpine clean test help default rest migrate-up mockgen-install mock-repository mock-transaction mock-usecase mock test-integration

GOTEST := go test -v
PACKAGES := $(shell go list ./internal/app/usecase/... ./internal/app/repository/... ./internal/pkg/... | grep -v /mocks)

test:
	@echo "=================================================================================="
	@echo "Coverage Test"
	@echo "=================================================================================="
	go fmt ./... && $(GOTEST) -coverprofile coverage.cov -cover ${PACKAGES}
	@echo "\n"
	@echo "=================================================================================="
	@echo "All Package Coverage"
	@echo "=================================================================================="
	go tool cover -func coverage.cov

mockgen-install:
	go install go.uber.org/mock/mockgen@latest

mock-repository:
	mockgen -destination=internal/app/repository/mocks/checkout.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/repository Checkout
	mockgen -destination=internal/app/repository/mocks/product.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/repository Product
	mockgen -destination=internal/app/repository/mocks/promotion.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/repository Promotion
	mockgen -destination=internal/app/repository/mocks/inventory.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/repository Inventory
	mockgen -destination=internal/app/repository/mocks/health_check.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/repository HealthCheck

mock-transaction:
	mockgen -destination=internal/pkg/transaction/mocks/manager.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/pkg/transaction ManagerInterface

mock-usecase:
	mockgen -destination=internal/app/usecase/health_check/mocks/usecase.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/usecase/health_check HealthCheckUsecase
	mockgen -destination=internal/app/usecase/product/mocks/usecase.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/usecase/product ProductUsecase
	mockgen -destination=internal/app/usecase/checkout/mocks/usecase.go -package=mocks github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout CheckoutUsecase

mock: mock-repository mock-transaction mock-usecase

test-integration:
	@echo "=================================================================================="
	@echo "Integration Tests"
	@echo "=================================================================================="
	go test -v -tags=integration ./tests/integration/...

swag-install:
	go install github.com/swaggo/swag/cmd/swag@latest
	
swag-init:
	swag init --parseDependency --parseInternal --parseDepth 1 --overridesFile .swaggo

rest:
	go run main.go server rest

migrate-up:
	go run main.go migrate up