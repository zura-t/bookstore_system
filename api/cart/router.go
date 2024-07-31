package cart

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
)

type cartRouter struct {
	log    *logrus.Logger
	config config.Config
	db     *gorm.DB
}

func NewCartRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB, token *token.JwtMaker) {
	r := &cartRouter{log, config, db}
	authRoutes := app.Group("/", auth.New(log, token))

	cartRoutes := authRoutes.Group("/cart")
	cartRoutes.Post("/", r.AddBookToCart)
	cartRoutes.Get("/", r.GetBooksInCart)
	cartRoutes.Delete("/:id", r.DeleteBookFromCart)
	cartRoutes.Delete("/", r.DeleteAllBooksFromCart)
}
