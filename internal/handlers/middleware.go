package handlers

import (
    "context"
    "net/http"
    "strings"

    "delivery-system/internal/auth"
)

// Adapter to inject claims into handlers
type authHandler func(http.ResponseWriter, *http.Request, *auth.Claims)

func authMiddleware(next authHandler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        h := r.Header.Get("Authorization")
        if h == "" {
            http.Error(w, "missing auth", http.StatusUnauthorized)
            return
        }
        parts := strings.SplitN(h, " ", 2)
        if len(parts) != 2 {
            http.Error(w, "bad auth", http.StatusUnauthorized)
            return
        }
        token := parts[1]
        claims, err := auth.ParseToken(token)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }
        // store claims in context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next(w, r.WithContext(ctx), claims)
    }
}
