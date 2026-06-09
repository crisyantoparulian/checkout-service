package product

import (
	"github.com/crisyantoparulian/checkout-service/internal/app/usecase/product"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/response"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/validator"
	"github.com/labstack/echo/v4"
)

// @Tags Products
// @Summary Get All Products
// @Description API for get all products
// @Produce json
// @Param page query int false "page" default(1)
// @Param per_page query int false "per_page" default(20)
// @Success 200 {object} response.DefaultResponse{data=product.GetAllProductResponse}
// @Failure 400 {object} response.ErrorResponse{data=nil}
// @Failure 500 {object} response.ErrorResponse{data=nil}
// @Router /products [get]
func (h *handler) GetAll(c echo.Context) (err error) {
	ctx := c.Request().Context()

	req := product.GetAllProductRequest{}
	if err = validator.Validate(c, &req); err != nil {
		return
	}

	resp, err := h.productUsecase.GetAll(ctx, req)
	if err != nil {
		return
	}

	return response.Success(c, resp)
}
