package jwt

import (
	"github.com/dgrijalva/jwt-go"
)

// UserClaims type Scheme
type UserClaims struct {
	ID          string `json:"id"`
	RoleID      string `json:"role_id"`
	IsCreator   bool   `json:"is_creator"`
	IsVerified  bool   `json:"is_verified"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	PhoneCode   string `json:"phone_code"`
	PhoneNumber string `json:"phone_number"`
	Language    string `json:"language"`
}

// RefreshClaims type Scheme
type RefreshClaims struct {
	jwt.StandardClaims
	Type string `json:"type"`
}

// AccessClaims type Scheme
type AccessClaims struct {
	jwt.StandardClaims
	Type string `json:"type"`
	UserClaims
}
