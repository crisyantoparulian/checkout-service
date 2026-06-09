package product

import "github.com/crisyantoparulian/checkout-service/internal/app/usecase/product"

type handler struct {
	productUsecase product.ProductUsecase
}

func NewHandler() *handler {
	return &handler{}
}

func (h *handler) SetProductUsecase(usecase product.ProductUsecase) *handler {
	h.productUsecase = usecase
	return h
}

func (h *handler) Validate() *handler {
	if h.productUsecase == nil {
		panic("productUsecase is nil")
	}
	return h
}
