package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
	"github.com/harundarat/investment-ledger-platform/internal/dto"
)

func handleServiceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(fail(
			"EMAIL_ALREADY_REGISTERED",
			"email already registered",
		))
	case errors.Is(err, domain.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fail(
			"USER_NOT_FOUND",
			"user not found",
		))
	case errors.Is(err, domain.ErrIdempotencyKeyRequired):
		return c.Status(fiber.StatusBadRequest).JSON(fail(
			"IDEMPOTENCY_KEY_REQUIRED",
			"Idempotency-Key header is required",
		))
	case errors.Is(err, domain.ErrIdempotencyKeyReused):
		return c.Status(fiber.StatusConflict).JSON(fail(
			"IDEMPOTENCY_KEY_REUSED",
			"Idempotency-Key was reused with a different request",
		))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fail(
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		))
	}
}

func validationDetails(err error) []dto.ErrorDetail {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	details := make([]dto.ErrorDetail, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		rule := fieldError.Tag()

		var message string
		switch rule {
		case "required":
			message = fmt.Sprintf("%s is required", field)
		case "email":
			message = fmt.Sprintf("%s must be a valid email address", field)
		case "min":
			message = fmt.Sprintf("%s must be at least %s characters", field, fieldError.Param())
		case "max":
			message = fmt.Sprintf("%s must be at most %s characters", field, fieldError.Param())
		default:
			message = fmt.Sprintf("%s is invalid", field)
		}

		details = append(details, dto.ErrorDetail{
			Field:   field,
			Rule:    rule,
			Message: message,
		})
	}

	return details
}

// ErrorHandler keeps errors returned by Fiber and middleware consistent with
// errors returned directly by the application's HTTP handlers.
func ErrorHandler(c fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	var fiberError *fiber.Error
	if errors.As(err, &fiberError) {
		status = fiberError.Code
	}

	code := "INTERNAL_SERVER_ERROR"
	message := "internal server error"

	switch status {
	case fiber.StatusNotFound:
		code = "ROUTE_NOT_FOUND"
		message = "route not found"
	case fiber.StatusMethodNotAllowed:
		code = "METHOD_NOT_ALLOWED"
		message = "method not allowed"
	default:
		if status >= 400 && status < 500 {
			code = "REQUEST_REJECTED"
			message = strings.ToLower(http.StatusText(status))
			if message == "" {
				message = "request rejected"
			}
		} else {
			status = fiber.StatusInternalServerError
		}
	}

	return c.Status(status).JSON(fail(code, message))
}
