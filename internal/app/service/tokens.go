package service

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/config"
	appErrors "github.com/ujwegh/shortener/internal/app/errors"
	"time"
)

type TokenService interface {
	GetUserUID(tokenString string) (string, error)
	GenerateToken(userUID *uuid.UUID) (string, error)
}

type Claims struct {
	jwt.RegisteredClaims
	UserUID string
}

type TokenServiceImpl struct {
	secretKey string
}

func NewTokenService(cfg config.AppConfig) *TokenServiceImpl {
	return &TokenServiceImpl{
		secretKey: cfg.TokenSecretKey,
	}
}

func (ts TokenServiceImpl) GetUserUID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(ts.secretKey), nil
		})
	if err != nil {
		return "", appErrors.New(err, "failed to parse token")
	}

	if !token.Valid {
		return "", appErrors.New(
			errors.New("token error"),
			"token is not valid",
		)
	}
	return claims.UserUID, nil
}

func (ts TokenServiceImpl) GenerateToken(userUID *uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject:  "auth token",
		},
		UserUID: userUID.String(),
	})

	tokenString, err := token.SignedString([]byte(ts.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
