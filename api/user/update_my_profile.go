package user

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm/clause"
)

type UserUpdate struct {
	Name string `json:"name" validate:"required,min=1"`
}

func (r *userRouter) UpdateMyProfile(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user = &UserUpdate{}
	if err := c.BodyParser(user); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusServiceUnavailable).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			validation_errs := pkg.ListValidationErrors(user, validationErrors)
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

	var res models.User
	err := r.db.Model(&res).Clauses(clause.Returning{}).Where("id = ?", data.UserId).Updates(&user).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	resp := ConvertUser(res)
	return c.JSON(resp)
}
