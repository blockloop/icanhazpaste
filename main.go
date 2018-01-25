package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-redis/redis"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	config = struct {
		redisAddr string
		staticDir string
	}{}

	cli = kingpin.New("pbpaste", "pbpaste server")
)

const (
	// Memory is fleeting
	Memory = "MEMORY"
)

func init() {
	cli.Version("0.0.1")
	cli.Flag("redis-url", "redis url").
		Envar("REDIS_URL").
		Default(Memory).
		StringVar(&config.redisAddr)

	if _, err := cli.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}

func main() {
	if config.redisAddr == Memory {
		srv, err := miniredis.Run()
		if err != nil {
			log.WithError(err).Fatal("failed to start miniredis")
		}
		config.redisAddr = srv.Addr()
		defer srv.Close()
	}

	redisClient, err := connectRedis(config.redisAddr)
	if err != nil {
		log.WithField("addr", config.redisAddr).WithError(err).Fatal("failed to parse redis URL")
	}
	defer redisClient.Close()
	log.WithField("addr", config.redisAddr).Info("connected to redis")

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
	ll := log.WithError(err)
	if err == nil || err == http.ErrServerClosed {
		ll.Info("shutting down")
	} else {
		ll.WithError(err).Error("shutting down")
	}
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
