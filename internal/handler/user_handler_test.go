package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/harundarat/investment-ledger-platform/internal/dto"
)

type fakeUserService struct {
	registerUser   *domain.User
	registerErr    error
	profileUser    *domain.User
	profileErr     error
	registerInput  dto.CreateUserInput
	registerCalled bool
}

func (service *fakeUserService) Register(_ context.Context, input dto.CreateUserInput) (*domain.User, error) {
	service.registerCalled = true
	service.registerInput = input
	return service.registerUser, service.registerErr
}

func (service *fakeUserService) GetProfile(_ context.Context, _ uuid.UUID) (*domain.User, error) {
	return service.profileUser, service.profileErr
}

func newUserTestApp(service *fakeUserService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	NewUserHandler(service, validator.New()).RegisterRoutes(app)
	return app
}

func assertErrorResponse(t *testing.T, response *http.Response, wantStatus int, want dto.ErrorResponse) {
	t.Helper()
	defer response.Body.Close()

	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", response.StatusCode, wantStatus)
	}
	if contentType := response.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var envelope dto.Envelope
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if envelope.Error == nil {
		t.Fatal("response error is nil")
	}
	if !reflect.DeepEqual(*envelope.Error, want) {
		t.Fatalf("error response = %#v, want %#v", *envelope.Error, want)
	}
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
	service := &fakeUserService{}
	app := newUserTestApp(service)
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"harun@example.com","password":"password-yang-kuat"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	assertErrorResponse(t, response, http.StatusBadRequest, dto.ErrorResponse{
		Code:    "IDEMPOTENCY_KEY_REQUIRED",
		Message: "Idempotency-Key header is required",
	})
	if service.registerCalled {
		t.Fatal("Register() was called for a request without an idempotency key")
	}
}

func TestUserHandlerCreateRejectsMalformedJSON(t *testing.T) {
	service := &fakeUserService{}
	app := newUserTestApp(service)
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","harun@example.com","password":"password-yang-kuat"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "register-harun-001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	assertErrorResponse(t, response, http.StatusBadRequest, dto.ErrorResponse{
		Code:    "MALFORMED_JSON",
		Message: "request body must contain valid JSON",
	})
	if service.registerCalled {
		t.Fatal("Register() was called for malformed JSON")
	}
}

func TestUserHandlerCreateReturnsValidationDetails(t *testing.T) {
	service := &fakeUserService{}
	app := newUserTestApp(service)
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"not-an-email","password":"short"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "register-harun-001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	assertErrorResponse(t, response, fiber.StatusUnprocessableEntity, dto.ErrorResponse{
		Code:    "VALIDATION_FAILED",
		Message: "request validation failed",
		Details: []dto.ErrorDetail{
			{Field: "email", Rule: "email", Message: "email must be a valid email address"},
			{Field: "password", Rule: "min", Message: "password must be at least 8 characters"},
		},
	})
	if service.registerCalled {
		t.Fatal("Register() was called for an invalid request")
	}
}

func TestUserHandlerCreateRejectsLongIdempotencyKey(t *testing.T) {
	service := &fakeUserService{}
	app := newUserTestApp(service)
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"harun@example.com","password":"password-yang-kuat"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", strings.Repeat("a", 256))

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	assertErrorResponse(t, response, fiber.StatusUnprocessableEntity, dto.ErrorResponse{
		Code:    "IDEMPOTENCY_KEY_TOO_LONG",
		Message: "Idempotency-Key must be at most 255 characters",
	})
	if service.registerCalled {
		t.Fatal("Register() was called for a long idempotency key")
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
	assertErrorResponse(t, response, http.StatusNotFound, dto.ErrorResponse{
		Code:    "USER_NOT_FOUND",
		Message: "user not found",
	})
}

func TestUserHandlerGetProfileRejectsInvalidID(t *testing.T) {
	app := newUserTestApp(&fakeUserService{})
	request := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	assertErrorResponse(t, response, http.StatusBadRequest, dto.ErrorResponse{
		Code:    "INVALID_USER_ID",
		Message: "user id must be a valid UUID",
	})
}

func TestUserHandlerCreateMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantError  dto.ErrorResponse
	}{
		{
			name:       "email conflict",
			serviceErr: domain.ErrEmailAlreadyExists,
			wantStatus: http.StatusConflict,
			wantError: dto.ErrorResponse{
				Code:    "EMAIL_ALREADY_REGISTERED",
				Message: "email already registered",
			},
		},
		{
			name:       "idempotency key required",
			serviceErr: domain.ErrIdempotencyKeyRequired,
			wantStatus: http.StatusBadRequest,
			wantError: dto.ErrorResponse{
				Code:    "IDEMPOTENCY_KEY_REQUIRED",
				Message: "Idempotency-Key header is required",
			},
		},
		{
			name:       "idempotency conflict",
			serviceErr: domain.ErrIdempotencyKeyReused,
			wantStatus: http.StatusConflict,
			wantError: dto.ErrorResponse{
				Code:    "IDEMPOTENCY_KEY_REUSED",
				Message: "Idempotency-Key was reused with a different request",
			},
		},
		{
			name:       "unexpected error",
			serviceErr: errors.New("database password should not be exposed"),
			wantStatus: http.StatusInternalServerError,
			wantError: dto.ErrorResponse{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newUserTestApp(&fakeUserService{registerErr: tt.serviceErr})
			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Harun","email":"harun@example.com","password":"password-yang-kuat"}`))
			request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			request.Header.Set("Idempotency-Key", "register-harun-001")

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			assertErrorResponse(t, response, tt.wantStatus, tt.wantError)
		})
	}
}

func TestErrorHandlerMapsFiberErrors(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/known", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})
	app.Get("/rejected", func(fiber.Ctx) error {
		return fiber.NewError(http.StatusRequestEntityTooLarge, "raw middleware message")
	})
	app.Get("/internal", func(fiber.Ctx) error {
		return errors.New("sensitive internal detail")
	})

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantError  dto.ErrorResponse
	}{
		{
			name:       "route not found",
			method:     http.MethodGet,
			path:       "/missing",
			wantStatus: http.StatusNotFound,
			wantError:  dto.ErrorResponse{Code: "ROUTE_NOT_FOUND", Message: "route not found"},
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			path:       "/known",
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  dto.ErrorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"},
		},
		{
			name:       "other client error",
			method:     http.MethodGet,
			path:       "/rejected",
			wantStatus: http.StatusRequestEntityTooLarge,
			wantError:  dto.ErrorResponse{Code: "REQUEST_REJECTED", Message: "request entity too large"},
		},
		{
			name:       "internal error",
			method:     http.MethodGet,
			path:       "/internal",
			wantStatus: http.StatusInternalServerError,
			wantError:  dto.ErrorResponse{Code: "INTERNAL_SERVER_ERROR", Message: "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest(tt.method, tt.path, nil))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			assertErrorResponse(t, response, tt.wantStatus, tt.wantError)
		})
	}
}
