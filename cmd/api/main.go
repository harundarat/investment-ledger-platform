package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/harundarat/investment-ledger-platform/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("config: error loading .env file")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "investment-ledger-platform",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	log.Fatal(app.Listen(":" + cfg.Port))
}
