package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "delivery-system/internal/handlers"
    "delivery-system/internal/database"
    "delivery-system/internal/services"

    "github.com/gorilla/mux"
    "context"
)

// Note: These tests are illustrative and may require a running DB/Redis.
func TestRegisterLoginCreateOrder(t *testing.T) {
    // attempt to connect to local DB/Redis — skip if not available
    ctx := context.Background()
    if err := database.InitPostgres(ctx, os.Getenv("DATABASE_URL")); err != nil {
        t.Skip("postgres not available:", err)
    }
    if err := database.InitRedis(ctx, os.Getenv("REDIS_ADDR")); err != nil {
        t.Skip("redis not available:", err)
    }
    // Start processor
    go services.StartOrderProcessor(ctx)

    r := mux.NewRouter()
    handlers.RegisterRoutes(r)

    // Register
    body := map[string]string{"username": "testuser", "password": "pass", "role": "customer"}
    b, _ := json.Marshal(body)
    req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(b))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("register failed: %v %s", w.Code, w.Body.String())
    }

    // Login
    lb, _ := json.Marshal(map[string]string{"username":"testuser","password":"pass"})
    req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(lb))
    w = httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("login failed: %v %s", w.Code, w.Body.String())
    }
    var resp map[string]string
    json.NewDecoder(w.Body).Decode(&resp)
    token := resp["token"]

    // Create order
    ob, _ := json.Marshal(map[string]string{"item":"widget"})
    req = httptest.NewRequest("POST", "/api/orders", bytes.NewReader(ob))
    req.Header.Set("Authorization", "Bearer "+token)
    w = httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("create order failed: %v %s", w.Code, w.Body.String())
    }
}
