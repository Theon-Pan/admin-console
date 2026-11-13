// Demo for the redis package: connect to Redis/Valkey and run common operations.
//
// Run:
//
//	REDIS_ADDR=localhost:6379 REDIS_PASSWORD= go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"redis-example/redis"
)

const opTimeout = 500 * time.Millisecond

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client, err := redis.NewClient(addr, os.Getenv("REDIS_PASSWORD"))
	if err != nil {
		log.Fatalf("unable to create redis client: %v", err)
	}
	defer client.Close()

	if err := redis.Ping(client); err != nil {
		log.Fatalf("unable to connect to redis: %v", err)
	}
	fmt.Println("connected:", redis.ServerVersion(client))

	// In a web handler, use the request context (r.Context()) instead.
	ctx := context.Background()

	must(stringOps(ctx, client))
	must(hashOps(ctx, client))
	must(pipelineOps(ctx, client))
	must(concurrentOps(ctx, client))

	s := redis.PoolStats(client)
	fmt.Printf("\npool stats: total=%d idle=%d hits=%d misses=%d timeouts=%d\n",
		s.TotalConns, s.IdleConns, s.Hits, s.Misses, s.Timeouts)

	cctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	client.Del(cctx, "demo:str", "demo:user:1", "demo:counter")
}

func stringOps(parent context.Context, client *goredis.Client) error {
	fmt.Println("\n== strings ==")
	ctx, cancel := context.WithTimeout(parent, opTimeout)
	defer cancel()

	if err := client.Set(ctx, "demo:str", "hello", 10*time.Second).Err(); err != nil {
		return err
	}
	val, err := client.Get(ctx, "demo:str").Result()
	if err != nil {
		return err
	}
	fmt.Printf("GET demo:str = %q\n", val)

	// redis.Nil means the key does not exist; it is not a failure.
	if _, err := client.Get(ctx, "demo:missing").Result(); errors.Is(err, goredis.Nil) {
		fmt.Println("GET demo:missing -> key does not exist")
	} else if err != nil {
		return err
	}
	return nil
}

func hashOps(parent context.Context, client *goredis.Client) error {
	fmt.Println("\n== hashes ==")
	ctx, cancel := context.WithTimeout(parent, opTimeout)
	defer cancel()

	if err := client.HSet(ctx, "demo:user:1", map[string]any{
		"name": "alice", "role": "admin", "logins": 0,
	}).Err(); err != nil {
		return err
	}
	if err := client.HIncrBy(ctx, "demo:user:1", "logins", 1).Err(); err != nil {
		return err
	}
	user, err := client.HGetAll(ctx, "demo:user:1").Result()
	fmt.Println("HGETALL demo:user:1 =", user)
	return err
}

// One context for the whole pipeline, which is sent in a single round trip.
func pipelineOps(parent context.Context, client *goredis.Client) error {
	fmt.Println("\n== pipeline ==")
	ctx, cancel := context.WithTimeout(parent, opTimeout)
	defer cancel()

	var incr *goredis.IntCmd
	_, err := client.Pipelined(ctx, func(p goredis.Pipeliner) error {
		p.Set(ctx, "demo:counter", 100, time.Minute)
		incr = p.IncrBy(ctx, "demo:counter", 5)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println("counter after pipeline =", incr.Val())
	return nil
}

// Many goroutines share one client; the pool handles connections safely.
func concurrentOps(parent context.Context, client *goredis.Client) error {
	fmt.Println("\n== concurrent goroutines ==")
	if err := client.Set(parent, "demo:counter", 0, time.Minute).Err(); err != nil {
		return err
	}

	const workers, perWorker = 50, 100
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				ctx, cancel := context.WithTimeout(parent, opTimeout)
				err := client.Incr(ctx, "demo:counter").Err()
				cancel()
				if err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	if err := <-errCh; err != nil {
		return err
	}

	n, err := client.Get(parent, "demo:counter").Int()
	fmt.Printf("%d goroutines x %d INCR = %d (expected %d) in %v\n",
		workers, perWorker, n, workers*perWorker, time.Since(start))
	return err
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
