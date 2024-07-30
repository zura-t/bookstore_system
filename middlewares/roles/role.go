package role

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
)

func New(log *logrus.Logger, db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
	payload := c.Locals("user")
		data, ok := payload.(*token.Payload)
		if !ok {
			err := fmt.Errorf("Forbidden")
			log.WithFields(logrus.Fields{
				"level": "Error",
			}).Error(err)
			return c.Status(fiber.StatusForbidden).JSON(pkg.ErrorResponse(err))
		}

		var user models.User

		err := db.First(&user, data.UserId).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = fmt.Errorf("User not found")
				log.WithFields(logrus.Fields{
					"level": "Error",
				}).Error(err)
				return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
			}

			log.WithFields(logrus.Fields{
				"level": "Error",
			}).Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
		}

		if !user.IsAuthor {
			err := fmt.Errorf("Forbidden")
			log.WithFields(logrus.Fields{
				"level": "Error",
			}).Error(err)
			return c.Status(fiber.StatusForbidden).JSON(pkg.ErrorResponse(err))
		}

		return c.Next()
	}
}
