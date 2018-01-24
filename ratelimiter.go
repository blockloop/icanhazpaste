package main

import (
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/go-redis/redis"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	sredis "github.com/ulule/limiter/drivers/store/redis"
)

func ipRateLimiter(client *redis.Client) func(http.Handler) http.Handler {
	// Define a limit rate to 4 requests per hour.
	rate := limiter.Rate{
		Limit:  200,
		Period: time.Hour * 24,
	}

	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter_http_example",
		MaxRetry: 3,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to create redis limiter store")
	}

	return stdlib.NewMiddleware(limiter.New(store, rate)).Handler
}
