package book

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zura-t/bookstore_fiber/models"
	"github.com/zura-t/bookstore_fiber/pkg"
)

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
