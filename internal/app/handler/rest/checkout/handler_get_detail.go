package checkout

import (
	"github.com/crisyantoparulian/checkout-service/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// @Tags Checkouts
// @Summary Get Checkout Detail
// @Description API for get checkout detail
// @Produce json
// @Param uuid path string true "checkout uuid"
// @Success 200 {object} response.DefaultResponse{data=checkout.CheckoutResponse}
// @Failure 404 {object} response.ErrorResponse{data=nil}
// @Failure 500 {object} response.ErrorResponse{data=nil}
// @Router /checkouts/{uuid} [get]
func (h *handler) GetDetail(c echo.Context) (err error) {
	ctx := c.Request().Context()

	resp, err := h.checkoutUsecase.GetDetail(ctx, c.Param("uuid"))
	if err != nil {
		return
	}

	return response.Success(c, resp)
}
