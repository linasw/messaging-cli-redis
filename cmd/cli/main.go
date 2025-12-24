package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"messaging-cli/internal/entities"
	"messaging-cli/internal/redis"
	"os"
)

type arrayFlags []string

func (p *arrayFlags) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *arrayFlags) Set(value string) error {
	*p = append(*p, value)
	return nil
}

var productIds arrayFlags

func main() {
	orderCreatedCmd := flag.NewFlagSet("order-created", flag.ExitOnError)
	orderCreatedIdPtr := orderCreatedCmd.String("orderId", "", "order id")
	orderCreatedCmd.Var(&productIds, "productIds", "product ids")

	orderCompletedCmd := flag.NewFlagSet("order-completed", flag.ExitOnError)
	orderCompletedIdPtr := orderCompletedCmd.String("orderId", "", "order id")

	if len(os.Args) < 2 {
		fmt.Println("expected 'order-created' or 'order-completed' subcommands")
		os.Exit(1)
	}

	watermillLogger := watermill.NewSlogLogger(nil)
	redisClient := redis.NewRedisClient()
	pub := redis.NewRedisPublisher(redisClient, watermillLogger)

	switch os.Args[1] {
	case "order-created":
		err := orderCreatedCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("Creating topic: order-created")
		fmt.Println("    Order ID:", *orderCreatedIdPtr)
		for _, id := range productIds {
			fmt.Println("    Product ID:", id)
		}

		orderCreated := entities.OrderCreated{
			OrderID:    *orderCreatedIdPtr,
			ProductIDs: productIds,
		}
		payload, err := json.Marshal(&orderCreated)
		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = pub.Publish("order-created", msg)
		if err != nil {
			fmt.Println("error:", err)
		}
	case "order-completed":
		err := orderCompletedCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("Creating topic: order-completed")
		fmt.Println("    Order ID:", *orderCompletedIdPtr)

		orderCompleted := entities.OrderCompleted{
			OrderID: *orderCompletedIdPtr,
		}
		payload, err := json.Marshal(&orderCompleted)
		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = pub.Publish("order-completed", msg)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
}
