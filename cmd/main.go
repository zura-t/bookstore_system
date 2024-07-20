package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zura-t/bookstore_fiber/api/user"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/database"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/token"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/register", user.Register)
	app.Post("/login", user.Login)
	app.Post("/renew_token", user.RenewAccessToken)
	app.Post("/logout", user.Logout)

	authRoutes := app.Group("/", auth.New)
	authRoutes.Get("/users", user.GetUsers)
	authRoutes.Get("/users/my_profile", user.GetMyProfile)
	authRoutes.Get("/users/:id", user.GetUser)
	authRoutes.Patch("/users/my_profile", user.UpdateMyProfile)
	authRoutes.Delete("/users/my_profile", user.DeleteMyProfile)
}

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("can't load config file:", err)
	}
	app := fiber.New()

	db, err := database.Connect(config)
	if err != nil {
		panic("failed to connect database")
	}
	fmt.Println("Connection opened to database")
	db.AutoMigrate(&models.User{})
	fmt.Println("Migrated")

	// app.Use(middleware.Logger())
	err = token.NewJwtMaker(config.TokenKey)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	app.Use(cors.New())

	setupRoutes(app)

	app.Listen("127.0.0.1:8080")
}
