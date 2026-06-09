package checkout

import "github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout"

type handler struct {
	checkoutUsecase checkout.CheckoutUsecase
}

func NewHandler() *handler {
	return &handler{}
}

func (h *handler) SetCheckoutUsecase(usecase checkout.CheckoutUsecase) *handler {
	h.checkoutUsecase = usecase
	return h
}

func (h *handler) Validate() *handler {
	if h.checkoutUsecase == nil {
		panic("checkoutUsecase is nil")
	}
	return h
}
