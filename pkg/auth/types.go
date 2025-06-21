package auth

import (
	"github.com/cockroachdb/errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

type JwtClaims struct {
	*jwt.RegisteredClaims
	UserID int32 `json:"user_id"`
}

func (j *JwtClaims) Valid() error {
	if j.RegisteredClaims == nil {
		return errors.New("registered claims is nil")
	}

	// check expires
	if j.ExpiresAt.Before(time.Now().UTC()) {
		return errors.New("token is expired")
	}

	return nil
}
