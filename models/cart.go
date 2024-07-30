package models

import "time"

type CartItem struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	BookID    uint      `json:"book_id"`
	Book      Book      `json:"book"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
