package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type Service struct {
	privateKey *rsa.PrivateKey
	ttl        time.Duration
}

func NewService(
	secretKey string,
	ttl time.Duration,
) (*Service, error) {
	block, _ := pem.Decode([]byte(secretKey))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private key: %v")
	}

	return &Service{
		privateKey: privateKey,
		ttl:        ttl,
	}, nil
}

func (j *Service) ValidateToken(
	_ context.Context,
	tokenString string,
) (*JwtClaims, error) {
	token, err := jwt2.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt2.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt2.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.privateKey.Public(), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}
	
	return j.CheckClaims(token)
}

func (j *Service) CheckClaims(
	token *jwt2.Token,
) (*JwtClaims, error) {
	if token == nil {
		return nil, errors.New("token is nil")
	}

	claims, ok := token.Claims.(*JwtClaims)
	if !ok || claims == nil {
		return nil, errors.New("invalid token claims")
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}

func (j *Service) GenerateToken(
	_ context.Context,
	user *database.User,
) (string, error) {
	claims := &JwtClaims{
		RegisteredClaims: &jwt2.RegisteredClaims{
			ExpiresAt: jwt2.NewNumericDate(time.Now().UTC().Add(j.ttl)),
			ID:        uuid.NewString(),
		},
		UserID: user.ID,
	}

	token := jwt2.NewWithClaims(jwt2.SigningMethodRS256, claims)
	a, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", err
	}

	return a, nil
}
