package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/api/book"
	"github.com/zura-t/bookstore_fiber/api/user"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
)

func NewRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	token, err := token.NewJwtMaker(log, config.TokenKey)
	if err != nil {
		log.WithFields(logrus.Fields{
			"level": "Fatal",
		}).Fatal(err)
	}

	{
		user.NewuserRouter(app, log, config, db, token)
		book.NewBookRouter(app, log, config, db, token)
	}
}
