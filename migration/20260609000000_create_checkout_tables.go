package migration

import (
	"time"

	"github.com/jmoiron/sqlx"
)

func init() {
	migrator.AddMigration(&Migration{
		ID:        0,
		Name:      "20260609000000_create_checkout_tables",
		Batch:     0,
		CreatedAt: time.Time{},
		Up:        mig_20260609000000_create_checkout_tables_up,
		Down:      mig_20260609000000_create_checkout_tables_down,
	})
}

func mig_20260609000000_create_checkout_tables_up(tx *sqlx.Tx) error {
	_, err := tx.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS products (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			sku VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
			inventory_qty INT NOT NULL CHECK (inventory_qty >= 0),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS promotions (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			code VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			start_at TIMESTAMP NULL,
			end_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS promotion_rules (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			promotion_uuid UUID NOT NULL REFERENCES promotions(uuid) ON DELETE CASCADE,
			promotion_type VARCHAR(50) NOT NULL,
			target_sku VARCHAR(50) NOT NULL REFERENCES products(sku),
			reward_sku VARCHAR(50) NULL REFERENCES products(sku),
			min_quantity INT NULL CHECK (min_quantity IS NULL OR min_quantity > 0),
			buy_quantity INT NULL CHECK (buy_quantity IS NULL OR buy_quantity > 0),
			pay_quantity INT NULL CHECK (pay_quantity IS NULL OR pay_quantity > 0),
			free_quantity INT NULL CHECK (free_quantity IS NULL OR free_quantity > 0),
			discount_percentage NUMERIC(5,2) NULL CHECK (discount_percentage IS NULL OR discount_percentage BETWEEN 0 AND 100),
			priority INT NOT NULL DEFAULT 100,
			is_stackable BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (promotion_type IN ('FREE_PRODUCT', 'BUY_X_PAY_Y', 'PERCENTAGE_DISCOUNT'))
		);

		CREATE TABLE IF NOT EXISTS checkouts (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			status VARCHAR(50) NOT NULL DEFAULT 'COMPLETED',
			subtotal_cents BIGINT NOT NULL CHECK (subtotal_cents >= 0),
			discount_cents BIGINT NOT NULL CHECK (discount_cents >= 0),
			total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (status IN ('COMPLETED', 'CANCELLED'))
		);

		CREATE TABLE IF NOT EXISTS checkout_items (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			checkout_uuid UUID NOT NULL REFERENCES checkouts(uuid) ON DELETE CASCADE,
			sku VARCHAR(50) NOT NULL REFERENCES products(sku),
			product_name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL CHECK (quantity > 0),
			unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
			total_price_cents BIGINT NOT NULL CHECK (total_price_cents >= 0),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS checkout_promotions (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			checkout_uuid UUID NOT NULL REFERENCES checkouts(uuid) ON DELETE CASCADE,
			promotion_uuid UUID NULL REFERENCES promotions(uuid),
			promotion_rule_uuid UUID NULL REFERENCES promotion_rules(uuid),
			promotion_code VARCHAR(100) NOT NULL,
			promotion_name VARCHAR(255) NOT NULL,
			promotion_type VARCHAR(50) NOT NULL,
			discount_cents BIGINT NOT NULL CHECK (discount_cents >= 0),
			rule_snapshot JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS inventory_movements (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			sku VARCHAR(50) NOT NULL REFERENCES products(sku),
			checkout_uuid UUID NULL REFERENCES checkouts(uuid),
			movement_type VARCHAR(50) NOT NULL,
			quantity INT NOT NULL,
			stock_before INT NOT NULL,
			stock_after INT NOT NULL,
			note TEXT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (movement_type IN ('INITIAL_STOCK', 'CHECKOUT_DEDUCT', 'RESTOCK', 'ADJUSTMENT', 'CANCEL_RESTORE'))
		);

		INSERT INTO products (sku, name, price_cents, inventory_qty)
		VALUES
			('120P90', 'Google Home', 4999, 10),
			('43N23P', 'MacBook Pro', 539999, 5),
			('A304SD', 'Alexa Speaker', 10950, 10),
			('234234', 'Raspberry Pi B', 3000, 2)
		ON CONFLICT (sku) DO NOTHING;

		INSERT INTO promotions (code, name, description, is_active)
		VALUES
			('MACBOOK_FREE_RASPBERRY', 'MacBook Pro Free Raspberry Pi B', 'Each MacBook Pro sale comes with a free Raspberry Pi B', TRUE),
			('GOOGLE_HOME_BUY_3_PAY_2', 'Google Home Buy 3 Pay 2', 'Buy 3 Google Homes for the price of 2', TRUE),
			('ALEXA_10_PERCENT_DISCOUNT', 'Alexa Speaker 10 Percent Discount', 'Buying at least 3 Alexa Speakers gets 10 percent discount', TRUE)
		ON CONFLICT (code) DO NOTHING;

		INSERT INTO promotion_rules (promotion_uuid, promotion_type, target_sku, reward_sku, min_quantity, free_quantity, priority)
		SELECT uuid, 'FREE_PRODUCT', '43N23P', '234234', 1, 1, 10
		FROM promotions WHERE code = 'MACBOOK_FREE_RASPBERRY'
		ON CONFLICT DO NOTHING;

		INSERT INTO promotion_rules (promotion_uuid, promotion_type, target_sku, buy_quantity, pay_quantity, priority)
		SELECT uuid, 'BUY_X_PAY_Y', '120P90', 3, 2, 20
		FROM promotions WHERE code = 'GOOGLE_HOME_BUY_3_PAY_2'
		ON CONFLICT DO NOTHING;

		INSERT INTO promotion_rules (promotion_uuid, promotion_type, target_sku, min_quantity, discount_percentage, priority)
		SELECT uuid, 'PERCENTAGE_DISCOUNT', 'A304SD', 3, 10.00, 30
		FROM promotions WHERE code = 'ALEXA_10_PERCENT_DISCOUNT'
		ON CONFLICT DO NOTHING;

		INSERT INTO inventory_movements (sku, movement_type, quantity, stock_before, stock_after, note)
		VALUES
			('120P90', 'INITIAL_STOCK', 10, 0, 10, 'Initial seed stock'),
			('43N23P', 'INITIAL_STOCK', 5, 0, 5, 'Initial seed stock'),
			('A304SD', 'INITIAL_STOCK', 10, 0, 10, 'Initial seed stock'),
			('234234', 'INITIAL_STOCK', 2, 0, 2, 'Initial seed stock');
	`)
	return err
}

func mig_20260609000000_create_checkout_tables_down(tx *sqlx.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS inventory_movements;
		DROP TABLE IF EXISTS checkout_promotions;
		DROP TABLE IF EXISTS checkout_items;
		DROP TABLE IF EXISTS checkouts;
		DROP TABLE IF EXISTS promotion_rules;
		DROP TABLE IF EXISTS promotions;
		DROP TABLE IF EXISTS products;
	`)
	return err
}
