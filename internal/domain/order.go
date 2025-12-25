package domain

import "context"

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusCompleted OrderStatus = "completed"
)

type Order struct {
	ID         string
	ProductIDs string
	Status     OrderStatus
}

type OrderCreated struct {
	OrderID    string   `json:"order_id"`
	ProductIDs []string `json:"product_ids"`
}

type OrderCompleted struct {
	OrderID string `json:"order_id"`
}

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	Complete(ctx context.Context, orderID string) error
}
