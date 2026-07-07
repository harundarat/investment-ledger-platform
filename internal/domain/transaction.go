package domain

import "github.com/google/uuid"

type OrderSide string

const (
	SideBuy  OrderSide = "buy"
	SideSell OrderSide = "sell"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusSettled OrderStatus = "settled"
	OrderStatusFailed  OrderStatus = "failed"
)

type Order struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	ProductID      uuid.UUID   `json:"product_id"`
	Side           OrderSide   `json:"side"`
	AmountIDR      int64       `json:"amount_idr"`
	Units          int64       `json:"units"`
	PriceUsed      int64       `json:"price_used"`
	Status         OrderStatus `json:"status"`
	JournalEntryID uuid.UUID   `json:"journal_entry_id"`
}

type CashTransactionDirection string

const (
	DirectionIn  CashTransactionDirection = "in"
	DirectionOut CashTransactionDirection = "out"
)

type CashTransactionStatus string

const (
	CashTransactionPending CashTransactionStatus = "pending"
	CashTransactionSuccess CashTransactionStatus = "success"
	CashTransactionFailed  CashTransactionStatus = "failed"
)

type CashTransaction struct {
	ID             uuid.UUID                `json:"id"`
	UserID         uuid.UUID                `json:"user_id"`
	Direction      CashTransactionDirection `json:"direction"`
	Amount         int64                    `json:"amount"`
	Status         CashTransactionStatus    `json:"status"`
	BankReference  string                   `json:"bank_reference"`
	JournalEntryID uuid.UUID                `json:"journal_entry_id"`
}
