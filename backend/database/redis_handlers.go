package database

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/skye-tan/trello/backend/utils/custom_errors"
	"golang.org/x/net/context"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func AddPictureToRedis(task_id string, encoded string) error {
	err := rdb.Set(ctx, task_id, encoded, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func RetrieveFile(task_id string) ([]byte, error) {
	encoded, err := rdb.Get(ctx, task_id).Result()
	if err != nil {
		return nil, custom_errors.ErrPictureLoadFailure
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, custom_errors.ErrPictureLoadFailure
	}

	return decoded, nil
}

func RedisInitial() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "10.5.0.7:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Println(err)
	}
}
