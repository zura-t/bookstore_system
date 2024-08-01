package book

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	role "github.com/zura-t/bookstore_fiber/middlewares/roles"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
)

type bookRouter struct {
	log    *logrus.Logger
	config config.Config
	db     *gorm.DB
}

func NewBookRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB, token *token.JwtMaker) {
	r := &bookRouter{log, config, db}
	app.Get("/authors", r.GetAuthors)
	app.Get("/books", r.GetBooks)
	app.Get("/books/:id", r.GetBook)

	authRoutes := app.Group("/", auth.New(log, token))
	authRoutes.Get("/readlist", r.GetReadList)
	authRoutes.Post("/readlist", r.AddBookToReadList)
	authRoutes.Delete("/readlist/:bookid", r.DeleteBookFromReadList)

	bookRoutes := authRoutes.Group("/books", role.New(log, db))
	bookRoutes.Post("/", r.UploadBook)
	bookRoutes.Patch("/", r.UpdateBook)
	bookRoutes.Delete("/:id", r.DeleteBook)
}
