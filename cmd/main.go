package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/api"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/database"
	"github.com/zura-t/bookstore_fiber/logger"
	"github.com/zura-t/bookstore_fiber/models"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		fmt.Errorf("can't load config file:", err)
	}
	app := fiber.New()
	log := logger.SetupLogger(config.Environment)

	db, err := database.Connect(config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"level": "Panic",
		}).Panic(err)
		panic("failed to connect database")
	}

	log.WithFields(logrus.Fields{
		"level": "Info",
	}).Info("Connection opened to database")

	db.AutoMigrate(&models.User{}, &models.Book{}, &models.UserBook{}, &models.CartItem{})

	log.WithFields(logrus.Fields{
		"level": "Info",
	}).Info("Migrated")

	// app.Use(middleware.Logger())

	app.Use(cors.New())

	api.NewRouter(app, log, config, db)

	app.Listen("127.0.0.1:8080")
}
