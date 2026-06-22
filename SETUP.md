# Order Processing — Setup & Run Guide

Step-by-step instructions to run the server, open the web UI, and sign in.

---

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| **Go** | 1.22+ | Required for local (non-Docker) runs |
| **Docker** | Recent | Required for Docker Compose setup |
| **Docker Compose** | v2+ | Usually included with Docker Desktop |

---

## Option 1: Run with Docker Compose (recommended)

This starts the app and MongoDB together.

### 1. Start the services

From the project root:

```bash
cd /home/shikha/go/src/order_processing
docker compose up --build
```

To run in the background:

```bash
docker compose up --build -d
```

### 2. Wait until the app is ready

You should see logs like:

```text
server listening on http://localhost:8080
pending orders move to PROCESSING after 10s (poll interval=5s)
```

MongoDB must be healthy before the app starts (handled automatically by `depends_on`).

### 3. Stop the services

```bash
docker compose down
```

To also remove the MongoDB data volume:

```bash
docker compose down -v
```

---

## Option 2: Run locally with Go

Use this if you do not want Docker. Data is stored **in memory** unless you set `MONGODB_URI`.

### 1. Install dependencies

```bash
cd /home/shikha/go/src/order_processing
go mod tidy
```

### 2. Start the server

```bash
go run ./cmd/server
```

The server listens on **port 8080** by default.

### 3. Optional: use MongoDB locally

If MongoDB is already running on your machine:

```bash
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=order_processing
go run ./cmd/server
```

---

## Access the Web UI

Once the server is running, open a browser and go to:

| URL | Description |
|-----|-------------|
| **http://localhost:8080** | Main web UI (shop + orders) |
| **http://localhost:8080/health** | Health check (`{"status":"ok"}`) |

If you changed the host port in Docker, use that port instead (see `APP_HOST_PORT` below).

---

## Login credentials

### Default admin account

The server creates a default user on startup:

| Field | Value |
|-------|-------|
| **Username** | `admin` |
| **Password** | `admin` |

These defaults apply when:

- Running with **Docker Compose** (`AUTH_USERNAME` / `AUTH_PASSWORD` in `docker-compose.yml`)
- Running **locally** without overriding env vars (`cmd/server/main.go` defaults)

### Sign in

1. Open **http://localhost:8080**
2. On the sign-in screen, enter:
   - Username: `admin`
   - Password: `admin`
3. Click **Sign in**

### Create your own account

You can also use **Create account** on the login page to register a new user. Each account gets its own customer ID and only sees its own orders.

---

## Using the UI

### Shop tab

1. Browse the product catalog.
2. Add items to your cart and adjust quantities.
3. Click **Place order**.
4. You are redirected to **My orders** automatically.

### My orders tab

- View all orders for your account.
- Filter by status (Pending, Processing, Shipped, etc.).
- **Cancel** an order while it is still **PENDING**.
- Status updates **automatically every 3 seconds** — no manual refresh needed.

### Order status flow

| Time | Status |
|------|--------|
| Right after placing an order | `PENDING` |
| After ~10 seconds | Background job moves it to `PROCESSING` |
| Later (manual API update) | `SHIPPED` → `DELIVERED` |

The background worker checks every **5 seconds** and only processes orders that have been pending for at least **10 seconds**.

---

## Environment variables

### Docker Compose defaults (`docker-compose.yml`)

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_HOST_PORT` | `8080` | Port on your machine (maps to container 8080) |
| `PORT` | `8080` | Port inside the app container |
| `MONGODB_URI` | `mongodb://mongodb:27017` | MongoDB connection string |
| `MONGODB_DATABASE` | `order_processing` | Database name |
| `AUTH_USERNAME` | `admin` | Default admin username |
| `AUTH_PASSWORD` | `admin` | Default admin password |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `JWT_EXPIRY` | `24h` | Token lifetime |
| `STATUS_UPDATE_INTERVAL` | `5s` | How often the background job runs |
| `PENDING_PROCESS_DELAY` | `10s` | Wait time before PENDING → PROCESSING |

### Local run defaults (no env vars set)

| Variable | Default |
|----------|---------|
| `PORT` | `8080` |
| `AUTH_USERNAME` | `admin` |
| `AUTH_PASSWORD` | `admin` |
| `JWT_SECRET` | `dev-jwt-secret-change-me` |
| `JWT_EXPIRY` | `24h` |
| `STATUS_UPDATE_INTERVAL` | `5s` |
| `PENDING_PROCESS_DELAY` | `10s` |
| `MONGODB_URI` | *(empty — uses in-memory storage)* |

### Example: custom port and faster status updates

```bash
PORT=3000 \
STATUS_UPDATE_INTERVAL=3s \
PENDING_PROCESS_DELAY=10s \
go run ./cmd/server
```

Then open **http://localhost:3000**.

---

## Verify the server is running

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

---

## Run tests

```bash
go test ./...
```

---

## Quick reference

| What | Where / Value |
|------|----------------|
| Web UI | http://localhost:8080 |
| Health check | http://localhost:8080/health |
| Default login | `admin` / `admin` |
| Start (Docker) | `docker compose up --build` |
| Start (Go) | `go run ./cmd/server` |
| Stop (Docker) | `docker compose down` |

---

## Troubleshooting

| Problem | What to try |
|---------|-------------|
| Port 8080 already in use | Set `APP_HOST_PORT=8081` (Docker) or `PORT=8081` (Go), then use that port in the browser |
| UI shows old behavior | Hard refresh: `Ctrl+Shift+R` (Windows/Linux) or `Cmd+Shift+R` (Mac) |
| Cannot sign in | Confirm the server is running and use `admin` / `admin` unless you changed `AUTH_USERNAME` / `AUTH_PASSWORD` |
| Orders stay PENDING | Wait at least 10–15 seconds; stay on **My orders** (auto-refresh is enabled) |
| Docker build fails | Ensure Docker is running and you are in the project root |

For API details (curl examples, status transitions), see [README.md](./README.md).
