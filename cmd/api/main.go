package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func main() {
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
