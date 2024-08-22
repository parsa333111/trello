package token_utils

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

const (
	Refresh TokenType = iota + 1
	Access
)

type TokenType int

type CustomClaims struct {
	UserID uint
	Type   TokenType
	jwt.RegisteredClaims
}

var secretKey = []byte("e903e653cff56bba3447b0857c18cd7d0e1680ec9ca6dd59d51ee926098bf4c1")

const refreshTokenValidPeriod = time.Hour * 24
const accessTokenValidPeriod = time.Minute * 30

func calculateValidPeriod(token_type TokenType) time.Time {
	if token_type == Refresh {
		return time.Now().Add(refreshTokenValidPeriod)
	} else {
		return time.Now().Add(accessTokenValidPeriod)
	}
}

func GenerateToken(user_id uint, token_type TokenType) (string, error) {
	claims := &CustomClaims{
		UserID: user_id,
		Type:   token_type,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(calculateValidPeriod(token_type)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token_string, err := token.SignedString(secretKey)
	if err != nil {
		log.Println("Error:", err)
		return "", custom_errors.ErrTokenFailure
	}

	return token_string, nil
}

func GenerateTokens(user_id uint) (string, string, error) {
	refresh_token, err := GenerateToken(user_id, Refresh)
	if err != nil {
		return "", "", err
	}

	access_token, err := GenerateToken(user_id, Access)
	if err != nil {
		return "", "", err
	}

	return refresh_token, access_token, nil
}

func ParseToken(token_string string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(token_string, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		log.Println("Error:", err)
		return nil, custom_errors.ErrTokenFailure
	}
}
