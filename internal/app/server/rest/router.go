package rest

import (
	"github.com/crisyantoparulian/checkout-service/internal/app/container"
	"github.com/crisyantoparulian/checkout-service/internal/app/handler/rest/checkout"
	"github.com/crisyantoparulian/checkout-service/internal/app/handler/rest/health_check"
	"github.com/crisyantoparulian/checkout-service/internal/app/handler/rest/product"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRouter(server *echo.Echo, container *container.Container) {
	// inject handler with usecase via container
	healthCheckHandler := health_check.NewHandler().SetHealthCheckUsecase(container.HealthCheckUsecase).Validate()
	productHandler := product.NewHandler().SetProductUsecase(container.ProductUsecase).Validate()
	checkoutHandler := checkout.NewHandler().SetCheckoutUsecase(container.CheckoutUsecase).Validate()

	server.GET("/health", healthCheckHandler.HealthCheck)
	server.GET("/swagger/*", echoSwagger.WrapHandler)

	products := server.Group("/products")
	{
		products.GET("", productHandler.GetAll)
	}

	checkouts := server.Group("/checkouts")
	{
		checkouts.POST("", checkoutHandler.Create)
		checkouts.GET("/:uuid", checkoutHandler.GetDetail)
	}
}
