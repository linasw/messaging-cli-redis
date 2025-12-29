package test

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"messaging-cli/internal/domain"
	"messaging-cli/internal/redis"
	"messaging-cli/internal/repository/postgres"
	"testing"
	"time"
)

//func setupTestDB(t *testing.T) *pgxpool.Pool {
//	connString := "host=localhost port=5432 user=orderuser password=orderpass dbname=orderdb"
//
//	poolConfig, err := pgxpool.ParseConfig(connString)
//	require.NoError(t, err)
//
//	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
//	require.NoError(t, err)
//
//	//clean up Orders DB
//	_, err = pool.Exec(context.Background(), "DELETE FROM orders")
//	require.NoError(t, err)
//
//	return pool
//}

func TestOrderCreatedFlow(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	redisClient := redis.NewRedisClient()
	defer redisClient.Close()

	watermillLogger := watermill.NewStdLogger(false, false)
	pub := redis.NewRedisPublisher(redisClient, watermillLogger)

	orderRepository := postgres.NewOrderRepository(pool)
	router := redis.NewWatermillRouter(redisClient, orderRepository, watermillLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerError := make(chan error, 1)
	go func() {
		consumerError <- router.Run(ctx)
	}()

	//give time for consumer to start
	time.Sleep(2 * time.Second)

	//check if consumer failed to start
	select {
	case err := <-consumerError:
		t.Fatalf("Consumer failed to start: %v", err)
	default:
		//consumer is running, continue with test
	}

	t.Run("order created", func(t *testing.T) {
		orderCreated := domain.OrderCreated{
			OrderID:    "integration-order-1",
			ProductIDs: []string{"product-1", "product-2"},
		}

		payload, err := json.Marshal(orderCreated)
		require.NoError(t, err)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = pub.Publish("order-created", msg)
		require.NoError(t, err)

		//wait for message to get processed
		time.Sleep(2 * time.Second)

		select {
		case err := <-consumerError:
			t.Fatalf("Consumer failed during processing: %v", err)
		default:
		}

		var order domain.Order
		query := "SELECT id, product_ids, status FROM orders WHERE id = $1"
		err = pool.QueryRow(context.Background(), query, orderCreated.OrderID).Scan(
			&order.ID,
			&order.ProductIDs,
			&order.Status,
		)

		require.NoError(t, err)
		assert.Equal(t, orderCreated.OrderID, order.ID)
		assert.Equal(t, "product-1,product-2", order.ProductIDs)
		assert.Equal(t, domain.OrderStatusNew, order.Status)
	})

	cancel()
	err := router.Close()
	assert.NoError(t, err)
}

func TestOrderCompletedFlow(t *testing.T) {
	ctx := context.Background()

	pool := setupTestDB(t)
	defer pool.Close()

	redisClient := redis.NewRedisClient()
	defer redisClient.Close()

	watermillLogger := watermill.NewStdLogger(false, false)
	pub := redis.NewRedisPublisher(redisClient, watermillLogger)

	orderRepository := postgres.NewOrderRepository(pool)
	testOrder := &domain.Order{
		ID:         "integration-order-2",
		ProductIDs: "prod-1,prod-2",
		Status:     domain.OrderStatusNew,
	}
	err := orderRepository.Create(ctx, testOrder)
	require.NoError(t, err)

	router := redis.NewWatermillRouter(redisClient, orderRepository, watermillLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerError := make(chan error, 1)
	go func() {
		consumerError <- router.Run(ctx)
	}()

	//give time for consumer to start
	time.Sleep(2 * time.Second)

	//check if consumer failed to start
	select {
	case err := <-consumerError:
		t.Fatalf("Consumer failed to start: %v", err)
	default:
		//consumer is running, continue with test
	}

	t.Run("order completed", func(t *testing.T) {
		orderCompleted := domain.OrderCompleted{
			OrderID: "integration-order-2",
		}

		payload, err := json.Marshal(orderCompleted)
		require.NoError(t, err)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = pub.Publish("order-completed", msg)
		require.NoError(t, err)

		//wait for message to get processed
		time.Sleep(2 * time.Second)

		select {
		case err := <-consumerError:
			t.Fatalf("Consumer failed during processing: %v", err)
		default:
		}

		var status domain.OrderStatus
		query := "SELECT status FROM orders WHERE id = $1"
		err = pool.QueryRow(ctx, query, orderCompleted.OrderID).Scan(&status)

		require.NoError(t, err)
		assert.Equal(t, domain.OrderStatusCompleted, status)
	})

	cancel()
	err = router.Close()
	assert.NoError(t, err)
}

func TestFullOrderFlow(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	redisClient := redis.NewRedisClient()
	defer redisClient.Close()

	watermillLogger := watermill.NewStdLogger(false, false)
	pub := redis.NewRedisPublisher(redisClient, watermillLogger)

	orderRepository := postgres.NewOrderRepository(pool)
	router := redis.NewWatermillRouter(redisClient, orderRepository, watermillLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerError := make(chan error, 1)
	go func() {
		consumerError <- router.Run(ctx)
	}()

	//give time for consumer to start
	time.Sleep(2 * time.Second)

	//check if consumer failed to start
	select {
	case err := <-consumerError:
		t.Fatalf("Consumer failed to start: %v", err)
	default:
		//consumer is running, continue with test
	}

	t.Run("full order flow processed", func(t *testing.T) {
		orderID := "full-flow-order"

		orderCreated := domain.OrderCreated{
			OrderID:    orderID,
			ProductIDs: []string{"product-1", "product-2"},
		}

		payload, err := json.Marshal(orderCreated)
		require.NoError(t, err)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = pub.Publish("order-created", msg)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		//check for error after creation
		select {
		case err := <-consumerError:
			t.Fatalf("Consumer failed after order creation: %v", err)
		default:
		}

		orderCompleted := domain.OrderCompleted{
			OrderID: orderID,
		}

		payload, err = json.Marshal(orderCompleted)
		require.NoError(t, err)

		msg = message.NewMessage(watermill.NewUUID(), payload)
		err = pub.Publish("order-completed", msg)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		//check for error after creation
		select {
		case err := <-consumerError:
			t.Fatalf("Consumer failed after order completion: %v", err)
		default:
		}

		var order domain.Order
		query := "SELECT id, product_ids, status FROM orders WHERE id = $1"
		err = pool.QueryRow(context.Background(), query, orderCompleted.OrderID).Scan(
			&order.ID,
			&order.ProductIDs,
			&order.Status,
		)

		require.NoError(t, err)
		assert.Equal(t, orderID, order.ID)
		assert.Equal(t, "product-1,product-2", order.ProductIDs)
		assert.Equal(t, domain.OrderStatusCompleted, order.Status)
	})

	cancel()
	err := router.Close()
	assert.NoError(t, err)
}
