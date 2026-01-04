package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type Service struct {
	privateKey *rsa.PrivateKey
	ttl        time.Duration
	cache      *expirable.LRU[string, error]
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
		cache:      expirable.NewLRU[string, error](1000, nil, time.Minute*5),
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

	claims, err := j.CheckClaims(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate token claims")
	}

	if cachedError, ok := j.cache.Get(claims.ID); ok {
		if cachedError != nil {
			return nil, cachedError
		}

		return claims, nil
	}

	if claims.TokenType == ServiceTokenType {
		var jtiRevocation database.JtiRevocation

		if err = database.GetDb(database.DbTypeReadonly).
			Where("id = ?", claims.ID).
			Limit(1).
			Find(&jtiRevocation).Error; err != nil {
			return nil, errors.Wrap(err, "failed to check jti revocation")
		}

		if jtiRevocation.ID != "" {
			err = errors.New("token has been revoked")
			j.cache.Add(claims.ID, err)

			return nil, err
		}
	}

	j.cache.Add(claims.ID, err)

	return claims, err
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
	_, token, err := j.generateTokenInternal(&GenerateTokenRequest{
		TTL:       j.ttl,
		TokenType: "web",
		User:      user,
	})

	return token, err
}

func (j *Service) RevokeServiceToken(
	ctx context.Context,
	jti string,
	originalExpiryTime time.Time,
) error {
	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	if err := db.Create(&database.JtiRevocation{
		ID:        jti,
		ExpiresAt: originalExpiryTime.Add(7 * 24 * time.Hour),
	}).Error; err != nil {
		return errors.Wrap(err, "failed to store jti for service token")
	}

	return nil
}

func (j *Service) CreateServiceToken(
	_ context.Context,
	req *GenerateTokenRequest,
) (*JwtClaims, string, error) {
	if req.User == nil {
		return nil, "", errors.New("user is required to generate service token")
	}

	req.TokenType = ServiceTokenType
	if req.TTL == 0 {
		return nil, "", errors.New("ttl is required to generate service token")
	}

	claims, token, err := j.generateTokenInternal(req)
	if err != nil {
		return nil, "", err
	}

	return claims, token, err
}

func (j *Service) generateTokenInternal(req *GenerateTokenRequest) (*JwtClaims, string, error) {
	claims := &JwtClaims{
		RegisteredClaims: &jwt2.RegisteredClaims{
			ExpiresAt: jwt2.NewNumericDate(time.Now().UTC().Add(j.ttl)),
			ID:        uuid.NewString(),
		},
		UserID:    req.User.ID,
		TokenType: req.TokenType,
	}

	token := jwt2.NewWithClaims(jwt2.SigningMethodRS256, claims)
	a, err := token.SignedString(j.privateKey)
	if err != nil {
		return nil, "", err
	}

	return claims, a, nil
}
