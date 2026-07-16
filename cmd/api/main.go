package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/harundarat/investment-ledger-platform/internal/config"
	"github.com/harundarat/investment-ledger-platform/internal/handler"
	"github.com/harundarat/investment-ledger-platform/internal/repository/postgres"
	"github.com/harundarat/investment-ledger-platform/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	db, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	userRepository := postgres.NewUserRepository(db)
	idempotencyRepository := postgres.NewUserRegistrationIdempotencyRepository(db)
	transactionManager := postgres.NewTransactionManager(db)
	userService := service.NewUserService(userRepository, transactionManager, idempotencyRepository, cfg.IdempotencyHashSecret)
	userHandler := handler.NewUserHandler(userService, validator.New())

	app := fiber.New(fiber.Config{
		AppName:      "investment-ledger-platform",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})
	userHandler.RegisterRoutes(app)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Printf("server: %v", err)
	}
}
