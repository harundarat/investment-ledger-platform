package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string

const (
	MutualFund ProductType = "mutual_fund"
	Stock      ProductType = "stock"
	Bond       ProductType = "bond"
)

type Product struct {
	ID   uuid.UUID   `json:"id"`
	Code int         `json:"code"`
	Type ProductType `json:"type"`
	Name string      `json:"name"`
}

type ProductPrices struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Price     int64     `json:"price"`
	PriceDate time.Time `json:"price_date"`
}
