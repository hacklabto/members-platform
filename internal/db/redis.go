package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis() error {
	log.Println("connecting to redis")
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return fmt.Errorf("missing REDIS_URL in environment")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		return err
	}

	Redis = redis.NewClient(opt)

	c := Redis.Ping(context.Background())
	return c.Err()
}
