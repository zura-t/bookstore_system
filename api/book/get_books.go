package book

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"gorm.io/gorm"
	"time"
)

type BookResponse struct {
	Id          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       uint      `json:"price"`
	AuthorID    uint      `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BookId struct {
	Id uint `uri:"id" json:"id" validate:"required,min=1"`
}

func (r *bookRouter) GetBooks(c *fiber.Ctx) error {
	var books []models.Book
	err := r.db.Preload("Author", func(tx *gorm.DB) *gorm.DB {
		return tx.Omit("users.password")
	}).Find(&books).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := make([]*BookResponse, len(books))
	for index, v := range books {
		book := ConvertBook(v)
		res[index] = &book
	}
	return c.JSON(res)
}
