package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	"github.com/zura-t/bookstore_fiber/token"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type userRouter struct {
	log    *logrus.Logger
	config config.Config
	db     *gorm.DB
	token  *token.JwtMaker
}

func NewuserRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB, token *token.JwtMaker) {
	r := &userRouter{log, config, db, token}

	app.Post("/register", r.Register)
	app.Post("/login", r.Login)
	app.Post("/renew_token", r.RenewAccessToken)
	app.Post("/logout", r.Logout)

	authRoutes := app.Group("/", auth.New(log, token))

	authRoutes.Get("/users", r.GetUsers)
	authRoutes.Get("/users/my_profile", r.GetMyProfile)
	authRoutes.Get("/users/:id", r.GetUser)
	authRoutes.Patch("/users/my_profile", r.UpdateMyProfile)
	authRoutes.Delete("/users/my_profile", r.DeleteMyProfile)
	authRoutes.Patch("/users/author", r.BecomeAuthor)
}
