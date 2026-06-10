//go:build integration

package usecase

import (
	"context"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/transaction"
	"github.com/crisyantoparulian/checkout-service/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsecase(t *testing.T) (checkout.CheckoutUsecase, *testutil.TestDB) {
	t.Helper()
	tdb := testutil.NewTestDB(t)

	txManager := transaction.NewManager(tdb.DB)
	checkoutRepo := repository.NewCheckoutRepository(tdb.DB)
	productRepo := repository.NewProductRepository(tdb.DB)
	promotionRepo := repository.NewPromotionRepository(tdb.DB)
	inventoryRepo := repository.NewInventoryRepository(tdb.DB)

	uc := checkout.NewUsecase().
		SetTxManager(txManager).
		SetCheckoutRepository(checkoutRepo).
		SetProductRepository(productRepo).
		SetPromotionRepository(promotionRepo).
		SetInventoryRepository(inventoryRepo).
		Validate()

	return uc, tdb
}

func getProductUUIDs(t *testing.T, tdb *testutil.TestDB) (macBook, raspberry, googleHome, alexa uuid.UUID) {
	t.Helper()
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '43N23P'`).Scan(&macBook)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '234234'`).Scan(&raspberry)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '120P90'`).Scan(&googleHome)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = 'A304SD'`).Scan(&alexa)
	require.NotEqual(t, uuid.Nil, macBook)
	return
}

func TestCheckoutUsecase_Create_Integration(t *testing.T) {
	uc, tdb := setupUsecase(t)
	macBookUUID, _, _, _ := getProductUUIDs(t, tdb)

	t.Run("MacBook auto adds free Raspberry promotion", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: macBookUUID.String(), Quantity: 1},
			},
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.CheckoutID)
		assert.Equal(t, "COMPLETED", resp.Status)
		assert.Equal(t, int64(542999), resp.SubtotalCents)
		assert.Equal(t, int64(3000), resp.DiscountCents)
		assert.Equal(t, int64(539999), resp.TotalCents)
		assert.Equal(t, "$5399.99", resp.FormattedTotal)
		assert.Len(t, resp.Items, 2)
		assert.Len(t, resp.AppliedPromotions, 1)
		assert.Equal(t, "FREE_PRODUCT", resp.AppliedPromotions[0].PromotionType)
	})

	// Re-seed since previous test deducted stock
	tdb.TruncateAll(t)
	tdb.Reseed(t)
	macBookUUID, raspberryUUID, _, _ := getProductUUIDs(t, tdb)

	t.Run("MacBook + Raspberry applies free product promotion", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: macBookUUID.String(), Quantity: 1},
				{ProductUUID: raspberryUUID.String(), Quantity: 1},
			},
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.CheckoutID)
		assert.Equal(t, int64(542999), resp.SubtotalCents)
		assert.Equal(t, int64(3000), resp.DiscountCents)
		assert.Equal(t, int64(539999), resp.TotalCents)
		assert.Len(t, resp.Items, 2)
		assert.Len(t, resp.AppliedPromotions, 1)
		assert.Equal(t, "FREE_PRODUCT", resp.AppliedPromotions[0].PromotionType)
	})

	tdb.TruncateAll(t)
	tdb.Reseed(t)
	_, _, googleHomeUUID, _ := getProductUUIDs(t, tdb)

	t.Run("3x Google Home applies buy 3 pay 2", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: googleHomeUUID.String(), Quantity: 3},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(14997), resp.SubtotalCents)  // 4999 * 3
		assert.Equal(t, int64(4999), resp.DiscountCents)    // 1 free Google Home
		assert.Equal(t, int64(9998), resp.TotalCents)       // 14997 - 4999
		assert.Len(t, resp.AppliedPromotions, 1)
		assert.Equal(t, "BUY_X_PAY_Y", resp.AppliedPromotions[0].PromotionType)
	})

	tdb.TruncateAll(t)
	tdb.Reseed(t)
	_, _, _, alexaUUID := getProductUUIDs(t, tdb)

	t.Run("3x Alexa applies 10% discount", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: alexaUUID.String(), Quantity: 3},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(32850), resp.SubtotalCents) // 10950 * 3
		assert.Equal(t, int64(3285), resp.DiscountCents)  // 10% of 32850 = 3285
		assert.Equal(t, int64(29565), resp.TotalCents)    // 32850 - 3285
		assert.Len(t, resp.AppliedPromotions, 1)
		assert.Equal(t, "PERCENTAGE_DISCOUNT", resp.AppliedPromotions[0].PromotionType)
	})

	tdb.TruncateAll(t)
	tdb.Reseed(t)
	macBookUUID, raspberryUUID, googleHomeUUID, alexaUUID = getProductUUIDs(t, tdb)

	t.Run("multiple promotions applied together", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: macBookUUID.String(), Quantity: 1},
				{ProductUUID: raspberryUUID.String(), Quantity: 1},
				{ProductUUID: googleHomeUUID.String(), Quantity: 3},
				{ProductUUID: alexaUUID.String(), Quantity: 3},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(11284), resp.DiscountCents) // 3000 + 4999 + 3285
		assert.Len(t, resp.AppliedPromotions, 3)
	})

	tdb.TruncateAll(t)
	tdb.Reseed(t)
	macBookUUID, _, _, _ = getProductUUIDs(t, tdb)

	t.Run("insufficient stock returns error", func(t *testing.T) {
		_, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: macBookUUID.String(), Quantity: 100},
			},
		})
		require.Error(t, err)
	})

	t.Run("invalid product UUID returns error", func(t *testing.T) {
		_, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: "not-a-uuid", Quantity: 1},
			},
		})
		require.Error(t, err)
	})

	t.Run("zero quantity returns error", func(t *testing.T) {
		_, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
			Items: []checkout.CheckoutItemRequest{
				{ProductUUID: macBookUUID.String(), Quantity: 0},
			},
		})
		require.Error(t, err)
	})
}

func TestCheckoutUsecase_GetDetail_Integration(t *testing.T) {
	uc, tdb := setupUsecase(t)
	macBookUUID, _, _, _ := getProductUUIDs(t, tdb)

	// Create a checkout first
	resp, err := uc.Create(context.Background(), checkout.CreateCheckoutRequest{
		Items: []checkout.CheckoutItemRequest{
			{ProductUUID: macBookUUID.String(), Quantity: 1},
		},
	})
	require.NoError(t, err)

	t.Run("get detail returns created checkout", func(t *testing.T) {
		detail, err := uc.GetDetail(context.Background(), resp.CheckoutID)
		require.NoError(t, err)
		assert.Equal(t, resp.CheckoutID, detail.CheckoutID)
		assert.Equal(t, "COMPLETED", detail.Status)
		assert.Equal(t, int64(539999), detail.TotalCents)
		assert.Len(t, detail.Items, 2)
		assert.Equal(t, macBookUUID.String(), detail.Items[0].ProductUUID)
		assert.Equal(t, "43N23P", detail.Items[0].SKU)
	})

	t.Run("get detail with invalid UUID returns error", func(t *testing.T) {
		_, err := uc.GetDetail(context.Background(), "not-a-uuid")
		require.Error(t, err)
	})

	t.Run("get detail with non-existent UUID returns error", func(t *testing.T) {
		_, err := uc.GetDetail(context.Background(), uuid.New().String())
		require.Error(t, err)
	})
}
