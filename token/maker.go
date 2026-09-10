package token

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type TokenMaker interface {
	CreateToken(username string, userID int64, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}

var DefaultMaker TokenMaker

func init() {
	DefaultMaker = NewJWTMaker("123456")
}
