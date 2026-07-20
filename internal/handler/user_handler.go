package handler

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/harundarat/investment-ledger-platform/internal/dto"
	"github.com/harundarat/investment-ledger-platform/internal/service"
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserService, validate *validator.Validate) *UserHandler {
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" {
			return field.Name
		}
		return name
	})

	return &UserHandler{
		userService: userService,
		validate:    validate,
	}
}

func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Post("/", h.Create)
	users.Get("/:id", h.GetProfile)
}

func (h *UserHandler) Create(c fiber.Ctx) error {
	var request dto.CreateUserRequest
	if err := c.Bind().Body(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fail(
			"MALFORMED_JSON",
			"request body must contain valid JSON",
		))
	}
	if err := h.validate.Struct(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fail(
			"VALIDATION_FAILED",
			"request validation failed",
			validationDetails(err)...,
		))
	}

	idempotencyKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fail(
			"IDEMPOTENCY_KEY_REQUIRED",
			"Idempotency-Key header is required",
		))
	}
	if len(idempotencyKey) > 255 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fail(
			"IDEMPOTENCY_KEY_TOO_LONG",
			"Idempotency-Key must be at most 255 characters",
		))
	}

	user, err := h.userService.Register(c.Context(), dto.CreateUserInput{
		Name:           request.Name,
		Email:          request.Email,
		Password:       request.Password,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return handleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(success(userResponse(user), "user created"))
}

func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fail(
			"INVALID_USER_ID",
			"user id must be a valid UUID",
		))
	}

	user, err := h.userService.GetProfile(c.Context(), id)
	if err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(success(userResponse(user), ""))
}

func userResponse(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
