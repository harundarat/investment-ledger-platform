package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	Asset     AccountType = "asset"
	Liability AccountType = "liability"
	Equity    AccountType = "equity"
	Revenue   AccountType = "revenue"
	Expense   AccountType = "expense"
)

type Account struct {
	ID       uuid.UUID   `json:"id"`
	Code     int         `json:"code"`
	Name     string      `json:"name"`
	Type     AccountType `json:"type"`
	UserID   *uuid.UUID  `json:"user_id"`
	Currency string      `json:"currency"`
}

type JournalEntryType string

const (
	EntryTopup      JournalEntryType = "topup"
	EntryBuy        JournalEntryType = "buy"
	EntryWithdrawal JournalEntryType = "withdrawal"
	EntrySell       JournalEntryType = "sell"
	EntryFee        JournalEntryType = "fee"
)

type JournalEntry struct {
	ID             uuid.UUID        `json:"id"`
	EntryType      JournalEntryType `json:"entry_type"`
	Description    *string          `json:"description"`
	IdempotencyKey string           `json:"idempotency_key"`
	OccuredAt      time.Time        `json:"occured_at"`
}

type JournalLineDirection string

const (
	Debit  JournalLineDirection = "debit"
	Credit JournalLineDirection = "credit"
)

type JournalLine struct {
	ID             uuid.UUID            `json:"id"`
	JournalEntryID uuid.UUID            `json:"journal_entry_id"`
	AccountID      uuid.UUID            `json:"account_id"`
	Direction      JournalLineDirection `json:"direction"`
	Amount         int64                `json:"amount"`
}
