package utils

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	hmacSecret = "eyJhbGcADA&*DHAJH!@(hjke"
)

// HashPassword  returns hash of provided password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// CheckPasswordHash returns true if hash is equivalent of password otherwhise false
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// UserClaims represents JWT claims with custom field UserId
type UserClaims struct {
	jwt.StandardClaims
	UserId string `json:"user_id"`
}

// NewUserToken creates new user token with provided data
func NewUserToken(userId string, expiresAt int64) (string, error) {
	claims := UserClaims{
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
		userId,
	}
	userToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return userToken.SignedString([]byte(hmacSecret))
}

// DecodeUserToken decodes tokenString to UserClaims object
func DecodeUserToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(hmacSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {

		return nil, fmt.Errorf("Invalid token: %v", token)
	}

	return claims, nil
}
