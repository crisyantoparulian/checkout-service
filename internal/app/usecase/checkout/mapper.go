package checkout

import (
	"fmt"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
)

func mapCheckoutResponse(checkout entity.Checkout, items []entity.CheckoutItem, promotions []appliedPromotion) CheckoutResponse {
	resp := CheckoutResponse{
		CheckoutID:        checkout.UUID.String(),
		Status:            checkout.Status,
		SubtotalCents:     checkout.SubtotalCents,
		DiscountCents:     checkout.DiscountCents,
		TotalCents:        checkout.TotalCents,
		FormattedTotal:    formatMoney(checkout.TotalCents),
		Items:             []CheckoutItemResponse{},
		AppliedPromotions: []AppliedPromotionResponse{},
	}
	for _, item := range items {
		resp.Items = append(resp.Items, CheckoutItemResponse{
			SKU:             item.SKU,
			Name:            item.ProductName,
			Quantity:        item.Quantity,
			UnitPriceCents:  item.UnitPriceCents,
			TotalPriceCents: item.TotalPriceCents,
		})
	}
	for _, promotion := range promotions {
		resp.AppliedPromotions = append(resp.AppliedPromotions, AppliedPromotionResponse{
			PromotionCode: promotion.PromotionCode,
			PromotionName: promotion.PromotionName,
			PromotionType: promotion.PromotionType,
			DiscountCents: promotion.DiscountCents,
		})
	}
	return resp
}

func formatMoney(cents int64) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}
