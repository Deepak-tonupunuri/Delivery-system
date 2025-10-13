package handlers

import (
    "net/http"
    "github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/register", registerHandler).Methods("POST")
    api.HandleFunc("/login", loginHandler).Methods("POST")

    // protected routes
    api.HandleFunc("/orders", authMiddleware(createOrderHandler)).Methods("POST")
    api.HandleFunc("/orders", authMiddleware(listOrdersHandler)).Methods("GET")
    api.HandleFunc("/orders/{id}/cancel", authMiddleware(cancelOrderHandler)).Methods("PUT")

    // admin
    api.HandleFunc("/admin/orders", authMiddleware(adminListOrders)).Methods("GET")
    api.HandleFunc("/admin/orders/{id}/cancel", authMiddleware(adminCancelOrder)).Methods("PUT")

    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Delivery Management System API"))
    }).Methods("GET")
}
