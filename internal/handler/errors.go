package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/harundarat/investment-ledger-platform/internal/domain"
)

func handleServiceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(fail("email already registered"))
	case errors.Is(err, domain.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fail("user not found"))
	case errors.Is(err, domain.ErrIdempotencyKeyRequired):
		return c.Status(fiber.StatusBadRequest).JSON(fail("Idempotency-Key header is required"))
	case errors.Is(err, domain.ErrIdempotencyKeyReused):
		return c.Status(fiber.StatusConflict).JSON(fail("Idempotency-Key was reused with a different request"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fail("internal server error"))
	}
}
