package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) Save(ctx context.Context, user *domain.User, account *domain.Account) error {
	const queryInsertUser = `INSERT INTO users(id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	const queryInsertAccount = `INSERT INTO accounts(id, code, name, type, user_id, currency) VALUES($1, $2, $3, $4, $5, $6)`

	db := databaseFromContext(ctx, ur.db)
	if _, err := db.Exec(ctx, queryInsertUser, user.ID, user.Name, user.Email, user.PasswordHash); err != nil {
		if isEmailUniqueViolation(err) {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	if _, err := db.Exec(ctx, queryInsertAccount, account.ID, account.Code, account.Name, account.Type, account.UserID, account.Currency); err != nil {
		return fmt.Errorf("create user wallet account: %w", err)
	}

	return nil
}

func (ur *userRepository) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `SELECT id, name, email FROM users WHERE id = $1`

	user := new(domain.User)
	err := databaseFromContext(ctx, ur.db).QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get profile: %w", err)
	}

	return user, nil
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `SELECT id, name, email, password_hash FROM users WHERE email = $1`

	user := new(domain.User)
	err := databaseFromContext(ctx, ur.db).QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get by email: %w", err)
	}

	return user, nil
}

func isEmailUniqueViolation(err error) bool {
	var postgresErr *pgconn.PgError
	return errors.As(err, &postgresErr) && postgresErr.Code == "23505" && postgresErr.ConstraintName == "users_email_key"
}

var _ domain.UserRepository = (*userRepository)(nil)
