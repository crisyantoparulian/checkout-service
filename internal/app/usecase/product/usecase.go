package product

import (
	"context"
	"fmt"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/pagination"
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

func (u *usecase) GetAll(ctx context.Context, req GetAllProductRequest) (resp GetAllProductResponse, err error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PerPage == 0 {
		req.PerPage = 20
	}

	filter := repository.ProductFilter{
		Limit:  req.PerPage,
		Offset: (req.Page - 1) * req.PerPage,
	}

	products, err := u.productRepository.GetAll(ctx, filter)
	if err != nil {
		return
	}

	total, err := u.productRepository.CountTotal(ctx)
	if err != nil {
		return
	}

	resp.Products = []ProductResponse{}
	for _, product := range products {
		resp.Products = append(resp.Products, ProductResponse{
			UUID:           product.UUID.String(),
			SKU:            product.SKU,
			Name:           product.Name,
			PriceCents:     product.PriceCents,
			FormattedPrice: formatMoney(product.PriceCents),
			InventoryQty:   product.InventoryQty,
		})
	}
	resp.Pagination = pagination.GeneratePaginationResponse(req.PerPage, req.Page, total)
	return
}

func formatMoney(cents int64) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}
