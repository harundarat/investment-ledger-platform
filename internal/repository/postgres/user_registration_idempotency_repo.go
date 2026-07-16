package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRegistrationIdempotencyRepository struct {
	db *pgxpool.Pool
}

func NewUserRegistrationIdempotencyRepository(db *pgxpool.Pool) service.UserRegistrationIdempotencyRepository {
	return &userRegistrationIdempotencyRepository{db: db}
}

func (ir *userRegistrationIdempotencyRepository) Acquire(
	ctx context.Context,
	key, requestFingerprint string,
	userID uuid.UUID,
) (*service.UserRegistrationIdempotency, bool, error) {
	const insertQuery = `
		INSERT INTO user_registration_idempotency_keys (idempotency_key, request_fingerprint, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING request_fingerprint, user_id`
	const selectQuery = `
		SELECT request_fingerprint, user_id
		FROM user_registration_idempotency_keys
		WHERE idempotency_key = $1`

	registration := new(service.UserRegistrationIdempotency)
	db := databaseFromContext(ctx, ir.db)
	err := db.QueryRow(ctx, insertQuery, key, requestFingerprint, userID).Scan(&registration.RequestFingerprint, &registration.UserID)
	if err == nil {
		return registration, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, fmt.Errorf("acquire registration idempotency key: %w", err)
	}

	if err := db.QueryRow(ctx, selectQuery, key).Scan(&registration.RequestFingerprint, &registration.UserID); err != nil {
		return nil, false, fmt.Errorf("get registration idempotency key: %w", err)
	}

	return registration, false, nil
}

var _ service.UserRegistrationIdempotencyRepository = (*userRegistrationIdempotencyRepository)(nil)
