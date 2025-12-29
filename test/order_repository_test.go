package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"messaging-cli/internal/domain"
	"messaging-cli/internal/repository/postgres"
	"messaging-cli/testhelpers"
	"testing"
)

func TestOrderRepository_Create(t *testing.T) {
	pool := testhelpers.SetupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	repo := postgres.NewOrderRepository(pool)

	t.Run("successfully create an order", func(t *testing.T) {
		order := &domain.Order{
			ID:         "test-order-1",
			ProductIDs: "prod1,prod2,prod3",
			Status:     domain.OrderStatusNew,
		}

		err := repo.Create(ctx, order)
		assert.NoError(t, err)

		var retrievedOrder domain.Order
		query := "SELECT id, product_ids, status FROM orders WHERE id = $1"
		err = pool.QueryRow(ctx, query, order.ID).Scan(
			&retrievedOrder.ID,
			&retrievedOrder.ProductIDs,
			&retrievedOrder.Status,
		)

		assert.NoError(t, err)
		assert.Equal(t, order.ID, retrievedOrder.ID)
		assert.Equal(t, order.ProductIDs, retrievedOrder.ProductIDs)
		assert.Equal(t, order.Status, retrievedOrder.Status)
	})

	t.Run("fail to create an duplicate order", func(t *testing.T) {
		order := &domain.Order{
			ID:         "test-duplicate-order",
			ProductIDs: "prod1,prod2,prod3",
			Status:     domain.OrderStatusNew,
		}

		err := repo.Create(ctx, order)
		assert.NoError(t, err)

		err = repo.Create(ctx, order)
		assert.Error(t, err)
	})
}

func TestOrderRepository_Complete(t *testing.T) {
	pool := testhelpers.SetupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	repo := postgres.NewOrderRepository(pool)

	t.Run("successfully complete an existing order", func(t *testing.T) {
		order := &domain.Order{
			ID:         "test-order-complete",
			ProductIDs: "prod1,prod2",
			Status:     domain.OrderStatusNew,
		}

		err := repo.Create(ctx, order)
		require.NoError(t, err)

		err = repo.Complete(ctx, order.ID)
		assert.NoError(t, err)

		var status domain.OrderStatus
		query := "SELECT status FROM orders WHERE id = $1"
		err = pool.QueryRow(ctx, query, order.ID).Scan(&status)

		assert.NoError(t, err)
		assert.Equal(t, domain.OrderStatusCompleted, status)
	})

	t.Run("completes non-existing order", func(t *testing.T) {
		err := repo.Complete(ctx, "non-existing-order")
		assert.NoError(t, err)
	})
}
