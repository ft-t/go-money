package auth

import (
	"github.com/cockroachdb/errors"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type JwtClaims struct {
	*jwt.RegisteredClaims
	UserID int32 `json:"user_id"`
}
