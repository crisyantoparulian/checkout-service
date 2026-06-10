# Checkout Service

Checkout Service is a Go REST API for product checkout, inventory deduction, and promotion calculation.

The service supports:

- Product catalog
- Checkout creation and checkout detail
- Inventory stock deduction
- Promotion engine
- Swagger/OpenAPI documentation
- Docker and Docker Compose runtime
- Unit and integration tests
- GitHub Actions CI

## Tech Stack

- **Language:** Go 1.26
- **HTTP Framework:** Echo
- **Database:** PostgreSQL
- **SQL Library:** sqlx
- **Migration:** Custom migration package
- **Mocking:** gomock / mockgen
- **Integration Test DB:** testcontainers-go
- **API Docs:** swaggo/swag
- **Container:** Docker, Docker Compose
- **CI:** GitHub Actions

## Project Structure

```text
.
├── cmd/                         # CLI commands: server, migration, worker
├── config/                      # Configuration loader
├── docs/                        # Generated Swagger docs
├── internal/
│   ├── app/
│   │   ├── container/           # Dependency injection
│   │   ├── driver/              # Database and external drivers
│   │   ├── entity/              # Database/domain structs
│   │   ├── handler/rest/        # REST handlers
│   │   ├── repository/          # SQL repositories
│   │   ├── server/              # REST server setup
│   │   └── usecase/             # Business logic
│   └── pkg/                     # Shared internal packages
├── migration/                   # Database migrations and seed data
├── tests/integration/           # Integration tests using testcontainers
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── main.go
```

## Business Rules

### Product Identity

- Product UUID is the canonical reference across the application.
- SKU is business/display information only.
- SKU is not used as the primary business reference in checkout or promotion logic.
- Some tables keep SKU as denormalized data for faster reads and historical snapshots.

### Checkout

- Checkout runs inside a database transaction.
- Products are locked with `SELECT ... FOR UPDATE` during checkout.
- Inventory is deducted only after checkout data is successfully created.
- Inventory movement records are created for each deducted product.

### Promotions

Supported promotion types:

| Type | Description | Example |
|---|---|---|
| `FREE_PRODUCT` | Buy target product and get reward product free | Buy MacBook Pro, get Raspberry Pi B free |
| `BUY_X_PAY_Y` | Buy X items, pay only Y items | Buy 3 Google Home, pay for 2 |
| `PERCENTAGE_DISCOUNT` | Discount by percentage after minimum quantity | Buy 3+ Alexa Speakers, get 10% off |

For `FREE_PRODUCT`, the reward product is automatically added to checkout items when eligible.

Example request with only MacBook Pro:

```json
{
  "items": [
    {
      "product_uuid": "MACBOOK_PRODUCT_UUID",
      "quantity": 1
    }
  ]
}
```

The response will include MacBook Pro and the free Raspberry Pi B item. The subtotal includes both items, and the discount offsets the free item price.

## Database Design

### Tables

#### `products`

Stores product catalog and inventory quantity.

| Column | Description |
|---|---|
| `uuid` | Product primary key |
| `sku` | Unique business SKU |
| `name` | Product name |
| `price_cents` | Product price in cents |
| `inventory_qty` | Current inventory quantity |
| `created_at`, `updated_at` | Audit timestamps |

#### `promotions`

Stores promotion header data.

| Column | Description |
|---|---|
| `uuid` | Promotion primary key |
| `code` | Unique promotion code |
| `name` | Promotion display name |
| `description` | Promotion description |
| `is_active` | Active flag |
| `start_date`, `end_date` | Optional validity window |
| `created_at`, `updated_at` | Audit timestamps |

#### `promotion_rules`

Stores promotion rule configuration.

| Column | Description |
|---|---|
| `uuid` | Rule primary key |
| `promotion_uuid` | FK to `promotions.uuid` |
| `promotion_type` | Promotion type constant |
| `target_product_uuid` | Target product UUID reference |
| `target_sku` | Denormalized target SKU |
| `reward_product_uuid` | Reward product UUID reference for free product rules |
| `reward_sku` | Denormalized reward SKU |
| `min_quantity` | Minimum target quantity |
| `free_quantity` | Free reward quantity |
| `buy_quantity` | Buy quantity for buy-X-pay-Y |
| `pay_quantity` | Pay quantity for buy-X-pay-Y |
| `discount_percentage` | Percentage discount value |
| `priority` | Rule ordering priority |
| `created_at`, `updated_at` | Audit timestamps |

#### `checkouts`

Stores checkout summary.

| Column | Description |
|---|---|
| `uuid` | Checkout primary key |
| `status` | Checkout status |
| `subtotal_cents` | Total before discount |
| `discount_cents` | Total discount |
| `total_cents` | Final total |
| `created_at`, `updated_at` | Audit timestamps |

#### `checkout_items`

Stores checkout item snapshot.

