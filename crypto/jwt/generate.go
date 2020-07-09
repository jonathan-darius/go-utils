package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/forkyid/go-utils/rest/restid"
)

var accessExpirationDuration = time.Duration(AccessTokenDurationMinute) * time.Minute
var refreshExpirationDuration = time.Duration(RefreshTokenDurationHour) * time.Hour
var jwtSigningMethod = jwt.SigningMethodHS256

// GenerateAccessToken params
// @memberID: string
// @username: string
// return error, string
func GenerateAccessToken(memberID, roleID restid.ID, username string) (string, error) {
	applicationName := os.Getenv("APPLICATION_NAME")
	if applicationName == "" {
		applicationName = AppName
	}
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    applicationName,
			ExpiresAt: time.Now().Add(accessExpirationDuration).Unix(),
		},
		MemberID:       memberID,
		MemberUsername: username,
		RoleID:         roleID,
		Type:           "Access-Token",
	}

	token := jwt.NewWithClaims(jwtSigningMethod, claims)

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_ACCESS_SIGNATURE_KEY")))
	if err != nil {
		return "", errors.New("token signing error")
	}

	return signedToken, nil
}

// GenerateRefreshToken params
// @memberID: string
// @username: string
// return error, string
func GenerateRefreshToken(memberID, roleID restid.ID, username string) (string, error) {
	applicationName := os.Getenv("APPLICATION_NAME")
	if applicationName == "" {
		applicationName = AppName
	}
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    applicationName,
			ExpiresAt: time.Now().Add(refreshExpirationDuration).Unix(),
		},
		MemberID:       memberID,
		MemberUsername: username,
		RoleID:         roleID,
		Type:           "Refresh-Token",
	}

	token := jwt.NewWithClaims(
		jwtSigningMethod,
		claims,
	)

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_REFRESH_SIGNATURE_KEY")))
	if err != nil {
		return "", errors.New("token signing error")
	}

	return signedToken, nil
}
