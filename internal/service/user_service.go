package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/harundarat/investment-ledger-platform/internal/dto"
	"github.com/harundarat/investment-ledger-platform/internal/pkg/hash"
)

type UserService interface {
	Register(ctx context.Context, userInput dto.CreateUserInput) (*UserRegistrationResult, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type UserRegistrationResult struct {
	User                *domain.User
	IdempotencyReplayed bool
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
}

type UserRegistrationIdempotency struct {
	RequestFingerprint string
	UserID             uuid.UUID
}

type UserRegistrationIdempotencyRepository interface {
	Acquire(ctx context.Context, key, requestFingerprint string, userID uuid.UUID) (*UserRegistrationIdempotency, bool, error)
}

type userService struct {
	userRepository        domain.UserRepository
	transactionManager    TransactionManager
	idempotencyRepository UserRegistrationIdempotencyRepository
	idempotencySecret     []byte
}

func NewUserService(
	repo domain.UserRepository,
	transactionManager TransactionManager,
	idempotencyRepository UserRegistrationIdempotencyRepository,
	idempotencySecret string,
) UserService {
	return &userService{
		userRepository:        repo,
		transactionManager:    transactionManager,
		idempotencyRepository: idempotencyRepository,
		idempotencySecret:     []byte(idempotencySecret),
	}
}

func (us *userService) Register(ctx context.Context, userInput dto.CreateUserInput) (*UserRegistrationResult, error) {
	idempotencyKey := strings.TrimSpace(userInput.IdempotencyKey)
	if idempotencyKey == "" {
		return nil, domain.ErrIdempotencyKeyRequired
	}

	newUserID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}
	newAccountID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate account id: %w", err)
	}

	name := strings.TrimSpace(userInput.Name)
	email := strings.ToLower(strings.TrimSpace(userInput.Email))
	requestFingerprint := registrationFingerprint(us.idempotencySecret, name, email, userInput.Password)

	passwordHash, err := hash.HashPassword(userInput.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	newUser := &domain.User{
		ID:           newUserID,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}
	newUserAccount := &domain.Account{
		ID:       newAccountID,
		Code:     int(domain.CodeUser),
		Name:     string(domain.UserWallet),
		Type:     domain.Liability,
		UserID:   &newUserID,
		Currency: string(domain.IDR),
	}

	result := &UserRegistrationResult{}
	err = us.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		registration, acquired, err := us.idempotencyRepository.Acquire(txCtx, idempotencyKey, requestFingerprint, newUserID)
		if err != nil {
			return err
		}
		if !acquired {
			if !hmac.Equal([]byte(registration.RequestFingerprint), []byte(requestFingerprint)) {
				return domain.ErrIdempotencyKeyReused
			}

			result.User, err = us.userRepository.GetProfile(txCtx, registration.UserID)
			result.IdempotencyReplayed = true
			return err
		}

		if err := us.userRepository.Save(txCtx, newUser, newUserAccount); err != nil {
			return err
		}

		result.User = newUser
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (us *userService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return us.userRepository.GetProfile(ctx, id)
}

func registrationFingerprint(secret []byte, name, email, password string) string {
	mac := hmac.New(sha256.New, secret)
	for _, value := range []string{name, email, password} {
		_, _ = fmt.Fprintf(mac, "%d:%s", len(value), value)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

var _ UserService = (*userService)(nil)
