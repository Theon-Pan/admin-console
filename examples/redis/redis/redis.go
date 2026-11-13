package redis // import "redis-example/redis"

import (
	"context"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Default pool settings used by NewClient.
const (
	DefaultMinIdleConns    = 2
	DefaultPoolSize        = 20
	DefaultConnMaxLifetime = 30 * time.Minute
)

// NewClient creates a Redis/Valkey client with default pool settings.
// addr is either "host:port" or a URL such as "redis://user:pass@host:6379/0"
// ("rediss://" enables TLS). A non-empty password overrides the one in the URL.
func NewClient(addr, password string) (*goredis.Client, error) {
	return NewConnectionPool(addr, password, 0, DefaultMinIdleConns, DefaultPoolSize, DefaultConnMaxLifetime)
}

// NewConnectionPool configures the Redis/Valkey connection pool.
// The returned client is safe for concurrent use; create it once and share it.
// Like sql.Open, connections are established lazily; call Ping to verify.
func NewConnectionPool(addr, password string, db, minIdleConns, poolSize int, connMaxLifetime time.Duration) (*goredis.Client, error) {
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
	opts.ConnMaxLifetime = connMaxLifetime
	opts.ConnMaxIdleTime = 5 * time.Minute
	opts.PoolTimeout = 4 * time.Second
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.MaxRetries = 3
	// Let per-call context deadlines also bound socket reads/writes.
	opts.ContextTimeoutEnabled = true

	return goredis.NewClient(opts), nil
}

// Ping checks if the Redis connection works.
func Ping(client *goredis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}

// ServerVersion returns the version reported by the server
// (redis_version for Redis, valkey_version for Valkey).
func ServerVersion(client *goredis.Client) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return err.Error()
	}

	version := ""
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "valkey_version:"); ok {
			return "valkey " + v
		}
		if v, ok := strings.CutPrefix(line, "redis_version:"); ok {
			version = "redis " + v
		}
	}
	return version
}

// PoolStats returns connection pool statistics.
func PoolStats(client *goredis.Client) *goredis.PoolStats {
	return client.PoolStats()
}
