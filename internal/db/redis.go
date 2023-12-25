package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisDB *redis.Client

func ConnectRedis() error {
	log.Println("connecting to redis")
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return fmt.Errorf("missing REDIS_URL in environment")
	}

	RedisDB = redis.NewClient(&redis.Options{Addr: url})

	c := RedisDB.Ping(context.Background())
	return c.Err()
}
