package users

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/cockroachdb/errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"time"
)

type JwtGenerator struct {
	privateKey *rsa.PrivateKey
}

func NewJwtGenerator(secretKey string) *JwtGenerator {
	block, _ := pem.Decode([]byte(secretKey))
	if block == nil {
		panic("Failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(errors.Wrap(err, "Failed to parse private key: %v"))
	}

	return &JwtGenerator{
		privateKey: privateKey,
	}
}

func (j *JwtGenerator) GenerateToken(
	_ context.Context,
	user *database.User,
) (string, error) {
	claims := &boilerplate.JwtClaims{
		RegisteredClaims: &jwt2.RegisteredClaims{
			ExpiresAt: jwt2.NewNumericDate(time.Now().UTC().Add(time.Hour * 24 * 2)),
		},
		UserID: user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	a, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", err
	}

	return a, nil
}
