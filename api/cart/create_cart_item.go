package cart

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
)

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
