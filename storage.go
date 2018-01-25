package main

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// DefaultFileTTL is the default amount of time files are active
const DefaultFileTTL = time.Hour * 72

type Store struct {
	client *redis.Client
	ttl    time.Duration
}

func NewStoreTTL(client *redis.Client, ttl time.Duration) *Store {
	s := &Store{
		client: client,
	}
	s.SetTTL(ttl)
	return s
}

func NewStore(client *redis.Client) *Store {
	return NewStoreTTL(client, DefaultFileTTL)
}

func (s *Store) SetTTL(dur time.Duration) {
	if dur.Nanoseconds() <= 0 {
		panic("ttl must be greater than zero")
	}
	s.ttl = dur
}

func (s *Store) Put(name, text string) error {
	st := s.client.Set(name, text, s.ttl)
	return errors.Wrap(st.Err(), "failed to put item")
}

func (s *Store) Get(name string) (text string, expires time.Time, err error) {
	tx := s.client.TxPipeline()
	defer tx.Close()

	body := tx.Get(name)
	ttl := tx.TTL(name)

	_, err = tx.Exec()
	if err == redis.Nil {
		err = nil
		return
	}
	if err != nil {
		err = errors.Wrap(err, "bad response from redis")
		return
	}

	text = body.Val()

	if ttl.Val().Nanoseconds() > 0 {
		expires = time.Now().UTC().Add(ttl.Val())
	}
	return
}
