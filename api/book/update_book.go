package book

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm/clause"
)

func (r *bookRouter) UpdateBook(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	var book *BookUpdate
	if err := c.BodyParser(&book); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(book); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			validation_errs := pkg.ListValidationErrors(book, validationErrors)
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

	file, err := c.FormFile("book")
	if err != nil {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	path := "public/uploads/" + file.Filename

	c.SaveFile(file, path)

	book.File = path

	var res models.Book
	err = r.db.Model(&res).Clauses(clause.Returning{}).Where(&models.Book{ID: book.Id, AuthorID: data.UserId}).Updates(&book).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.JSON(res)
}

type BookUpdate struct {
	Id          uint   `json:"id" validate:"required,min=1"`
	Title       string `json:"title" validate:"min=1"`
	Description string `json:"description"`
	Price       uint   `json:"price" validate:"min=1"`
	File        string `json:"file"`
}
