package book

import (
	"fmt"
	"mime/multipart"
	"time"

	// "github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
)

func (r *bookRouter) UploadBook(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var book = &UploadBook{}
	if err := c.BodyParser(book); err != nil {
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
		err := fmt.Errorf("Can't get file")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}
	if file == nil {
		err := fmt.Errorf("You didn't attach the file")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	path := "public/uploads/" + file.Filename

	c.SaveFile(file, path)

	arg := models.Book{
		Title:       book.Title,
		Description: book.Description,
		AuthorID:    data.UserId,
		Price:       book.Price,
		File:        path,
	}

	err = r.db.Create(&arg).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := convertBook(arg)

	return c.JSON(res)
}

type UploadBook struct {
	Title       string                `form:"title" validate:"required,min=1"`
	Description string                `form:"description" validate:"required,min=1"`
	Price       uint                  `form:"price" validate:"required,min=1"`
	Book        *multipart.FileHeader `form:"book"`
}

type UploadBookResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       uint      `json:"price"`
	AuthorID    uint      `json:"author_id"`
	File        string    `json:"file"`
}

func convertBook(book models.Book) *UploadBookResponse {
	return &UploadBookResponse{
		ID:          book.ID,
		Title:       book.Title,
		Description: book.Description,
		Price:       book.Price,
		AuthorID:    book.AuthorID,
		File:        book.File,
		CreatedAt:   book.CreatedAt,
		UpdatedAt:   book.UpdatedAt,
	}
}
