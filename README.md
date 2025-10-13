# Delivery Management System (Go) - Skeleton

## What is included
- REST API in Go (net/http + gorilla/mux)
- PostgreSQL integration (simple migrations created at startup)
- Redis integration for status caching and cancellation signaling
- JWT-based authentication (customers and admins)
- Background goroutines to simulate order lifecycle (created -> dispatched -> in_transit -> delivered)
- Dockerfile and docker-compose for easy local deployment
- Basic automated tests skeleton

## Quickstart (with docker-compose)
1. Build and start services:
   ```
   docker-compose up --build
   ```
2. The API will be available at `http://localhost:8080`.
3. Endpoints:
   - `POST /api/register` - {"username","password","role"}
   - `POST /api/login` - {"username","password"} -> returns JWT
   - `POST /api/orders` - create order (Authorization: Bearer <token>)
   - `GET /api/orders` - list my orders
   - `PUT /api/orders/{id}/cancel` - cancel order
   - `GET /api/admin/orders` - admin list (requires admin token)

## Notes & Design Decisions
- Passwords are stored in plaintext in this skeleton for brevity. **Hash passwords in real projects.**
- JWT secret is configurable via `JWT_KEY` env var.
- Redis is used to publish status and cancellation flags.
- The order processor is intentionally simple and uses an in-memory queue + goroutines.
- Tests live in `internal/tests` and demonstrate basic flows.

## Run tests
```
go test ./...
```

## Files of interest
- `cmd/main.go` - app entrypoint
- `internal/database` - Postgres and Redis initialization
- `internal/handlers` - HTTP routes and handlers
- `internal/services` - background order processor
