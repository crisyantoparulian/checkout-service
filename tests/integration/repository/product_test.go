//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepo_Integration(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewProductRepository(tdb.DB)

	// Get seeded product UUIDs for tests
	var macBookUUID, raspberryUUID, googleHomeUUID, alexaUUID uuid.UUID
	err := tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '43N23P'`).StructScan(&struct {
		UUID uuid.UUID `db:"uuid"`
	}{})
	require.NoError(t, err)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '43N23P'`).Scan(&macBookUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '234234'`).Scan(&raspberryUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '120P90'`).Scan(&googleHomeUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = 'A304SD'`).Scan(&alexaUUID)

	t.Run("GetAll returns seeded products", func(t *testing.T) {
		products, err := repo.GetAll(context.Background(), repository.ProductFilter{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, products, 4)
	})

	t.Run("GetAll with pagination", func(t *testing.T) {
		products, err := repo.GetAll(context.Background(), repository.ProductFilter{Limit: 2, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, products, 2)
	})

	t.Run("CountTotal returns 4", func(t *testing.T) {
		total, err := repo.CountTotal(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(4), total)
	})

	t.Run("GetByUUIDs returns correct products", func(t *testing.T) {
		products, err := repo.GetByUUIDs(context.Background(), []uuid.UUID{macBookUUID, googleHomeUUID})
		require.NoError(t, err)
		assert.Len(t, products, 2)

		skus := map[string]bool{}
		for _, p := range products {
			skus[p.SKU] = true
		}
		assert.True(t, skus["43N23P"])
		assert.True(t, skus["120P90"])
	})

	t.Run("GetByUUIDs with non-existent UUID returns empty", func(t *testing.T) {
		products, err := repo.GetByUUIDs(context.Background(), []uuid.UUID{uuid.New()})
		require.NoError(t, err)
		assert.Len(t, products, 0)
	})

	t.Run("GetByUUIDsForUpdate locks within transaction", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)
		defer tx.Rollback()

		products, err := repo.GetByUUIDsForUpdate(context.Background(), tx, []uuid.UUID{macBookUUID})
		require.NoError(t, err)
		assert.Len(t, products, 1)
		assert.Equal(t, "43N23P", products[0].SKU)
		assert.Equal(t, int64(539999), products[0].PriceCents)
		assert.Equal(t, 5, products[0].InventoryQty)
	})

	t.Run("DeductStock reduces inventory", func(t *testing.T) {
		// Use a transaction and rollback to not affect other tests
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)
		defer tx.Rollback()

		err = repo.DeductStock(context.Background(), tx, macBookUUID, 2)
		require.NoError(t, err)

		var qty int
		err = tx.Get(&qty, `SELECT inventory_qty FROM products WHERE uuid = $1`, macBookUUID)
		require.NoError(t, err)
		assert.Equal(t, 3, qty)
	})

	t.Run("GetAll returns correct product fields", func(t *testing.T) {
		products, err := repo.GetAll(context.Background(), repository.ProductFilter{Limit: 10, Offset: 0})
		require.NoError(t, err)

		productMap := map[string]struct {
			Name         string
			PriceCents   int64
			InventoryQty int
		}{}
		for _, p := range products {
			productMap[p.SKU] = struct {
				Name         string
				PriceCents   int64
				InventoryQty int
			}{Name: p.Name, PriceCents: p.PriceCents, InventoryQty: p.InventoryQty}
		}

		assert.Equal(t, "MacBook Pro", productMap["43N23P"].Name)
		assert.Equal(t, int64(539999), productMap["43N23P"].PriceCents)
		assert.Equal(t, 5, productMap["43N23P"].InventoryQty)

		assert.Equal(t, "Raspberry Pi B", productMap["234234"].Name)
		assert.Equal(t, int64(3000), productMap["234234"].PriceCents)

		assert.Equal(t, "Google Home", productMap["120P90"].Name)
		assert.Equal(t, int64(4999), productMap["120P90"].PriceCents)

		assert.Equal(t, "Alexa Speaker", productMap["A304SD"].Name)
		assert.Equal(t, int64(10950), productMap["A304SD"].PriceCents)
	})
}
