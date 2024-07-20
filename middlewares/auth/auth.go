package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zura-t/bookstore_fiber/token"
)

func New(c *fiber.Ctx) error {
	req := c.Get("Authorization")
	if req == "" {
		return c.Status(403).SendString("Forbidden")
	}

	splitToken := strings.Split(req, "Bearer ")
	usertoken := splitToken[1]
	payload, err := token.Jwtmaker.VerifyToken(usertoken)
	if err != nil {
		return c.Status(403).SendString(err.Error())
	}

	c.Locals("user", payload)

	return c.Next()
}
