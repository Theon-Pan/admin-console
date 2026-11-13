package redis // import "admin-console/internal/redis

import (
	"context"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// NewClient configures the Redis/Valkey connection pool.
// The returned client is safe for concurrent use; create it once and share it.
// Like sql.Open, connections are established lazily; call Ping to verify.
func NewClient(addr, password string, db, minIdleConns, poolSize int, connMaxLifetime time.Duration) (*goredis.Client, error) {
	if addr == "" {
		return nil, errors.New("redis: address is empty")
	}

	opts := &goredis.Options{Addr: addr, DB: db}
	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		parsed, err := goredis.ParseURL(addr)
		if err != nil {
			return nil, err
		}
		opts = parsed
	}
	if password != "" {
		opts.Password = password
	}

	opts.MinIdleConns = minIdleConns
	opts.PoolSize = poolSize
	opts.ConnMaxIdleTime = 5 * time.Minute
	opts.ConnMaxLifetime = connMaxLifetime
	opts.PoolTimeout = 4 * time.Second
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.MaxRetries = 3
	// Let per-call context deadlines also bound socket reads/writes.
	opts.ContextTimeoutEnabled = true

	client := goredis.NewClient(opts)

	// Verify if the redis client works.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client, client.Ping(ctx).Err()
}