| Column | Description |
|---|---|
| `uuid` | Checkout item primary key |
| `checkout_uuid` | FK to `checkouts.uuid` |
| `product_uuid` | Product UUID reference |
| `sku` | Denormalized SKU snapshot |
| `product_name` | Product name snapshot |
| `quantity` | Item quantity |
| `unit_price_cents` | Unit price snapshot |
| `total_price_cents` | Item total |
| `created_at` | Created timestamp |

#### `checkout_promotions`

Stores applied promotion snapshot.

| Column | Description |
|---|---|
| `uuid` | Applied promotion primary key |
| `checkout_uuid` | FK to `checkouts.uuid` |
| `promotion_uuid` | FK to `promotions.uuid` |
| `promotion_rule_uuid` | FK to `promotion_rules.uuid` |
| `promotion_code` | Promotion code snapshot |
| `promotion_name` | Promotion name snapshot |
| `promotion_type` | Promotion type snapshot |
| `discount_cents` | Discount amount |
| `rule_snapshot` | JSON rule snapshot |
| `created_at` | Created timestamp |

#### `inventory_movements`

Stores stock movement audit history.

| Column | Description |
|---|---|
| `uuid` | Movement primary key |
| `product_uuid` | Product UUID reference |
| `sku` | Denormalized SKU snapshot |
| `checkout_uuid` | Related checkout UUID |
| `movement_type` | Movement type, e.g. `CHECKOUT_DEDUCT` |
| `quantity` | Stock movement quantity, negative for deduction |
| `stock_before` | Stock before movement |
| `stock_after` | Stock after movement |
| `note` | Optional note |
| `created_at` | Created timestamp |

### Relationship Summary

```text
products.uuid
  ├── promotion_rules.target_product_uuid
  ├── promotion_rules.reward_product_uuid
  ├── checkout_items.product_uuid
  └── inventory_movements.product_uuid

promotions.uuid
  ├── promotion_rules.promotion_uuid
  └── checkout_promotions.promotion_uuid

checkouts.uuid
  ├── checkout_items.checkout_uuid
  ├── checkout_promotions.checkout_uuid
  └── inventory_movements.checkout_uuid
```

## Environment Setup

Create `.env` from example:

```bash
cp .env.example .env
```

Important variables:

```env
APP_HTTP_PORT=9000
DB_DRIVER=postgres
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=checkout_db
DB_HOST=127.0.0.1
DB_PORT=5432
DB_SSL_MODE=disable
```

The app reads `.env`, and environment variables can override `.env` values.

## Running Locally

Install dependencies:

```bash
go mod download
```

Run migration:

```bash
make migrate-up
```

Run REST server:

```bash
make rest
```

REST API runs on:

```text
http://localhost:9000
```

Swagger UI:

```text
http://localhost:9000/swagger/index.html
```

## Running with Docker Compose

Docker Compose starts PostgreSQL 17 and the app.

```bash
docker compose up --build
```

The app container runs migration before starting the REST server:

```bash
/app/checkout-service migrate up && /app/checkout-service server rest
```

Stop containers:

```bash
docker compose down
```

Remove database volume:

```bash
docker compose down -v
```

## Docker Only

Build image:

```bash
docker build -t checkout-service .
```

Run image:

```bash
docker run --rm -p 9000:9000 checkout-service
```

If connecting to a local host database from Docker:

```bash
docker run --rm -p 9000:9000 \
  -e DB_HOST=host.docker.internal \
  checkout-service
```

## Make Commands

| Command | Description |
|---|---|
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests with testcontainers |
| `make mockgen-install` | Install mockgen |
| `make mock` | Generate mocks |
| `make swag-init` | Generate Swagger docs |
| `make rest` | Run REST server |
| `make migrate-up` | Run database migration up |

## Testing

Run unit tests:

```bash
make test
```

Run REST handler tests:

```bash
go test -v ./internal/app/handler/rest/...
```

Run checkout usecase tests:

```bash
go test -v ./internal/app/usecase/checkout/...
```

Run integration tests:

```bash
make test-integration
```

Integration tests use `testcontainers-go` and require Docker to be running.

## Mock Generation

This project uses `go.uber.org/mock/gomock` and `mockgen`.

Install mockgen:

```bash
make mockgen-install
```

Generate mocks:

```bash
make mock
```

Generated mocks are located in:

- `internal/app/repository/mocks/`
- `internal/app/usecase/*/mocks/`
- `internal/pkg/transaction/mocks/`

## API Documentation

Swagger annotations are in `main.go` and REST handler comments.

Generate Swagger docs:

```bash
make swag-init
```

Generated Swagger files:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

## CI

GitHub Actions workflow is available at:

```text
.github/workflows/ci.yml
```

CI runs on `master` branch push and pull request:

- Formatting check
- Unit tests
- Docker build
- Integration tests

## Claude Code Context

Project-specific Claude Code instructions are documented in:

```text
CLAUDE.md
```

## Author

Crisyanto Parulian

## License

This project is licensed under the MIT License.
