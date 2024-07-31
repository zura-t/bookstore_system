package user

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"gorm.io/gorm"
)

func (r *userRouter) GetUser(c *fiber.Ctx) error {
	var req = &UserId{}
	if err := c.ParamsParser(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			validation_errs := pkg.ListValidationErrors(req, validationErrors)
			r.log.WithFields(logrus.Fields{
				"level": "Error",
			}).Error(validation_errs)
			return c.Status(fiber.StatusBadRequest).JSON(pkg.MultipleErrorsResponse(validation_errs))
		}
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user models.User

	err := r.db.First(&user, req.Id).Error
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
