package auth

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
)

func New(log *logrus.Logger, token *token.JwtMaker) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		req := c.Get("Authorization")
		if req == "" {
			return c.Status(403).JSON(pkg.ErrorResponse(fmt.Errorf("Forbidden")))
		}

		splitToken := strings.Split(req, "Bearer ")
		usertoken := splitToken[1]
		payload, err := token.VerifyToken(usertoken)
		if err != nil {
			log.WithFields(logrus.Fields{
				"level": "Error",
			}).Error(err)
			return c.Status(403).JSON(pkg.ErrorResponse(err))
		}

		c.Locals("user", payload)

		return c.Next()
	}
}
