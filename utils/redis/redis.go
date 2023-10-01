package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Context = context.Background()
var Client *redis.Client

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := Client.Ping(Context).Result()
	if err != nil {
		log.Println("Failed to connect to Redis:", err)
		Client = nil // Set client to nil to indicate failure
	} else {
		fmt.Println("Connected to Redis")
	}
}

func SetWithExpired(key string, value string, expiredTime time.Duration) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	// Set the tokenString in Redis with a 24-hour expiration time
	err := Client.Set(Context, key, value, expiredTime).Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func IsExisted(key string, value string) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	// Check if the tokenString exists in Redis
	val, err := Client.Get(Context, key).Result()

	if err != nil {
		return false, err
	}
	return val == value, nil
}

func Delete(key string) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	err := Client.Del(Context, key).Err()
	if err != nil {
		return false, err
	}
	return true, nil
}
