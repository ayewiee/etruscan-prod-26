package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func NewRedisClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // no password in docker
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     50,
	})

	// Ping on startup
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic("redis connection failed: " + err.Error())
	}

	return &Client{rdb}
}
