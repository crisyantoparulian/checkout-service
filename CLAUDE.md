# CLAUDE.md

This file provides project-specific context and instructions for Claude Code when working in this repository.

## Project Overview

This is a Go checkout service using Echo, PostgreSQL, sqlx, and a promotion engine. The service supports product listing, checkout creation/detail, inventory deduction, and promotions.

## Important Business Rules

- Product UUID is the canonical reference key across the codebase.
- SKU is business/display information only and should not be used as a foreign key or lookup reference in business logic.
- Denormalized SKU columns are intentionally kept in checkout, promotion, and inventory tables for read/display performance.
- Checkout must run in a database transaction.
- Product locking during checkout should use `GetByUUIDsForUpdate`.
- Free product promotions should automatically add eligible reward products into checkout items.
- Inventory must be deducted for both requested products and automatically added reward products.
- Promotion types are public constants in `internal/pkg/constants/promotion.go`.

## Promotion Types

- `FREE_PRODUCT`
  - Example: buy MacBook Pro, get Raspberry Pi B free.
  - Reward product can be auto-added when the customer only submits the target product.
- `BUY_X_PAY_Y`
  - Example: buy 3 Google Home, pay for 2.
- `PERCENTAGE_DISCOUNT`
  - Example: buy 3+ Alexa Speakers and get 10% discount.

## Testing Standards

- Use table-driven tests.
- Use `go.uber.org/mock/gomock` and `mockgen` for mocks.
- Do not use mockery.
- Do not use stretchr testify mock.
- Repository DB-level tests may use `go-sqlmock` where appropriate.
- Integration tests use `testcontainers-go` with PostgreSQL.
- Integration tests are behind the `integration` build tag.

## Common Commands

```bash
make test
make test-integration
make mock
make swag-init
make rest
make migrate-up
```

Run checkout usecase tests:

```bash
go test -v ./internal/app/usecase/checkout/...
```

Run REST handler tests:

```bash
go test -v ./internal/app/handler/rest/...
```

Run integration tests:

```bash
go test -v -tags=integration ./tests/integration/...
```

## Mock Generation

Use mockgen:

```bash
make mockgen-install
make mock
```

Generated mocks live in:

- `internal/app/repository/mocks/`
- `internal/app/usecase/*/mocks/`
- `internal/pkg/transaction/mocks/`

## Docker

Docker runs migrations before starting the REST server:

```bash
docker compose up --build
```

The Dockerfile command runs:

```bash
/app/checkout-service migrate up && /app/checkout-service server rest
```

## Database Notes

- PostgreSQL is used.
- Migrations are in `migration/`.
- Seed data is created by `20260609000000_create_checkout_tables.go`.
- Products, checkouts, promotions, and inventory movements use UUID identifiers.
- `promotion_rules` stores both target/reward product UUIDs and denormalized SKUs.

## Swagger

Swagger docs are generated from annotations in `main.go` and REST handler comments.

```bash
make swag-init
```

Generated files:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

## CI

GitHub Actions workflow is in `.github/workflows/ci.yml` and runs unit tests, Docker build, and integration tests on `master`.