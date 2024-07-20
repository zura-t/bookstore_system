package entity

import "time"

type User struct {
	Id        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsAuthor  bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
