package auth

import (
	"context"
	"time"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type ServiceTokenService struct {
	JwtSvc JwtSvc
}

func NewServiceTokenService(
	jwtSvc JwtSvc,
) *ServiceTokenService {
	return &ServiceTokenService{
		JwtSvc: jwtSvc,
	}
}

func (s *ServiceTokenService) CreateServiceToken(
	ctx context.Context,
	req *CreateServiceTokenRequest,
) (*configurationv1.CreateServiceTokenResponse, error) {
	tx := database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	if req.Req.ExpiresAt == nil {
		return nil, errors.New("expiresAt is required")
	}

	var user *database.User
	if err := tx.Where("id = ?", req.CurrentUserID).First(&user).Error; err != nil {
		return nil, err
	}

	generated, str, err := s.JwtSvc.CreateServiceToken(ctx, &GenerateTokenRequest{
		TTL:  time.Until(req.Req.ExpiresAt.AsTime()),
		User: user,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate service token")
	}

	token := &database.ServiceToken{
		ID:        generated.ID,
		Name:      req.Req.Name,
		ExpiresAt: time.Time{},
		DeletedAt: gorm.DeletedAt{},
		CreatedAt: time.Now().UTC(),
	}

	if err = tx.Create(token).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create service token")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	resp := &configurationv1.CreateServiceTokenResponse{
		ServiceToken: &gomoneypbv1.ServiceToken{
			Id:        token.ID,
			Name:      token.Name,
			CreatedAt: timestamppb.New(token.CreatedAt),
			ExpiresAt: timestamppb.New(token.ExpiresAt),
		},
		Token: str,
	}

	if token.DeletedAt.Valid {
		resp.ServiceToken.DeletedAt = timestamppb.New(token.DeletedAt.Time)
	}

	return resp, nil
}
