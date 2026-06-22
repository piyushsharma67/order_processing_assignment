# Order Processing System

A Go backend for an e-commerce order processing system. Customers can place orders with multiple items, track status, list orders, and cancel pending orders. A background worker automatically moves `PENDING` orders to `PROCESSING` every 5 minutes.

## Features

- **Create order** — place an order with multiple line items
- **Get order** — fetch order details by ID
- **List orders** — retrieve all orders, optionally filtered by status
- **Update status** — transition orders through `PENDING` → `PROCESSING` → `SHIPPED` → `DELIVERED`
- **Cancel order** — allowed only while status is `PENDING`
- **Background job** — auto-updates all `PENDING` orders to `PROCESSING` every 5 minutes

## Project Structure

```
cmd/server/              # Application entry point
internal/domain/         # Order models, status rules, validation
internal/repository/     # In-memory persistence layer
internal/service/        # Business logic
internal/handler/        # HTTP handlers and routing
internal/worker/         # Background status updater
```

## Requirements

- Go 1.22+

## Quick Start

```bash
# Download dependencies
go mod tidy

# Run tests
go test ./...

# Start the server (default port 8080)
go run ./cmd/server
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `STATUS_UPDATE_INTERVAL` | `5m` | Background job interval (e.g. `30s` for testing) |

## API Reference

### Health Check

```bash
curl http://localhost:8080/health
```

### Create Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-123",
    "items": [
      {
        "product_id": "prod-1",
        "product_name": "Wireless Mouse",
        "quantity": 2,
        "unit_price": 29.99
      },
      {
        "product_id": "prod-2",
        "product_name": "USB-C Cable",
        "quantity": 1,
        "unit_price": 12.50
      }
    ]
  }'
```

### Get Order by ID

```bash
curl http://localhost:8080/orders/{order_id}
```

### List All Orders

```bash
curl http://localhost:8080/orders
```

### List Orders by Status

```bash
curl "http://localhost:8080/orders?status=PENDING"
```

Valid statuses: `PENDING`, `PROCESSING`, `SHIPPED`, `DELIVERED`, `CANCELLED`

### Update Order Status

```bash
curl -X PATCH http://localhost:8080/orders/{order_id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "SHIPPED"}'
```

Allowed transitions:

| From | To |
|------|----|
| `PENDING` | `PROCESSING`, `CANCELLED` |
| `PROCESSING` | `SHIPPED` |
| `SHIPPED` | `DELIVERED` |

Use the cancel endpoint (below) instead of PATCH for cancellation.

### Cancel Order

```bash
curl -X POST http://localhost:8080/orders/{order_id}/cancel
```

Only works when order status is `PENDING`.

## Background Worker

On startup, a goroutine runs every 5 minutes (configurable via `STATUS_UPDATE_INTERVAL`) and moves all `PENDING` orders to `PROCESSING`. For local testing:

```bash
STATUS_UPDATE_INTERVAL=10s go run ./cmd/server
```

## Design Notes

- **In-memory storage** — thread-safe map with `sync.RWMutex`; swap the repository interface for PostgreSQL/MySQL in production
- **Layered architecture** — domain → repository → service → handler keeps business rules testable
- **Status validation** — enforced in the domain layer with explicit transition rules
- **Graceful shutdown** — SIGINT/SIGTERM stops the worker and HTTP server cleanly

## AI-Assisted Development

This project was built with Cursor AI assistance for:

- Scaffolding the layered Go project structure
- Defining status transition rules and validation
- Writing HTTP handlers with Go 1.22+ method-based routing
- Creating unit and integration tests
- Documenting the API and setup instructions

Issues encountered and resolved during development:

1. **Status transition edge cases** — ensured cancellation is only via the dedicated cancel endpoint, not PATCH, to avoid ambiguous API behavior
2. **Thread safety** — repository returns cloned orders to prevent accidental mutation of stored state
3. **Empty list responses** — list endpoint returns `[]` instead of `null` for better client compatibility
# order_processing_assignment
