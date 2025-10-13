package handlers

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"

    "delivery-system/internal/auth"
    "delivery-system/internal/database"
    "delivery-system/internal/models"
    "delivery-system/internal/services"

    "github.com/gorilla/mux"
)

type registerReq struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Role     string `json:"role"`
}

func writeJSON(w http.ResponseWriter, v interface{}, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(v)
}

// register
func registerHandler(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    // NOTE: password stored plaintext here for brevity (DO NOT DO IN PROD)
    var id int64
    err := database.DB.QueryRowContext(r.Context(),
        "INSERT INTO users (username, password, role) VALUES ($1,$2,$3) RETURNING id",
        req.Username, req.Password, req.Role).Scan(&id)
    if err != nil {
        http.Error(w, "cannot create user: "+err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, map[string]interface{}{"id": id}, http.StatusCreated)
}

type loginReq struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    var id int64
    var role string
    var pwd string
    err := database.DB.QueryRowContext(r.Context(),
        "SELECT id, password, role FROM users WHERE username=$1", req.Username).
        Scan(&id, &pwd, &role)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "bad credentials", http.StatusUnauthorized)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if pwd != req.Password {
        http.Error(w, "bad credentials", http.StatusUnauthorized)
        return
    }
    tok, err := auth.GenerateToken(id, role)
    if err != nil {
        http.Error(w, "token error", http.StatusInternalServerError)
        return
    }
    writeJSON(w, map[string]string{"token": tok}, http.StatusOK)
}

// create order
type createOrderReq struct {
    Item string `json:"item"`
}

func createOrderHandler(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
    var req createOrderReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    var id int64
    err := database.DB.QueryRowContext(r.Context(),
        "INSERT INTO orders (user_id, item, status) VALUES ($1,$2,$3) RETURNING id",
        claims.UserID, req.Item, "created").Scan(&id)
    if err != nil {
        http.Error(w, "cannot create order", http.StatusInternalServerError)
        return
    }
    // enqueue for processing
    services.EnqueueOrder(id)
    writeJSON(w, map[string]interface{}{"order_id": id}, http.StatusCreated)
}

// list orders (customers -> own orders, admin -> all)
func listOrdersHandler(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
    rows, err := database.DB.QueryContext(r.Context(),
        "SELECT id, user_id, item, status, cancelled, created_at FROM orders WHERE user_id=$1 ORDER BY id DESC", claims.UserID)
    if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    var out []models.Order
    for rows.Next() {
        var o models.Order
        if err := rows.Scan(&o.ID, &o.UserID, &o.Item, &o.Status, &o.Cancelled, &o.CreatedAt); err != nil {
            continue
        }
        out = append(out, o)
    }
    writeJSON(w, out, http.StatusOK)
}

func cancelOrderHandler(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    oid, _ := strconv.ParseInt(idStr, 10, 64)
    // ensure owner
    var owner int64
    err := database.DB.QueryRowContext(r.Context(), "SELECT user_id FROM orders WHERE id=$1", oid).Scan(&owner)
    if err != nil {
        http.Error(w, "order not found", http.StatusNotFound)
        return
    }
    if owner != claims.UserID && claims.Role != "admin" {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }
    _, err = database.DB.ExecContext(r.Context(), "UPDATE orders SET cancelled=TRUE WHERE id=$1", oid)
    if err != nil {
        http.Error(w, "cannot cancel", http.StatusInternalServerError)
        return
    }
    // signal cancellation to processor via Redis
    services.SignalCancel(oid)
    writeJSON(w, map[string]bool{"cancelled": true}, http.StatusOK)
}

// admin handlers
func adminListOrders(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
    if claims.Role != "admin" {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }
    rows, err := database.DB.QueryContext(r.Context(),
        "SELECT id, user_id, item, status, cancelled, created_at FROM orders ORDER BY id DESC")
    if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    var out []models.Order
    for rows.Next() {
        var o models.Order
        if err := rows.Scan(&o.ID, &o.UserID, &o.Item, &o.Status, &o.Cancelled, &o.CreatedAt); err != nil {
            continue
        }
        out = append(out, o)
    }
    writeJSON(w, out, http.StatusOK)
}

func adminCancelOrder(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
    if claims.Role != "admin" {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }
    vars := mux.Vars(r)
    idStr := vars["id"]
    oid, _ := strconv.ParseInt(idStr, 10, 64)
    _, err := database.DB.ExecContext(r.Context(), "UPDATE orders SET cancelled=TRUE WHERE id=$1", oid)
    if err != nil {
        http.Error(w, "cannot cancel", http.StatusInternalServerError)
        return
    }
    services.SignalCancel(oid)
    writeJSON(w, map[string]bool{"cancelled": true}, http.StatusOK)
}
