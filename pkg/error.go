package pkg

import "github.com/gofiber/fiber/v2"

func ErrorResponse(err error) fiber.Map {
	return fiber.Map{"error": err.Error()}
}
