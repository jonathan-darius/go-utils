package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/forkyid/go-utils/rest/restid"
)

// Claims claims
type Claims struct {
	jwt.StandardClaims
	MemberID       restid.ID `json:"id"`
	RoleID         restid.ID `json:"role_id"`
	MemberUsername string    `json:"username"`
	Type           string    `json:"type"`
}
