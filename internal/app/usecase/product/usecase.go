package product

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
)

type ProductUsecase interface {
	GetAll(ctx context.Context, req GetAllProductRequest) (GetAllProductResponse, error)
}

type usecase struct {
	productRepository repository.Product
}

func NewUsecase() *usecase {
	return &usecase{}
}

func (u *usecase) SetProductRepository(repo repository.Product) *usecase {
	u.productRepository = repo
	return u
}

func (u *usecase) Validate() ProductUsecase {
	if u.productRepository == nil {
		panic("productRepository is nil")
	}
	return u
}
