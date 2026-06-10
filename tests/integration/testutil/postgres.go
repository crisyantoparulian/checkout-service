//go:build integration

package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/crisyantoparulian/checkout-service/migration"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB wraps a PostgreSQL testcontainer and provides a clean *sqlx.DB.
type TestDB struct {
	DB        *sqlx.DB
	container *postgrescontainer.PostgresContainer
}

// NewTestDB spins up a Postgres container, runs migrations, and returns a ready-to-use TestDB.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	ctx := context.Background()

	container, err := postgrescontainer.Run(ctx,
		"postgres:16-alpine",
		postgrescontainer.WithDatabase("checkout_test"),
		postgrescontainer.WithUsername("postgres"),
		postgrescontainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	// Run migrations
	migrator, err := migration.Init(db)
	require.NoError(t, err)

	// Register migrations (auto-registered via init())
	_ = migrator // migrator already has migrations registered

	err = migrator.Up()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	})

	return &TestDB{
		DB:        db,
		container: container,
	}
}

// TruncateAll truncates all data tables but keeps the schema (for test isolation).
func (tdb *TestDB) TruncateAll(t *testing.T) {
	t.Helper()
	_, err := tdb.DB.Exec(`
		TRUNCATE TABLE inventory_movements, checkout_promotions, checkout_items, checkouts, promotion_rules, promotions, products RESTART IDENTITY CASCADE;
	`)
	require.NoError(t, err)
}

// Reseed inserts the seed data from the migration (products, promotions, rules).
func (tdb *TestDB) Reseed(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// Seed products
	_, err := tdb.DB.ExecContext(ctx, `
		INSERT INTO products (sku, name, price_cents, inventory_qty)
		VALUES
			('120P90', 'Google Home', 4999, 10),
			('43N23P', 'MacBook Pro', 539999, 5),
			('A304SD', 'Alexa Speaker', 10950, 10),
			('234234', 'Raspberry Pi B', 3000, 2)
		ON CONFLICT (sku) DO NOTHING
	`)
	require.NoError(t, err)

	// Seed promotions
	_, err = tdb.DB.ExecContext(ctx, `
		INSERT INTO promotions (code, name, description, is_active)
		VALUES
			('MACBOOK_FREE_RASPBERRY', 'MacBook Pro Free Raspberry Pi B', 'Each MacBook Pro sale comes with a free Raspberry Pi B', TRUE),
			('GOOGLE_HOME_BUY_3_PAY_2', 'Google Home Buy 3 Pay 2', 'Buy 3 Google Homes for the price of 2', TRUE),
			('ALEXA_10_PERCENT_DISCOUNT', 'Alexa Speaker 10 Percent Discount', 'Buying at least 3 Alexa Speakers gets 10 percent discount', TRUE)
		ON CONFLICT (code) DO NOTHING
	`)
	require.NoError(t, err)

	// Seed promotion rules
	rules := []struct {
		promoCode   string
		promoType   string
		targetSKU   string
		rewardSKU   string
		minQty      *int
		freeQty     *int
		buyQty      *int
		payQty      *int
		discountPct *float64
		priority    int
	}{
		{
			promoCode: "MACBOOK_FREE_RASPBERRY", promoType: "FREE_PRODUCT",
			targetSKU: "43N23P", rewardSKU: "234234",
			minQty: intPtr(1), freeQty: intPtr(1),
			priority: 10,
		},
		{
			promoCode: "GOOGLE_HOME_BUY_3_PAY_2", promoType: "BUY_X_PAY_Y",
			targetSKU: "120P90",
			buyQty:    intPtr(3), payQty: intPtr(2),
			priority: 20,
		},
		{
			promoCode: "ALEXA_10_PERCENT_DISCOUNT", promoType: "PERCENTAGE_DISCOUNT",
			targetSKU: "A304SD",
			minQty:    intPtr(3), discountPct: floatPtr(10.0),
			priority: 30,
		},
	}

	for _, r := range rules {
		query := fmt.Sprintf(`
			INSERT INTO promotion_rules (promotion_uuid, promotion_type, target_product_uuid, target_sku, reward_product_uuid, reward_sku, min_quantity, free_quantity, buy_quantity, pay_quantity, discount_percentage, priority)
			SELECT p.uuid, '%s', (SELECT pr.uuid FROM products pr WHERE pr.sku = '%s'), '%s',
		`, r.promoType, r.targetSKU, r.targetSKU)

		if r.rewardSKU != "" {
			query += fmt.Sprintf(`(SELECT pr.uuid FROM products pr WHERE pr.sku = '%s'), '%s',`, r.rewardSKU, r.rewardSKU)
		} else {
			query += `NULL, NULL,`
		}

		query += fmt.Sprintf(`%s, %s, %s, %s, %s, %d`,
			nullIntPtr(r.minQty), nullIntPtr(r.freeQty), nullIntPtr(r.buyQty), nullIntPtr(r.payQty), nullFloatPtr(r.discountPct), r.priority)

		query += fmt.Sprintf(` FROM promotions p WHERE p.code = '%s' ON CONFLICT DO NOTHING;`, r.promoCode)

		_, err := tdb.DB.ExecContext(ctx, query)
		require.NoError(t, err)
	}
}

func intPtr(v int) *int           { return &v }
func floatPtr(v float64) *float64 { return &v }

func nullIntPtr(v *int) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%d", *v)
}

func nullFloatPtr(v *float64) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%g", *v)
}
