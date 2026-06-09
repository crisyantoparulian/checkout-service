package checkout

import (
	"database/sql"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testProducts = map[string]entity.Product{
	"120P90": {SKU: "120P90", Name: "Google Home", PriceCents: 4999, InventoryQty: 10},
	"43N23P": {SKU: "43N23P", Name: "MacBook Pro", PriceCents: 539999, InventoryQty: 5},
	"A304SD": {SKU: "A304SD", Name: "Alexa Speaker", PriceCents: 10950, InventoryQty: 10},
	"234234": {SKU: "234234", Name: "Raspberry Pi B", PriceCents: 3000, InventoryQty: 2},
}

func TestApplyPromotions_MacBookFreeRaspberry(t *testing.T) {
	u := NewUsecase()
	rules := []entity.PromotionRule{freeProductRule()}
	quantities := map[string]int{"43N23P": 1, "234234": 1}

	promotions, discount, err := u.applyPromotions(quantities, testProducts, rules)

	require.NoError(t, err)
	assert.Equal(t, int64(3000), discount)
	require.Len(t, promotions, 1)
	assert.Equal(t, "MACBOOK_FREE_RASPBERRY", promotions[0].PromotionCode)
	assert.Equal(t, promotionTypeFreeProduct, promotions[0].PromotionType)
}

func TestApplyPromotions_GoogleHomeBuy3Pay2(t *testing.T) {
	u := NewUsecase()
	rules := []entity.PromotionRule{buyXPayYRule()}
	quantities := map[string]int{"120P90": 3}

	promotions, discount, err := u.applyPromotions(quantities, testProducts, rules)

	require.NoError(t, err)
	assert.Equal(t, int64(4999), discount)
	require.Len(t, promotions, 1)
	assert.Equal(t, "GOOGLE_HOME_BUY_3_PAY_2", promotions[0].PromotionCode)
	assert.Equal(t, promotionTypeBuyXPayY, promotions[0].PromotionType)
}

func TestApplyPromotions_AlexaPercentageDiscount(t *testing.T) {
	u := NewUsecase()
	rules := []entity.PromotionRule{percentageRule()}
	quantities := map[string]int{"A304SD": 3}

	promotions, discount, err := u.applyPromotions(quantities, testProducts, rules)

	require.NoError(t, err)
	assert.Equal(t, int64(3285), discount)
	require.Len(t, promotions, 1)
	assert.Equal(t, "ALEXA_10_PERCENT_DISCOUNT", promotions[0].PromotionCode)
	assert.Equal(t, promotionTypePercentage, promotions[0].PromotionType)
}

func TestApplyPromotions_NoPromotion(t *testing.T) {
	u := NewUsecase()
	rules := []entity.PromotionRule{buyXPayYRule(), percentageRule()}
	quantities := map[string]int{"120P90": 1, "A304SD": 1}

	promotions, discount, err := u.applyPromotions(quantities, testProducts, rules)

	require.NoError(t, err)
	assert.Equal(t, int64(0), discount)
	assert.Empty(t, promotions)
}

func TestApplyPromotions_MultiplePromotions(t *testing.T) {
	u := NewUsecase()
	rules := []entity.PromotionRule{freeProductRule(), buyXPayYRule(), percentageRule()}
	quantities := map[string]int{"43N23P": 1, "234234": 1, "120P90": 3, "A304SD": 3}

	promotions, discount, err := u.applyPromotions(quantities, testProducts, rules)

	require.NoError(t, err)
	assert.Equal(t, int64(11284), discount)
	require.Len(t, promotions, 3)
}

func TestMapCheckoutResponse(t *testing.T) {
	checkoutUUID := uuid.New()
	checkout := entity.Checkout{
		UUID:          checkoutUUID,
		Status:        statusCompleted,
		SubtotalCents: 542999,
		DiscountCents: 3000,
		TotalCents:    539999,
	}
	items := []entity.CheckoutItem{
		{SKU: "43N23P", ProductName: "MacBook Pro", Quantity: 1, UnitPriceCents: 539999, TotalPriceCents: 539999},
		{SKU: "234234", ProductName: "Raspberry Pi B", Quantity: 1, UnitPriceCents: 3000, TotalPriceCents: 3000},
	}
	promotions := []appliedPromotion{
		{PromotionCode: "MACBOOK_FREE_RASPBERRY", PromotionName: "MacBook Pro Free Raspberry Pi B", PromotionType: promotionTypeFreeProduct, DiscountCents: 3000},
	}

	resp := mapCheckoutResponse(checkout, items, promotions)

	assert.Equal(t, checkoutUUID.String(), resp.CheckoutID)
	assert.Equal(t, "$5399.99", resp.FormattedTotal)
	assert.Equal(t, int64(542999), resp.SubtotalCents)
	assert.Equal(t, int64(3000), resp.DiscountCents)
	assert.Equal(t, int64(539999), resp.TotalCents)
	require.Len(t, resp.Items, 2)
	require.Len(t, resp.AppliedPromotions, 1)
}

func freeProductRule() entity.PromotionRule {
	return entity.PromotionRule{
		UUID:          uuid.New(),
		PromotionUUID: uuid.New(),
		PromotionCode: "MACBOOK_FREE_RASPBERRY",
		PromotionName: "MacBook Pro Free Raspberry Pi B",
		PromotionType: promotionTypeFreeProduct,
		TargetSKU:     "43N23P",
		RewardSKU:     sql.NullString{String: "234234", Valid: true},
		MinQuantity:   sql.NullInt64{Int64: 1, Valid: true},
		FreeQuantity:  sql.NullInt64{Int64: 1, Valid: true},
	}
}

func buyXPayYRule() entity.PromotionRule {
	return entity.PromotionRule{
		UUID:          uuid.New(),
		PromotionUUID: uuid.New(),
		PromotionCode: "GOOGLE_HOME_BUY_3_PAY_2",
		PromotionName: "Google Home Buy 3 Pay 2",
		PromotionType: promotionTypeBuyXPayY,
		TargetSKU:     "120P90",
		BuyQuantity:   sql.NullInt64{Int64: 3, Valid: true},
		PayQuantity:   sql.NullInt64{Int64: 2, Valid: true},
	}
}

func percentageRule() entity.PromotionRule {
	return entity.PromotionRule{
		UUID:               uuid.New(),
		PromotionUUID:      uuid.New(),
		PromotionCode:      "ALEXA_10_PERCENT_DISCOUNT",
		PromotionName:      "Alexa Speaker 10 Percent Discount",
		PromotionType:      promotionTypePercentage,
		TargetSKU:          "A304SD",
		MinQuantity:        sql.NullInt64{Int64: 3, Valid: true},
		DiscountPercentage: sql.NullFloat64{Float64: 10, Valid: true},
	}
}
