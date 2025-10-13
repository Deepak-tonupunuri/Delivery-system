package services

import (
    "context"
    "fmt"
    "time"

    "delivery-system/internal/database"
    "sync"
    "github.com/go-redis/redis/v8"
)

var (
    orderQueue = make(chan int64, 100)
    cancelMu sync.Mutex
    cancelled = map[int64]bool{}
)

func EnqueueOrder(id int64) {
    select {
    case orderQueue <- id:
    default:
        // drop if queue full
    }
}

func SignalCancel(orderID int64) {
    cancelMu.Lock()
    cancelled[orderID] = true
    cancelMu.Unlock()
    // also set in redis for external visibility
    if database.RDB != nil {
        _ = database.RDB.Set(context.Background(), fmt.Sprintf("order:%d:cancelled", orderID), "1", 0).Err()
    }
}

func isCancelled(orderID int64) bool {
    cancelMu.Lock()
    defer cancelMu.Unlock()
    return cancelled[orderID]
}

// StartOrderProcessor consumes the local queue and performs status changes asynchronously
func StartOrderProcessor(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case id := <-orderQueue:
            go process(ctx, id)
        }
    }
}

func process(ctx context.Context, orderID int64) {
    statuses := []string{"dispatched", "in_transit", "delivered"}
    for _, s := range statuses {
        // before each step, check cancelled from DB and local map
        if isCancelled(orderID) {
            return
        }
        // also check Redis canceled flag
        if database.RDB != nil {
            val, _ := database.RDB.Get(ctx, fmt.Sprintf("order:%d:cancelled", orderID)).Result()
            if val == "1" {
                SignalCancel(orderID)
                return
            }
        }

        // small sleep to simulate time
        time.Sleep(5 * time.Second)

        if isCancelled(orderID) {
            return
        }
        // update DB
        _, _ = database.DB.ExecContext(ctx, "UPDATE orders SET status=$1 WHERE id=$2", s, orderID)
        // publish to redis for real-time tracking
        if database.RDB != nil {
            _ = database.RDB.Set(ctx, fmt.Sprintf("order:%d:status", orderID), s, 0).Err()
        }
    }
}
