package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"messaging-cli/internal/domain"
	"messaging-cli/internal/repository/postgres"
	"strings"
)

func NewWatermillRouter(
	rdb *redis.Client,
	orderRepository *postgres.OrderRepository,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	ctx := context.Background()
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
			var orderCreated domain.OrderCreated
			err := json.Unmarshal(msg.Payload, &orderCreated)
			if err != nil {
				return err
			}

			fmt.Println("Got order-created message")
			fmt.Printf("OrderID: %s\n", orderCreated.OrderID)
			for _, prod := range orderCreated.ProductIDs {
				fmt.Printf("ProductID: %s\n", prod)
			}

			productIDs := strings.Join(orderCreated.ProductIDs, ",")
			order := domain.Order{
				ID:         orderCreated.OrderID,
				ProductIDs: productIDs,
				Status:     domain.OrderStatusNew,
			}

			err = orderRepository.Create(ctx, &order)
			if err != nil {
				return err
			}

			return nil
		})

	router.AddConsumerHandler(
		"order-completed-handler",
		"order-completed",
		orderCompletedSub,
		func(msg *message.Message) error {
			var orderCompleted domain.OrderCompleted
			err := json.Unmarshal(msg.Payload, &orderCompleted)
			if err != nil {
				return err
			}

			fmt.Println("Got order-completed message")
			fmt.Printf("OrderID: %s\n", orderCompleted.OrderID)

			err = orderRepository.Complete(ctx, orderCompleted.OrderID)
			if err != nil {
				return err
			}

			return nil
		})

	return router
}
