package cart

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

func (r *cartRouter) DeleteAllBooksFromCart(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	err := r.db.Delete(models.CartItem{}, models.CartItem{UserID: data.UserId}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := fmt.Errorf("CartItem not found")
			r.log.WithFields(logrus.Fields{
				"level": "Error",
			}).Error(err)
			return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
		}

		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.SendString("CartItems deleted")
}
