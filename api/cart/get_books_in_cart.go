package cart

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm"
	"time"
)

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
