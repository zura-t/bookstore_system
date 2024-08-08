package book

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GetBooks struct {
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset" validate:"min=0"`
	Title     string `form:"title"`
	AuthorId  int    `form:"author_id" validate:"min=0"`
	OrderDesc bool   `form:"order_desc"`
}

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
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	authorId := c.QueryInt("author_id")
	title := c.Query("title")
	orderDesc := c.QueryBool("order_desc")

	req := &GetBooks{
		Limit:     limit,
		Offset:    offset,
		Title:     title,
		AuthorId:  authorId,
		OrderDesc: orderDesc,
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			var validation_errs = pkg.ListValidationErrors(req, validationErrors)
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

	var books []models.Book
	err := r.db.Preload("Author", func(tx *gorm.DB) *gorm.DB {
		return tx.Omit("users.password")
	}).Where(&models.Book{AuthorID: uint(authorId), Title: title}).Order(clause.OrderByColumn{
		Column: clause.Column{Name: "title"},
		Desc:   orderDesc,
	}).Limit(limit).Offset(offset).Find(&books).Error
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
