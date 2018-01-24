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
	if dur <= time.Duration(0) {
		panic("ttl must be greater than zero")
	}
	s.ttl = dur
}

func (s *Store) Put(name, text string) error {
	st := s.client.Set(name, text, s.ttl)
	return errors.Wrap(st.Err(), "failed to put item")
}

func (s *Store) Get(name string) ([]byte, error) {
	st := s.client.Get(name)
	if err := st.Err(); err != nil {
		return nil, err
	}
	return []byte(st.Val()), nil
}
