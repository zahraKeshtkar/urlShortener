package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx context.Context

func init() {
	ctx = context.Background()
}

func Connect(host string,
	password string,
	database int,
	port int) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       database,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func Disconnect(redis *redis.Client) error {
	return redis.Close()
}
