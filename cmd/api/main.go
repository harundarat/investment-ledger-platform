package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/harundarat/investment-ledger-platform/internal/config"
)

func main() {
	_, err := config.Load()
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

	log.Fatal(app.Listen(":8000"))
}
