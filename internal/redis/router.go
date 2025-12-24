package redis

import (
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"messaging-cli/internal/entities"
)

func NewWatermillRouter(rdb *redis.Client, watermillLogger watermill.LoggerAdapter) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)

	orderCreatedSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "order-created-group",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	orderCompletedSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "order-completed-group",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"order-created-handler",
		"order-created",
		orderCreatedSub,
		func(msg *message.Message) error {
			var orderCreated entities.OrderCreated
			err := json.Unmarshal(msg.Payload, &orderCreated)
			if err != nil {
				return err
			}

			fmt.Println("Got order-created message")
			fmt.Printf("OrderID: %s\n", orderCreated.OrderID)
			for _, prod := range orderCreated.ProductIDs {
				fmt.Printf("ProductID: %s\n", prod)
			}

			return nil
		})

	router.AddConsumerHandler(
		"order-completed-handler",
		"order-completed",
		orderCompletedSub,
		func(msg *message.Message) error {
			var orderCompleted entities.OrderCompleted
			err := json.Unmarshal(msg.Payload, &orderCompleted)
			if err != nil {
				return err
			}

			fmt.Println("Got order-completed message")
			fmt.Printf("OrderID: %s\n", orderCompleted.OrderID)

			return nil
		})

	return router
}
