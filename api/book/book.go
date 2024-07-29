package book

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	role "github.com/zura-t/bookstore_fiber/middlewares/roles"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
)

type bookRouter struct {
	log    *logrus.Logger
	config config.Config
	db     *gorm.DB
}

func NewBookRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB, token *token.JwtMaker) {
	r := &bookRouter{log, config, db}

	authRoutes := app.Group("/", auth.New(log, token))
	authRoutes.Get("/authors", r.GetAuthors)
	authRoutes.Get("/", r.GetBooks)
	authRoutes.Get("/:id", r.GetBook)

	bookRoutes := authRoutes.Group("/books", role.New(log))
	bookRoutes.Post("/", r.UploadBook)
	bookRoutes.Patch("/:id", r.UpdateBook)
	bookRoutes.Delete("/:id", r.DeleteBook)
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
	Id uint `uri:"id" validate:"required,min=1"`
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

func (r *bookRouter) GetBook(c *fiber.Ctx) error {
	var req BookId
	if err := c.ParamsParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
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

func (r *bookRouter) GetAuthors(c *fiber.Ctx) error {
	var authors []models.User
	err := r.db.Preload("AuthorBooks").Where(models.User{IsAuthor: true}).Find(&authors).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := make([]*AuthorsResponse, len(authors))
	for index, v := range authors {
		author := ConvertAuthors(v)
		res[index] = &author
	}

	return c.JSON(res)
}

type AuthorsResponse struct {
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	BooksQuantity int    `json:"books_quantity"`
}

func ConvertAuthors(author models.User) AuthorsResponse {
	return AuthorsResponse{
		Id:            author.ID,
		Name:          author.Name,
		BooksQuantity: len(author.AuthorBooks),
	}
}

func (r *bookRouter) GetAuthorBooks(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var books []models.Book
	err := r.db.Find(&books).Where("AuthorID", data.ID).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.JSON(books)
}

// upload file
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

	var book UploadBook
	if err := c.BodyParser(&book); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(book); err != nil {
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

	return c.JSON(arg)
}

type UploadBook struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Price       uint   `json:"price" validate:"required,min=1"`
}

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

	var book BookUpdate
	if err := c.BodyParser(&book); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(book); err != nil {
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

	err = r.db.Model(&models.Book{}).Where(&models.Book{ID: book.Id, AuthorID: data.UserId}).Updates(&book).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	var res models.Book
	err = r.db.First(&res, book.Id).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.JSON(res)
}

type BookUpdate struct {
	Id          uint   `json:"id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Price       uint   `json:"price" validate:"required,min=1"`
	File        string `json:"file"`
}

func (r *bookRouter) DownloadBook(c *fiber.Ctx) error {
	var req BookId
	if err := c.ParamsParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}
	var book models.Book
	err := r.db.First(&book, req.Id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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

	c.Response().Header.Set("Content-Disposition", "attachment; filename=WHATEVER_YOU_WANT")
	content_type := string(c.Request().Header.ContentType())
	c.Response().Header.Set("Content-Type", content_type)
	return c.SendFile(book.File)
}

func (r *bookRouter) DeleteBook(c *fiber.Ctx) error {
	var req BookId
	if err := c.ParamsParser(&req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var book models.Book
	err := r.db.Delete(&book, &models.Book{ID: req.Id, AuthorID: data.UserId}).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.SendString("Book deleted")
}
