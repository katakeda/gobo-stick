package main

import (
	"os"

	"github.com/go-redis/redis"
)

// RedisClient ...
type RedisClient struct {
	client *redis.Client
}

func (r *RedisClient) init() {
	var client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})

	r.client = client
}

var redisClient RedisClient
