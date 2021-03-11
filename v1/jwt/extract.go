package jwt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/forkyid/go-utils/v1/aes"
)

// ExtractClaims extracts claims from JWT, returns claims as map
func ExtractClaims(tokenStr string) (jwt.MapClaims, bool) {
	hmacSecretString := os.Getenv("JWT_ACCESS_SIGNATURE_KEY")
	hmacSecret := []byte(hmacSecretString)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// check token signing method etc
		return hmacSecret, nil
	})

	if err != nil {
		log.Println(err.Error())
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("Invalid JWT Token")
		return nil, false
	}

	return claims, true
}

// ExtractRefreshClaims params
// @tokenStr: string
// return jwt.MapClaims, error
func ExtractRefreshClaims(tokenStr string) (jwt.MapClaims, error) {
	hmacSecretString := os.Getenv("JWT_REFRESH_SIGNATURE_KEY")
	hmacSecret := []byte(hmacSecretString)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, err
	} else {
		return nil, err
	}
}

// ExtractID extracts only the id from JWT
func ExtractID(ah string) (int, error) {
	ts := strings.Replace(ah, "Bearer ", "", -1)
	claimsMap, claimRes := ExtractClaims(ts)
	if !claimRes {
		return -1, fmt.Errorf("Failed on claiming token")
	}
	id := aes.Decrypt(claimsMap["id"].(string))
	if id == -1 {
		return -1, fmt.Errorf("Invalid ID")
	}
	return id, nil
}

// ExtractClient extracts only the id from JWT
func ExtractClient(ah string) (*Claims, error) {
	ts := strings.Replace(ah, "Bearer ", "", -1)
	claims := Claims{}

	claimsMap, claimRes := ExtractClaims(ts)
	if !claimRes {
		return nil, fmt.Errorf("Failed on claiming token")
	}

	j, err := json.Marshal(&claimsMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

// ExtractRefresh extracts only the id from JWT
func ExtractRefresh(ah string) (*Claims, error) {
	ts := strings.Replace(ah, "Bearer ", "", -1)
	claims := Claims{}

	claimsMap, claimErr := ExtractRefreshClaims(ts)
	if claimErr != nil {
		return nil, fmt.Errorf("Failed on claiming token")
	}

	j, err := json.Marshal(&claimsMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}
