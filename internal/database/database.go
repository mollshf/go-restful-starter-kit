package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/ums/internal/shared/web"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Tuning pool untuk SIAKAD
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

type Database struct {
	db *pgxpool.Pool
}

func NewDatabase(db *pgxpool.Pool) *Database {
	return &Database{db: db}
}

func (d *Database) Health(c *gin.Context) error {
	stats := make(map[string]string)

	stats["status"] = "Up"
	stats["message"] = "It's healthy"

	dbStats := d.db.Stat()
	stats["open_connections"] = fmt.Sprintf("Active connections: %d", dbStats.NewConnsCount())
	stats["in_use"] = strconv.Itoa(int(dbStats.AcquiredConns()))
	stats["idle"] = strconv.Itoa(int(dbStats.IdleConns()))
	stats["max_connections"] = strconv.Itoa(int(dbStats.MaxConns()))
	stats["wait_count"] = strconv.Itoa(int(dbStats.EmptyAcquireCount()))
	stats["wait_duration"] = strconv.Itoa(int(dbStats.EmptyAcquireWaitTime()))
	stats["max_idle_closed"] = strconv.Itoa(int(dbStats.MaxIdleDestroyCount()))
	stats["max_lifetime_closed"] = strconv.Itoa(int(dbStats.MaxLifetimeDestroyCount()))

	// Evaluate stats to provide a health message
	if dbStats.NewConnsCount() > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.EmptyAcquireCount() > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleDestroyCount() > int64(dbStats.NewConnsCount())/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeDestroyCount() > int64(dbStats.NewConnsCount())/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	web.OK(c, stats)

	return nil
}
