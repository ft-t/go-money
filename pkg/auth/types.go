package auth

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

const (
	ServiceTokenType = "service_token"
)

type JwtClaims struct {
	*jwt.RegisteredClaims
	UserID    int32  `json:"user_id"`
	TokenType string `json:"token_type"`
}

type GenerateTokenRequest struct {
	TTL       time.Duration
	TokenType string
	User      *database.User
}
