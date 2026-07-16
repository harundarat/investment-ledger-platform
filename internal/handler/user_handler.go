package handler

import (
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
		return c.Status(fiber.StatusBadRequest).JSON(fail("invalid request body"))
	}
	if err := h.validate.Struct(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fail("invalid request"))
	}

	idempotencyKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fail("Idempotency-Key header is required"))
	}
	if len(idempotencyKey) > 255 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fail("Idempotency-Key must be at most 255 characters"))
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
		return c.Status(fiber.StatusBadRequest).JSON(fail("invalid user id"))
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
