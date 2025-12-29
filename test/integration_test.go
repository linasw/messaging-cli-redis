package test

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"messaging-cli/internal/domain"
	"messaging-cli/internal/repository/postgres"
	"messaging-cli/testhelpers"
	"testing"
	"time"
)

func TestOrderCreatedFlow(t *testing.T) {
	pool := testhelpers.SetupTestDB(t)
	defer pool.Close()

	orderRepository := postgres.NewOrderRepository(pool)

	consumer := testhelpers.SetupConsumer(t, orderRepository)
	defer testhelpers.CleanupConsumer(t, consumer)

	t.Run("order created", func(t *testing.T) {
		orderCreated := domain.OrderCreated{
			OrderID:    "integration-order-1",
			ProductIDs: []string{"product-1", "product-2"},
		}

		payload, err := json.Marshal(orderCreated)
		require.NoError(t, err)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = consumer.Publisher.Publish("order-created", msg)
		require.NoError(t, err)

		//wait for message to get processed
		time.Sleep(2 * time.Second)

		testhelpers.CheckConsumerError(t, consumer)

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
}

func TestOrderCompletedFlow(t *testing.T) {
	ctx := context.Background()

	pool := testhelpers.SetupTestDB(t)
	defer pool.Close()

	orderRepository := postgres.NewOrderRepository(pool)
	testOrder := &domain.Order{
		ID:         "integration-order-2",
		ProductIDs: "prod-1,prod-2",
		Status:     domain.OrderStatusNew,
	}
	err := orderRepository.Create(ctx, testOrder)
	require.NoError(t, err)

	consumer := testhelpers.SetupConsumer(t, orderRepository)
	defer testhelpers.CleanupConsumer(t, consumer)

	t.Run("order completed", func(t *testing.T) {
		orderCompleted := domain.OrderCompleted{
			OrderID: "integration-order-2",
		}

		payload, err := json.Marshal(orderCompleted)
		require.NoError(t, err)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = consumer.Publisher.Publish("order-completed", msg)
		require.NoError(t, err)

		//wait for message to get processed
		time.Sleep(2 * time.Second)

		testhelpers.CheckConsumerError(t, consumer)

		var status domain.OrderStatus
		query := "SELECT status FROM orders WHERE id = $1"
		err = pool.QueryRow(ctx, query, orderCompleted.OrderID).Scan(&status)

		require.NoError(t, err)
		assert.Equal(t, domain.OrderStatusCompleted, status)
	})
}

func TestFullOrderFlow(t *testing.T) {
	pool := testhelpers.SetupTestDB(t)
	defer pool.Close()

	orderRepository := postgres.NewOrderRepository(pool)

	consumer := testhelpers.SetupConsumer(t, orderRepository)
	defer testhelpers.CleanupConsumer(t, consumer)

	t.Run("full order flow processed", func(t *testing.T) {
		orderID := "full-flow-order"

		orderCreated := domain.OrderCreated{
			OrderID:    orderID,
			ProductIDs: []string{"product-1", "product-2"},
		}

		payload, err := json.Marshal(orderCreated)
		require.NoError(t, err)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = consumer.Publisher.Publish("order-created", msg)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		//check for error after creation
		testhelpers.CheckConsumerError(t, consumer)

		orderCompleted := domain.OrderCompleted{
			OrderID: orderID,
		}

		payload, err = json.Marshal(orderCompleted)
		require.NoError(t, err)

		msg = message.NewMessage(watermill.NewUUID(), payload)
		err = consumer.Publisher.Publish("order-completed", msg)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		//check for error after completion
		testhelpers.CheckConsumerError(t, consumer)

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
}
