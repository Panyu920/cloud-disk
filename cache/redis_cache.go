package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "panyu",
		Username: "panyu",
		DB:       0,
		OnConnect: func(ctx context.Context, conn *redis.Conn) error {
			log.Println("OnConnect redis success")
			return nil
		},
	})
}
