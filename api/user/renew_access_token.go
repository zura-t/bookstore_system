package user

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/pkg"
	"time"
)

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (r *userRouter) RenewAccessToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		err := fmt.Errorf("can't renew the token")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusUnauthorized).JSON(pkg.ErrorResponse(err))
	}

	refreshPayload, err := r.token.VerifyToken(refreshToken)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusUnauthorized).JSON(pkg.ErrorResponse(err))
	}

	accessToken, accessPayload, err := r.token.CreateToken(refreshPayload.UserId, refreshPayload.Email, r.config.AccessTokenDuration)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	return c.JSON(rsp)
}
