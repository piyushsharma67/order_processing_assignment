# Order Processing System

A Go e-commerce order processing app with a web UI, REST API, JWT auth, and MongoDB persistence. Customers browse a product catalog, place orders, track status, and cancel pending orders. A background worker moves orders from `PENDING` to `PROCESSING` 10 seconds after they are placed.

## Features

- **Web UI** — shop catalog, cart, place orders, view orders with live status updates
- **REST API** — create, get, list, update status, and cancel orders
- **Authentication** — sign in / sign up with JWT
- **Background job** — auto-updates `PENDING` → `PROCESSING` after a 10-second delay
- **MongoDB** — orders, users, and products stored in MongoDB when run via Docker

## Project Structure

```
cmd/server/              # Application entry point
internal/domain/         # Order models, status rules, validation
internal/repository/     # In-memory and MongoDB persistence
internal/service/        # Business logic
internal/handler/        # HTTP handlers and routing
internal/worker/         # Background status updater
web/                     # Web UI (HTML/CSS/JS)
```

## Requirements

- Docker and Docker Compose (recommended)
- Go 1.22+ (optional, for local development without Docker)

## How to Run

From the project root:

```bash
APP_HOST_PORT=8081 docker compose up --build
```

To run in the background:

```bash
APP_HOST_PORT=8081 docker compose up --build -d
```

To stop:

```bash
docker compose down
```

## Web UI

Open in your browser:

**http://localhost:8081/**

### Login credentials

| Username | Password |
|----------|----------|
| `admin`    | `admin`    |

You can also create a new account from the **Create account** tab on the login page.

### Using the UI

1. Sign in with `admin` / `admin`
2. **Shop** — add products to cart and place an order
3. **My orders** — view orders; status updates automatically every few seconds
4. Cancel an order while it is still **PENDING**

## Health Check

```bash
curl http://localhost:8081/health
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_HOST_PORT` | `8080` | Host port mapped to the app (use `8081` as shown above) |
| `AUTH_USERNAME` | `admin` | Default admin username |
| `AUTH_PASSWORD` | `admin` | Default admin password |
| `PENDING_PROCESS_DELAY` | `10s` | Wait before PENDING → PROCESSING |
| `STATUS_UPDATE_INTERVAL` | `5s` | Background job poll interval |
| `MONGODB_URI` | `mongodb://mongodb:27017` | MongoDB connection (Docker) |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |

## Order Status Flow

| Status | Description |
|--------|-------------|
| `PENDING` | Order just placed; can be cancelled |
| `PROCESSING` | Set automatically ~10 seconds after placement |
| `SHIPPED` | Updated via API |
| `DELIVERED` | Updated via API |
| `CANCELLED` | Cancelled while pending |

## Run Locally (without Docker)

```bash
go mod tidy
go test ./...
go run ./cmd/server
```

Default URL: **http://localhost:8080/** (uses in-memory storage unless `MONGODB_URI` is set)

## API Reference

All API requests require a JWT token (`Authorization: Bearer <token>`) except `/auth/login`, `/auth/signup`, and `/health`.

### Sign in

```bash
curl -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

### List orders

```bash
curl http://localhost:8081/orders \
  -H "Authorization: Bearer <token>"
```

### Create order

```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "customer_id": "<customer_id>",
    "items": [
      {"product_id": "B08N5WRWNW", "product_name": "Echo Dot", "quantity": 1, "unit_price": 49.99}
    ]
  }'
```

### Cancel order

```bash
curl -X POST http://localhost:8081/orders/{order_id}/cancel \
  -H "Authorization: Bearer <token>"
```

Valid status transitions:

| From | To |
|------|----|
| `PENDING` | `PROCESSING`, `CANCELLED` |
| `PROCESSING` | `SHIPPED` |
| `SHIPPED` | `DELIVERED` |
