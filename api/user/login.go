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
	"time"
)

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginUserResponse struct {
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  UserResponse `json:"user"`
}

func (r *userRouter) Login(c *fiber.Ctx) error {
	var req = &LoginUserRequest{}
	if err := c.BodyParser(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusServiceUnavailable).JSON(pkg.ErrorResponse(err))
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

	err = pkg.CheckPassword(req.Password, user.Password)
	if err != nil {
		err = fmt.Errorf("Error incorrect password, %s", err)
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	accessToken, accessPayload, err := r.token.CreateToken(user.ID, user.Email, r.config.AccessTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %s", err)
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	refreshToken, refreshPayload, err := r.token.CreateToken(user.ID, user.Email, r.config.RefreshTokenDuration)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %s", err)
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	res := LoginUserResponse{
		User:                  ConvertUser(user),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}

	return c.JSON(res)
}
