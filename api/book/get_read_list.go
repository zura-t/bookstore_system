package book

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"github.com/zura-t/bookstore_fiber/token"
	"gorm.io/gorm/clause"
)

type GetReadList struct {
	Limit  int `form:"limit,default=20"`
	Offset int `form:"offset,default=0"`
}

func (r bookRouter) GetReadList(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	payload := c.Locals("user")
	data, ok := payload.(*token.Payload)
	if !ok {
		err := fmt.Errorf("Can't get payload")
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var user = models.User{ID: data.UserId}
	user_book := models.UserBook{UserID: user.ID}
	books := []models.UserBook{}

	err := r.db.Preload("Book.Author").Where(&user_book).Order(clause.OrderByColumn{
		Column: clause.Column{Name: "created_at"},
		Desc:   true,
	}).Limit(limit).Offset(offset).Find(&books).Error

	if err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(pkg.ErrorResponse(err))
	}

	res := make([]*BookResponse, len(books))
	for index, v := range books {
		v.Book.AuthorID = user.ID
		book := ConvertBook(v.Book)
		res[index] = &book
	}

	return c.JSON(res)
}
