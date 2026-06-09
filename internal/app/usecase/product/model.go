package product

import "github.com/crisyantoparulian/checkout-service/internal/pkg/pagination"

type GetAllProductRequest struct {
	Page    int `query:"page" validate:"omitempty,gte=1"`
	PerPage int `query:"per_page" validate:"omitempty,gte=1,lte=100"`
}

type ProductResponse struct {
	UUID           string `json:"uuid"`
	SKU            string `json:"sku"`
	Name           string `json:"name"`
	PriceCents     int64  `json:"price_cents"`
	FormattedPrice string `json:"formatted_price"`
	InventoryQty   int    `json:"inventory_qty"`
}

type GetAllProductResponse struct {
	Products   []ProductResponse             `json:"products"`
	Pagination pagination.PaginationResponse `json:"pagination"`
}
