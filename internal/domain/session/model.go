package session

import "time"

type Session struct {
	ID               string    `json:"id" bson:"id" gorm:"primaryKey;type:uuid"`
	UserID           string    `json:"user_id" bson:"user_id" gorm:"not null"`
	IPAddress        string    `json:"ip_address" bson:"ip_address"`
	UserAgent        string    `json:"user_agent" bson:"user_agent"`
	Valid            bool      `json:"valid" bson:"valid" gorm:"default:true"`
	ExpiresAt        time.Time `json:"expires_at" bson:"expires_at"`
	RefreshTokenHash string    `json:"-" bson:"refresh_token_hash" gorm:"column:refresh_token_hash"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" bson:"updated_at"`
}

func (Session) TableName() string {
	return "sessions"
}

type SessionQueryParams struct {
	UserID *string
	Valid  *bool
}
