package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/harundarat/investment-ledger-platform/internal/dto"
	"github.com/harundarat/investment-ledger-platform/internal/pkg/hash"
)

type fakeTransactionManager struct {
	calls int
	err   error
}

func (tm *fakeTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	tm.calls++
	if tm.err != nil {
		return tm.err
	}
	return fn(ctx)
}

type fakeUserRepository struct {
	users        map[uuid.UUID]*domain.User
	saveCalls    int
	savedAccount *domain.Account
	saveErr      error
	profileErr   error
}

func (repo *fakeUserRepository) Save(_ context.Context, user *domain.User, account *domain.Account) error {
	repo.saveCalls++
	if repo.saveErr != nil {
		return repo.saveErr
	}
	if repo.users == nil {
		repo.users = make(map[uuid.UUID]*domain.User)
	}
	repo.users[user.ID] = user
	repo.savedAccount = account
	return nil
}

func (repo *fakeUserRepository) GetProfile(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if repo.profileErr != nil {
		return nil, repo.profileErr
	}
	user, ok := repo.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (repo *fakeUserRepository) GetByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

type fakeIdempotencyRepository struct {
	registrations map[string]*UserRegistrationIdempotency
}

func (repo *fakeIdempotencyRepository) Acquire(_ context.Context, key, fingerprint string, userID uuid.UUID) (*UserRegistrationIdempotency, bool, error) {
	if repo.registrations == nil {
		repo.registrations = make(map[string]*UserRegistrationIdempotency)
	}
	if registration, ok := repo.registrations[key]; ok {
		return registration, false, nil
	}

	registration := &UserRegistrationIdempotency{RequestFingerprint: fingerprint, UserID: userID}
	repo.registrations[key] = registration
	return registration, true, nil
}

func TestUserServiceRegisterCreatesUserAndWallet(t *testing.T) {
	repo := &fakeUserRepository{}
	transactionManager := &fakeTransactionManager{}
	idempotencyRepo := &fakeIdempotencyRepository{}
	service := NewUserService(repo, transactionManager, idempotencyRepo, "test-secret")

	result, err := service.Register(context.Background(), dto.CreateUserInput{
		Name:           "  Harun  ",
		Email:          "HARUN@EXAMPLE.COM ",
		Password:       "password-yang-kuat",
		IdempotencyKey: "register-harun-001",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	user := result.User

	if user.Name != "Harun" {
		t.Fatalf("user.Name = %q, want %q", user.Name, "Harun")
	}
	if user.Email != "harun@example.com" {
		t.Fatalf("user.Email = %q, want lowercase normalized email", user.Email)
	}
	if user.ID == uuid.Nil {
		t.Fatal("user.ID must be generated")
	}
	matches, err := hash.ComparePassword("password-yang-kuat", user.PasswordHash)
	if err != nil || !matches {
		t.Fatalf("password hash does not match: matches=%v err=%v", matches, err)
	}
	if transactionManager.calls != 1 {
		t.Fatalf("transaction calls = %d, want 1", transactionManager.calls)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("Save calls = %d, want 1", repo.saveCalls)
	}
	if repo.savedAccount == nil {
		t.Fatal("wallet account was not saved")
	}
	if repo.savedAccount.Type != domain.Liability {
		t.Fatalf("wallet account type = %q, want %q", repo.savedAccount.Type, domain.Liability)
	}
	if repo.savedAccount.Code != int(domain.CodeUser) {
		t.Fatalf("wallet account code = %d, want %d", repo.savedAccount.Code, domain.CodeUser)
	}
	if repo.savedAccount.UserID == nil || *repo.savedAccount.UserID != user.ID {
		t.Fatal("wallet account must belong to the created user")
	}
	if result.IdempotencyReplayed {
		t.Fatal("first registration must not be marked as an idempotency replay")
	}
}

func TestUserServiceRegisterReplaysIdempotentRequest(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, &fakeTransactionManager{}, &fakeIdempotencyRepository{}, "test-secret")
	input := dto.CreateUserInput{
		Name:           "Harun",
		Email:          "harun@example.com",
		Password:       "password-yang-kuat",
		IdempotencyKey: "register-harun-001",
	}

	firstResult, err := service.Register(context.Background(), input)
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	secondResult, err := service.Register(context.Background(), input)
	if err != nil {
		t.Fatalf("second Register() error = %v", err)
	}

	if firstResult.IdempotencyReplayed {
		t.Fatal("first registration must not be marked as an idempotency replay")
	}
	if !secondResult.IdempotencyReplayed {
		t.Fatal("second registration must be marked as an idempotency replay")
	}
	if firstResult.User.ID != secondResult.User.ID {
		t.Fatalf("idempotent retry returned user ID %s, want %s", secondResult.User.ID, firstResult.User.ID)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("Save calls = %d, want 1 after retry", repo.saveCalls)
	}
}

func TestUserServiceRegisterRejectsReusedKeyWithDifferentRequest(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, &fakeTransactionManager{}, &fakeIdempotencyRepository{}, "test-secret")
	input := dto.CreateUserInput{
		Name:           "Harun",
		Email:          "harun@example.com",
		Password:       "password-yang-kuat",
		IdempotencyKey: "register-harun-001",
	}
	if _, err := service.Register(context.Background(), input); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	input.Name = "Harun D"
	_, err := service.Register(context.Background(), input)
	if !errors.Is(err, domain.ErrIdempotencyKeyReused) {
		t.Fatalf("Register() error = %v, want %v", err, domain.ErrIdempotencyKeyReused)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("Save calls = %d, want 1", repo.saveCalls)
	}
}

func TestUserServiceGetProfileDelegatesToRepository(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	expected := &domain.User{ID: id, Name: "Harun", Email: "harun@example.com"}
	repo := &fakeUserRepository{users: map[uuid.UUID]*domain.User{id: expected}}
	service := NewUserService(repo, &fakeTransactionManager{}, &fakeIdempotencyRepository{}, "test-secret")

	user, err := service.GetProfile(context.Background(), id)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if user != expected {
		t.Fatal("GetProfile() did not return repository user")
	}
}
