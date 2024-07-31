package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
)

type UserId struct {
	Id uint `uri:"id" json:"id" validate:"required,min=1"`
}

func (r *userRouter) GetUsers(c *fiber.Ctx) error {
	var users []models.User
	err := r.db.Find(&users).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := make([]*UserResponse, len(users))
	for index, v := range users {
		user := ConvertUser(v)
		res[index] = &user
	}
	return c.JSON(res)
}
