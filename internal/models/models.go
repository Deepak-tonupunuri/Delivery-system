package models

import "time"

type User struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"` // stored hashed in real apps
    Role     string `json:"role"` // "customer" or "admin"
}

type Order struct {
    ID        int64     `json:"id"`
    UserID    int64     `json:"user_id"`
    Item      string    `json:"item"`
    Status    string    `json:"status"`
    Cancelled bool      `json:"cancelled"`
    CreatedAt time.Time `json:"created_at"`
}
