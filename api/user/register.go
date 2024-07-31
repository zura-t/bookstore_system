package user

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"gorm.io/gorm"
	"time"
)

type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserResponse struct {
	Id        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsAuthor  bool      `json:"is_author"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

func (r *userRouter) Register(c *fiber.Ctx) error {
	var req *RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(503).JSON(pkg.ErrorResponse(err))
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
	err := r.db.First(&user, models.User{Email: req.Email}).Error
	if err == nil {
		err = fmt.Errorf("User with this email already exists")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	hashedPassword, err := pkg.HashPassword(req.Password)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	new_user := models.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: hashedPassword,
	}

	err = r.db.Create(&new_user).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	res := ConvertUser(new_user)

	return c.JSON(res)
}

func ConvertUser(user models.User) UserResponse {
	return UserResponse{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAuthor:  user.IsAuthor,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
