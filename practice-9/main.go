package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	task := flag.String("task", "retry", "scenario to run: retry or idempotency")
	redisAddr := flag.String("redis-addr", "localhost:6379", "Redis address for the idempotency task")
	redisPassword := flag.String("redis-password", "", "Redis password for the idempotency task")
	redisDB := flag.Int("redis-db", 0, "Redis database number for the idempotency task")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	switch *task {
	case "retry":
		if err := runRetryScenario(logger); err != nil {
			logger.Fatalf("retry scenario failed: %v", err)
		}
	case "idempotency":
		cfg := RedisConfig{
			Addr:     *redisAddr,
			Password: *redisPassword,
			DB:       *redisDB,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		store, err := NewRedisIdempotencyStore(ctx, cfg)
		if err != nil {
			logger.Fatalf("failed to connect to redis: %v", err)
		}
		defer store.Close()

		if err := runIdempotencyScenario(ctx, logger, store); err != nil {
			logger.Fatalf("idempotency scenario failed: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown task %q, use retry or idempotency\n", *task)
		os.Exit(2)
	}
}
