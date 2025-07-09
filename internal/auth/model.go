package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token     string   `json:"token"`
	Refresh   string   `json:"refresh_token"`
	ExpiresAt int64    `json:"expires_at"`
	User      UserInfo `json:"user"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type RefreshToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"default:false"`
}
