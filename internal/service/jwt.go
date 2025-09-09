package service

import (
	"errors"
	"time"

	"detector_plagio/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secretKey []byte
}

func NewJWT(secretKey string) *JWT {
	return &JWT{secretKey: []byte(secretKey)}
}

func (s *JWT) GenerateToken(user *domain.User) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWT) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})
}
