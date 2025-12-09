package user

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           string    `json:"id" bson:"id"`
	Username     string    `json:"username" bson:"username"`
	Fullname     string    `json:"fullname" bson:"fullname"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-" bson:"password_hash"`
	Avatar       *string   `json:"avatar" bson:"avatar"`
	Role         Role      `json:"role" bson:"role"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}
