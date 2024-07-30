package cart

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/config"
	"github.com/zura-t/bookstore_fiber/middlewares/auth"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
)

type cartRouter struct {
	log    *logrus.Logger
	config config.Config
	db     *gorm.DB
}

func NewCartRouter(app *fiber.App, log *logrus.Logger, config config.Config, db *gorm.DB, token *token.JwtMaker) {
	r := &cartRouter{log, config, db}
	authRoutes := app.Group("/", auth.New(log, token))

	cartRoutes := authRoutes.Group("/cart")
	cartRoutes.Post("/", r.AddBookToCart)
	cartRoutes.Get("/", r.GetBooksInCart)
	cartRoutes.Delete("/:id", r.DeleteBookFromCart)
	cartRoutes.Delete("/", r.DeleteAllBooksFromCart)
}

type AddBookToCart struct {
	BookId uint `json:"book_id" validate:"required,min=1"`
}

func ConvertCartItem(cartItem models.CartItem) CartItemResponse {
	return CartItemResponse{
		ID:     cartItem.ID,
		UserID: cartItem.UserID,
		BookID: cartItem.BookID,
		Book: BookItemResponse{
			Id:         cartItem.Book.ID,
			Title:      cartItem.Book.Title,
			Price:      cartItem.Book.Price,
			AuthorID:   cartItem.Book.AuthorID,
			AuthorName: cartItem.Book.Author.Name,
			CreatedAt:  cartItem.Book.CreatedAt,
			UpdatedAt:  cartItem.Book.UpdatedAt,
		},
		CreatedAt: cartItem.CreatedAt,
		UpdatedAt: cartItem.UpdatedAt,
	}
}

func (r cartRouter) AddBookToCart(c *fiber.Ctx) error {
	var req = &AddBookToCart{}
	if err := c.BodyParser(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			var validation_errs = make([]string, len(validationErrors))
			for l, validation_err := range validationErrors {
				field, ok := reflect.TypeOf(req).Elem().FieldByName(validation_err.StructField())
				fieldName := field.Tag.Get("json")
				if !ok {
					panic("Field not found")
				}
				validation_errs[l] = pkg.MsgForTag(validation_err, fieldName)
			}
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

	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var cartItem models.CartItem
	err := r.db.First(&cartItem, models.CartItem{BookID: req.BookId, UserID: data.UserId}).Error
	if err == nil {
		err := fmt.Errorf("You've already added this book to cart")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusNotFound).JSON(pkg.ErrorResponse(err))
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	var book models.Book

	err = r.db.Preload("Author").First(&book, req.BookId).Error
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

	cartItem = models.CartItem{
		UserID: data.UserId,
		BookID: book.ID,
	}

	err = r.db.Create(&cartItem).First(&cartItem).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	cartItem.Book = book
	res := ConvertCartItem(cartItem)

	return c.JSON(res)
}

type CartItemResponse struct {
	ID        uint             `json:"id"`
	UserID    uint             `json:"user_id"`
	BookID    uint             `json:"book_id"`
	Book      BookItemResponse `json:"book"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type BookItemResponse struct {
	Id         uint      `json:"id"`
	Title      string    `json:"title"`
	Price      uint      `json:"price"`
	AuthorID   uint      `json:"author_id"`
	AuthorName string    `json:"author_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (r cartRouter) GetBooksInCart(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var cartItems []models.CartItem
	err := r.db.Preload("Book", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Author")
	}).Find(&cartItems, &models.CartItem{UserID: data.UserId}).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := make([]*CartItemResponse, len(cartItems))
	for index, v := range cartItems {
		cartItem := ConvertCartItem(v)
		res[index] = &cartItem
	}
	return c.JSON(res)
}

func (r cartRouter) DeleteBookFromCart(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var req = &DeleteBookFromCart{}
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
			var validation_errs = make([]string, len(validationErrors))
			for l, validation_err := range validationErrors {
				field, ok := reflect.TypeOf(req).Elem().FieldByName(validation_err.StructField())
				fieldName := field.Tag.Get("json")
				if !ok {
					panic("Field not found")
				}
				validation_errs[l] = pkg.MsgForTag(validation_err, fieldName)
			}
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

	err := r.db.First(&book, req.Id).Error
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

	err = r.db.Delete(&models.CartItem{}, &models.CartItem{BookID: book.ID, UserID: data.UserId}).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	return c.SendString("Book has deleted")
}

type DeleteBookFromCart struct {
	Id uint `uri:"book_id" json:"book_id" validate:"required,min=1"`
}

func (r *cartRouter) DeleteAllBooksFromCart(c *fiber.Ctx) error {
	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	err := r.db.Delete(models.CartItem{}, models.CartItem{UserID: data.UserId}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := fmt.Errorf("CartItem not found")
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

	return c.SendString("CartItems deleted")
}
