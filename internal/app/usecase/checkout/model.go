package checkout

type CreateCheckoutRequest struct {
	Items []CheckoutItemRequest `json:"items" validate:"required,min=1,dive"`
}

type CheckoutItemRequest struct {
	ProductUUID string `json:"product_uuid" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
}

type CheckoutItemResponse struct {
	ProductUUID     string `json:"product_uuid"`
	SKU             string `json:"sku"`
	Name            string `json:"name"`
	Quantity        int    `json:"quantity"`
	UnitPriceCents  int64  `json:"unit_price_cents"`
	TotalPriceCents int64  `json:"total_price_cents"`
}

type AppliedPromotionResponse struct {
	PromotionCode string `json:"promotion_code"`
	PromotionName string `json:"promotion_name"`
	PromotionType string `json:"promotion_type"`
	DiscountCents int64  `json:"discount_cents"`
}

type CheckoutResponse struct {
	CheckoutID        string                     `json:"checkout_id"`
	Status            string                     `json:"status"`
	Items             []CheckoutItemResponse     `json:"items"`
	SubtotalCents     int64                      `json:"subtotal_cents"`
	DiscountCents     int64                      `json:"discount_cents"`
	TotalCents        int64                      `json:"total_cents"`
	FormattedTotal    string                     `json:"formatted_total"`
	AppliedPromotions []AppliedPromotionResponse `json:"applied_promotions"`
	CreatedAt         string                     `json:"created_at,omitempty"`
}

type appliedPromotion struct {
	PromotionRuleUUID string
	PromotionUUID     string
	PromotionCode     string
	PromotionName     string
	PromotionType     string
	DiscountCents     int64
	RuleSnapshot      []byte
}
