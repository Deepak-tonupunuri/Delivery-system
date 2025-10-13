package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "delivery-system/internal/database"
    "delivery-system/internal/handlers"
    "delivery-system/internal/services"

    "github.com/gorilla/mux"
)

func main() {
    // Read config from env (fallbacks provided)
    pgUrl := os.Getenv("DATABASE_URL")
    if pgUrl == "" {
        pgUrl = "postgres://postgres:admin@localhost:5432/deliverydb?sslmode=disable"
    }
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }

    ctx := context.Background()
    // Init DB and Redis
    if err := database.InitPostgres(ctx, pgUrl); err != nil {
        log.Fatalf("postgres init: %v", err)
    }
    if err := database.InitRedis(ctx, redisAddr); err != nil {
        log.Fatalf("redis init: %v", err)
    }

    // Start background service manager
    go services.StartOrderProcessor(ctx)

    r := mux.NewRouter()
    handlers.RegisterRoutes(r)

    srv := &http.Server{
        Handler:      r,
        Addr:         ":8080",
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    log.Println("server starting on :8080")
    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
