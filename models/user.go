package models

import "time"

type User struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Email       string    `gorm:"uniqueIndex" json:"email"`
	Password    string    `json:"password"`
	IsAuthor    bool      `gorm:"default:false" json:"is_author"`
	Cart        Cart      `json:"cart"`
	ReadList    []Book    `gorm:"many2many:user_books;" json:"read_list"`
	AuthorBooks []Book    `gorm:"foreignKey:AuthorID" json:"author_books"`
}
