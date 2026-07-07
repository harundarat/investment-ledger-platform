package domain

import "github.com/google/uuid"

type Holding struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Units       int64     `json:"units"`
	AvgBuyPrice int64     `json:"avg_buy_price"`
}
