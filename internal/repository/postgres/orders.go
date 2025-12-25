package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"messaging-cli/internal/domain"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, product_ids, status) 
		VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, order.ID, order.ProductIDs, order.Status)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) Complete(ctx context.Context, orderID string) error {
	query := `
		UPDATE orders 
		SET status = $1 
		WHERE id = $2`

	_, err := r.pool.Exec(ctx, query, domain.OrderStatusCompleted, orderID)
	if err != nil {
		return err
	}

	return nil
}
