# Checkout Service

A checkout service built with Go, implementing promotion engine with support for free products, buy-X-pay-Y, and percentage discounts.

## Tech Stack

- **Language:** Go 1.26
- **Framework:** [Echo](https://echo.labstack.com/)
- **Database:** PostgreSQL with [sqlx](https://jmoiron.github.io/sqlx/)
- **Worker:** [Asynq](https://github.com/hibiken/asynq)
- **Logging:** [Gobang Logger](https://github.com/insaneadinesia/gobang/tree/master/logger)
- **Tracing:** OpenTelemetry via [Gobang Gotel](https://github.com/insaneadinesia/gobang/tree/master/gotel)
- **Documentation:** Swagger/OpenAPI

## Project Structure

```
.
├── cmd/                   # Application entry points
├── config/                # Configuration management
├── docs/                  # Swagger documentation
├── internal/
│   ├── app/
│   │   ├── container/     # Dependency injection
│   │   ├── driver/        # Database & external service drivers
│   │   ├── entity/        # Domain entities
│   │   ├── handler/       # REST handlers
│   │   ├── repository/    # Data access layer (sqlx, prepared statements)
│   │   ├── server/        # Server implementations
│   │   └── usecase/       # Business logic
│   └── pkg/               # Internal shared packages
├── migration/             # Database migrations
└── main.go                # Application entry point
```

## Features

- **Product catalog** with UUID primary key and SKU as unique business identifier
- **Checkout** with transactional consistency (auto rollback on error)
- **Promotion engine** supporting:
  - Free product (e.g., buy MacBook, get free Raspberry Pi)
  - Buy X Pay Y (e.g., buy 3 Google Home, pay for 2)
  - Percentage discount (e.g., 10% off 3+ Alexa Speakers)
- **Inventory management** with stock deduction and movement tracking
- **Prepared statements** for all write operations (SQL injection protection)

## Getting Started

1. **Clone the repository**
   ```bash
   git clone https://github.com/crisyantoparulian/checkout-service.git
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run the application**
   ```bash
   # REST API Server
   go run main.go server rest

   # Database Create Migration
   go run main.go migrate create --name=xxx

   # Database Run Migration
   go run main.go migrate up

   # Database Rollback Migration
   go run main.go migrate down
   ```

## Testing

```bash
make test
```

## API Documentation

Generate Swagger documentation:
```bash
make swag-init
```

Access Swagger UI at: `http://localhost:9000/swagger/index.html`

## Author

[crisyantoparulian](https://github.com/crisyantoparulian)

## License

This project is licensed under the MIT License.
