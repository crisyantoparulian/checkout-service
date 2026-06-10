package checkout

import (
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapCheckoutResponse(t *testing.T) {
	checkoutUUID := uuid.New()

	tests := []struct {
		name           string
		checkout       entity.Checkout
		items          []entity.CheckoutItem
		promotions     []appliedPromotion
		wantItemCount  int
		wantPromoCount int
		wantFormatted  string
	}{
		{
			name: "happy path with items and promotions",
			checkout: entity.Checkout{
				UUID:          checkoutUUID,
				Status:        statusCompleted,
				SubtotalCents: 542999,
				DiscountCents: 3000,
				TotalCents:    539999,
			},
			items: []entity.CheckoutItem{
				{ProductUUID: testMacBookUUID, SKU: "43N23P", ProductName: "MacBook Pro", Quantity: 1, UnitPriceCents: 539999, TotalPriceCents: 539999},
				{ProductUUID: testRaspberryUUID, SKU: "234234", ProductName: "Raspberry Pi B", Quantity: 1, UnitPriceCents: 3000, TotalPriceCents: 3000},
			},
			promotions: []appliedPromotion{
				{PromotionCode: "MACBOOK_FREE_RASPBERRY", PromotionName: "MacBook Pro Free Raspberry Pi B", PromotionType: constants.PromotionTypeFreeProduct, DiscountCents: 3000},
			},
			wantItemCount:  2,
			wantPromoCount: 1,
			wantFormatted:  "$5399.99",
		},
		{
			name: "empty items and promotions",
			checkout: entity.Checkout{
				UUID:       checkoutUUID,
				Status:     statusCompleted,
				TotalCents: 0,
			},
			items:          nil,
			promotions:     nil,
			wantItemCount:  0,
			wantPromoCount: 0,
			wantFormatted:  "$0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mapCheckoutResponse(tt.checkout, tt.items, tt.promotions)

			assert.Equal(t, checkoutUUID.String(), resp.CheckoutID)
			assert.Equal(t, tt.checkout.Status, resp.Status)
			assert.Equal(t, tt.checkout.SubtotalCents, resp.SubtotalCents)
			assert.Equal(t, tt.checkout.DiscountCents, resp.DiscountCents)
			assert.Equal(t, tt.checkout.TotalCents, resp.TotalCents)
			assert.Equal(t, tt.wantFormatted, resp.FormattedTotal)
			require.Len(t, resp.Items, tt.wantItemCount)
			require.Len(t, resp.AppliedPromotions, tt.wantPromoCount)

			// Verify items are never nil (empty slices)
			assert.NotNil(t, resp.Items)
			assert.NotNil(t, resp.AppliedPromotions)
		})
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{name: "zero", cents: 0, want: "$0.00"},
		{name: "single digit cents", cents: 1050, want: "$10.50"},
		{name: "two digit cents", cents: 4999, want: "$49.99"},
		{name: "one digit cent", cents: 3005, want: "$30.05"},
		{name: "large value", cents: 539999, want: "$5399.99"},
		{name: "one cent", cents: 1, want: "$0.01"},
		{name: "99 cents", cents: 99, want: "$0.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatMoney(tt.cents))
		})
	}
}
