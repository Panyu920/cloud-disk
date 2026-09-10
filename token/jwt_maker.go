package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct {
	secret string
}

func NewJWTMaker(secret string) *JWTMaker {
	return &JWTMaker{secret: secret}
}

func (m *JWTMaker) CreateToken(username string, userID int64, duration time.Duration) (string, *Payload, error) {
	p, err := NewPayload(username, userID, duration)
	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", nil, err
	}
	return tokenString, p, nil
}
func (m *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.secret), nil
	}
	tokenMaker, err := jwt.ParseWithClaims(token, &Payload{}, keyfunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
	}
	p, ok := tokenMaker.Claims.(*Payload)

	if !tokenMaker.Valid || p.Valid() != nil || !ok {
		return nil, ErrInvalidToken
	}
	return p, nil
}
