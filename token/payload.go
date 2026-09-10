package token

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

type Payload struct {
	Username string `json:"username"`
	UserID   int64  `json:"user_id"`
	jwt.RegisteredClaims
}

func NewPayload(username string, userID int64, duration time.Duration) (*Payload, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &Payload{
		Username: username,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}, nil
}

func (p *Payload) Valid() error {
	if p.ExpiresAt.Before(time.Now()) {
		return errors.New("token expired")
	}
	return nil
}
