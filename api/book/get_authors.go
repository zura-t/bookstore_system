package book

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
	"gorm.io/gorm/clause"
)

type GetAuthors struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	Name      string `json:"name"`
	OrderDesc bool   `json:"order_desc"`
}

func (r *bookRouter) GetAuthors(c *fiber.Ctx) error {
	req := &GetAuthors{Limit: 20, Offset: 0, Name: "", OrderDesc: false}
	if err := c.BodyParser(req); err != nil {
		r.log.WithFields(logrus.Fields{
			"level": "Error",
		}).Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(pkg.ErrorResponse(err))
	}

	var authors []models.User
	err := r.db.Preload("AuthorBooks").Where(models.User{IsAuthor: true, Name: req.Name}).Order(clause.OrderByColumn{
		Column: clause.Column{Name: "name"},
		Desc:   req.OrderDesc,
	}).Limit(req.Limit).Offset(req.Offset).Find(&authors).Error
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
