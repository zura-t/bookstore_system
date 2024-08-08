package models

import (
	"time"
)

type Book struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       uint      `json:"price"`
	ReadList    []User    `gorm:"many2many:user_books;" json:"read_list"`
	AuthorID    uint      `json:"author_id"`
	Author      User      `gorm:"foreignKey:AuthorID" json:"author"`
	File        string    `json:"file"`
}

type UserBook struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at" json:"created_at"`
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;references:ID"`
	BookID    uint      `gorm:"primaryKey" json:"book_id"`
	Book      Book      `gorm:"foreignKey:BookID;references:ID"`
}
