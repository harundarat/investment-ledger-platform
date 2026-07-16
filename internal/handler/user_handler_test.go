package handler

import (
	"context"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/harundarat/investment-ledger-platform/internal/dto"
)

type fakeUserService struct {
	registerUser  *domain.User
	registerErr   error
	profileUser   *domain.User
	profileErr    error
	registerInput dto.CreateUserInput
}

func (service *fakeUserService) Register(_ context.Context, input dto.CreateUserInput) (*domain.User, error) {
	service.registerInput = input
	return service.registerUser, service.registerErr
}

func (service *fakeUserService) GetProfile(_ context.Context, _ uuid.UUID) (*domain.User, error) {
	return service.profileUser, service.profileErr
}

func newUserTestApp(service *fakeUserService) *fiber.App {
	app := fiber.New()
	NewUserHandler(service, validator.New()).RegisterRoutes(app)
	return app
}

func TestUserHandlerCreate(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	service := &fakeUserService{registerUser: &domain.User{ID: id, Name: "Harun", Email: "harun@example.com"}}
	app := newUserTestApp(service)

	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"harun@example.com","password":"password-yang-kuat"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "register-harun-001")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	if service.registerInput.IdempotencyKey != "register-harun-001" {
		t.Fatalf("idempotency key = %q, want header value", service.registerInput.IdempotencyKey)
	}
}

func TestUserHandlerCreateRejectsMissingIdempotencyKey(t *testing.T) {
	app := newUserTestApp(&fakeUserService{})
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"harun@example.com","password":"password-yang-kuat"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func TestUserHandlerCreateRejectsInvalidRequest(t *testing.T) {
	app := newUserTestApp(&fakeUserService{})
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"not-an-email","password":"short"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "register-harun-001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusUnprocessableEntity)
	}
}

func TestUserHandlerGetProfile(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	service := &fakeUserService{profileUser: &domain.User{ID: id, Name: "Harun", Email: "harun@example.com"}}
	app := newUserTestApp(service)

	request := httptest.NewRequest(http.MethodGet, "/users/"+id.String(), nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
}

func TestUserHandlerGetProfileMapsNotFound(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	app := newUserTestApp(&fakeUserService{profileErr: domain.ErrUserNotFound})

	request := httptest.NewRequest(http.MethodGet, "/users/"+id.String(), nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestUserHandlerCreateMapsEmailConflict(t *testing.T) {
	app := newUserTestApp(&fakeUserService{registerErr: domain.ErrEmailAlreadyExists})
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"harun@example.com","password":"password-yang-kuat"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "register-harun-001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusConflict)
	}
}
