package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-redis/redis"
)

var (
	debug     bool
	redisAddr = os.Getenv("REDIS_URL")
)

func init() {
	debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

func main() {
	ll := log.WithField("debug", debug)
	if redisAddr == "" {
		srv, err := miniredis.Run()
		if err != nil {
			log.WithError(err).Fatal("failed to start miniredis")
		}
		redisAddr = srv.Addr()
		defer srv.Close()
	}
	ll = ll.WithField("redis", redisAddr)

	redisClient, err := connectRedis(redisAddr)
	if err != nil {
		ll.WithError(err).Fatal("failed to parse redis URL")
	}
	defer redisClient.Close()
	ll.Info("connected to redis")

	mux := chi.NewMux()
	mux.Use(
		maxContentLength(1<<20), // 1MB
		middleware.RealIP,
		middleware.RequestID,
		middleware.Timeout(time.Second*10),
		middleware.Logger,
		middleware.Recoverer,
	)

	handler := NewHandler(redisClient)
	handler.RegisterRoutes(mux)

	err = http.ListenAndServe(":3000", mux)
	if err != nil {
		ll = ll.WithError(err)
	}
	ll.Error("shutting down")
}

func connectRedis(addr string) (*redis.Client, error) {
	// Create a redis client.
	prefix := "redis://"
	if !strings.HasPrefix(addr, prefix) {
		addr = prefix + addr
	}

	option, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(option), nil
}

func maxContentLength(max int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > max {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
