package book

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"gorm.io/gorm"
)

func (r *bookRouter) GetBook(c *fiber.Ctx) error {
	var req = &BookId{}
	if err := c.ParamsParser(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
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

	var book models.Book

	err := r.db.Preload("Author", func(tx *gorm.DB) *gorm.DB {
		return tx.Omit("users.password")
	}).First(&book, req.Id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := fmt.Errorf("Book not found")
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
	res := ConvertBook(book)
	return c.JSON(res)
}

func ConvertBook(book models.Book) BookResponse {
	return BookResponse{
		Id:          book.ID,
		Title:       book.Title,
		Description: book.Description,
		Price:       book.Price,
		AuthorID:    book.AuthorID,
		AuthorName:  book.Author.Name,
		CreatedAt:   book.CreatedAt,
		UpdatedAt:   book.UpdatedAt,
	}
}
