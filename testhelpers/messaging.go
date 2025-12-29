package testhelpers

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/require"
	"messaging-cli/internal/redis"
	"messaging-cli/internal/repository/postgres"
	"testing"
	"time"
)

type ConsumerSetup struct {
	Router        *message.Router
	Publisher     message.Publisher
	ConsumerError chan error
	Context       context.Context
	Cancel        context.CancelFunc
}

func SetupConsumer(t *testing.T, orderRepo *postgres.OrderRepository) *ConsumerSetup {
	t.Helper()

	redisClient := redis.NewRedisClient()
	watermillLogger := watermill.NewStdLogger(false, false)

	pub := redis.NewRedisPublisher(redisClient, watermillLogger)
	router := redis.NewWatermillRouter(redisClient, orderRepo, watermillLogger)

	ctx, cancel := context.WithCancel(context.Background())
	consumerError := make(chan error, 1)

	go func() {
		consumerError <- router.Run(ctx)
	}()

	time.Sleep(2 * time.Second)

	//check if consumer failed to start
	select {
	case err := <-consumerError:
		t.Fatalf("Consumer failed to start: %v", err)
	default:
		//consumer is running, continue with test
	}

	return &ConsumerSetup{
		Router:        router,
		Publisher:     pub,
		ConsumerError: consumerError,
		Context:       ctx,
		Cancel:        cancel,
	}
}

func CleanupConsumer(t *testing.T, setup *ConsumerSetup) {
	t.Helper()

	setup.Cancel()
	err := setup.Router.Close()
	require.NoError(t, err)
}

func CheckConsumerError(t *testing.T, setup *ConsumerSetup) {
	t.Helper()

	select {
	case err := <-setup.ConsumerError:
		t.Fatalf("Consumer failed: %v", err)
	default:
	}
}
