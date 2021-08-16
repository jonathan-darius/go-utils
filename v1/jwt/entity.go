package jwt

import (
	"github.com/dgrijalva/jwt-go"
)

// Claims claims
type Claims struct {
	jwt.StandardClaims
	MemberID       string `json:"id"`
	RoleID         string `json:"role_id"`
	MemberUsername string `json:"username"`
	Type           string `json:"type"`
}
