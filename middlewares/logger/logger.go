package logger

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func New(c *fiber.Ctx) error {

		log.Print("logger middleware enabled")

		return c.Next()
}