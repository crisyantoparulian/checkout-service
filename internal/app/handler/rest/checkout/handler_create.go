package checkout

import (
	"github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/response"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/validator"
	"github.com/labstack/echo/v4"
)

// @Tags Checkouts
// @Summary Create Checkout
// @Description API for checkout products
// @Accept json
// @Produce json
// @Param request body checkout.CreateCheckoutRequest true "checkout request"
// @Success 200 {object} response.DefaultResponse{data=checkout.CheckoutResponse}
// @Failure 400 {object} response.ErrorResponse{data=nil}
// @Failure 404 {object} response.ErrorResponse{data=nil}
// @Failure 409 {object} response.ErrorResponse{data=nil}
// @Failure 500 {object} response.ErrorResponse{data=nil}
// @Router /checkouts [post]
func (h *handler) Create(c echo.Context) (err error) {
	ctx := c.Request().Context()

	req := checkout.CreateCheckoutRequest{}
	if err = validator.Validate(c, &req); err != nil {
		return
	}

	resp, err := h.checkoutUsecase.Create(ctx, req)
	if err != nil {
		return
	}

	return response.Success(c, resp)
}
