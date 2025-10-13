package database

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/go-redis/redis/v8"
    "time"
)

var DB *sql.DB
var RDB *redis.Client

func InitPostgres(ctx context.Context, dsn string) error {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return err
    }
    db.SetConnMaxLifetime(time.Minute * 5)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)

    // Try ping
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        return err
    }
    DB = db
    // Ensure tables exist (simple migration)
    if err := ensureTables(); err != nil {
        return err
    }
    log.Println("postgres connected")
    return nil
}

func ensureTables() error {
    users := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        role TEXT NOT NULL
    );`
    orders := `
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id),
        item TEXT NOT NULL,
        status TEXT NOT NULL,
        cancelled BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT NOW()
    );`
    if _, err := DB.Exec(users); err != nil {
        return err
    }
    if _, err := DB.Exec(orders); err != nil {
        return err
    }
    return nil
}

func InitRedis(ctx context.Context, addr string) error {
    rdb := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    if err := rdb.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("redis ping: %w", err)
    }
    RDB = rdb
    log.Println("redis connected")
    return nil
}
