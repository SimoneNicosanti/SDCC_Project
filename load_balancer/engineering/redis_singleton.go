package engineering

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

var redisConn *redis.Client = nil

func getRedisClient() *redis.Client {
	if redisConn == nil {
		redisConn = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	}
	return redisConn
}

func PutOnRedis(key string, value string) error {
	ctx := context.Background()
	err := getRedisClient().Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetFromRedis(key string) (string, error) {
	ctx := context.Background()
	value, err := getRedisClient().Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}
