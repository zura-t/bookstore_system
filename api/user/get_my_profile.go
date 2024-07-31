package user

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

func (r *userRouter) GetMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User

	err := r.db.First(&user, data.UserId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("User not found")
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

	res := ConvertUser(user)
	return c.JSON(res)
}
