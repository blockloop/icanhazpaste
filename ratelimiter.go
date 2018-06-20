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
	rate := limiter.Rate{
		Limit:  20,
		Period: time.Hour,
	}

	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "ratelimiter",
		MaxRetry: 3,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to create redis limiter store")
	}

	return stdlib.NewMiddleware(limiter.New(store, rate)).Handler
}
