package database

import (
	"github.com/zura-t/bookstore_fiber/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DbConn *gorm.DB
)

func Connect(config config.Config) (*gorm.DB, error) {
	// dsn := "host=localhost user=postgres password=root dbname=book_store port=5432 sslmode=disable"
	var err error
	DbConn, err = gorm.Open(postgres.Open(config.DbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return DbConn, err
}
