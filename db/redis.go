package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"url-shortner/log"
)

func NewRedisConnection(host string,
	password string,
	database int,
	port int,
	retryTimeout time.Duration,
	retry int) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       database,
	})
	tickerChannel := time.NewTicker(retryTimeout)
	counter := 0
	ctx := context.Background()
	for ; true; <-tickerChannel.C {
		counter++
		_, err := client.Ping(ctx).Result()
		if err == nil {
			tickerChannel.Stop()

			break
		}

		log.Errorf("Cannot connect to redis %s: %s", host, err)
		if counter >= retry {
			log.Errorf("Cannot connect to redis %s after %d retries: %s", host, counter, err)
			tickerChannel.Stop()

			return client, err
		}
	}

	log.Infof("Connected to redis : %s", host)

	return client, nil
}

func Disconnect(redis *redis.Client) error {
	return redis.Close()
}
