package entities

type OrderCreated struct {
	OrderID    string   `json:"order_id"`
	ProductIDs []string `json:"product_ids"`
}

type OrderCompleted struct {
	OrderID string `json:"order_id"`
}
