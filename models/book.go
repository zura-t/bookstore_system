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
	AuthorID    uint      `json:"author_id"`
	Author      User      `gorm:"foreignKey:AuthorID" json:"author"`
	File        string    `json:"file"`
}
