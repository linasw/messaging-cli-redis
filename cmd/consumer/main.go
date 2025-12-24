package main

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"messaging-cli/internal/redis"
	"os"
	"os/signal"
)

func main() {
	watermillLogger := watermill.NewSlogLogger(nil)
	redisClient := redis.NewRedisClient()
	router := redis.NewWatermillRouter(redisClient, watermillLogger)

	consumerError := make(chan error, 1)
	go func() {
		fmt.Println("Starting consumer...")
		consumerError <- router.Run(context.Background())
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-consumerError:
		panic(err)
	case _ = <-shutdown:
		err := router.Close()
		if err != nil {
			panic(err)
		}
	}
}
