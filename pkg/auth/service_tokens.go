package auth

import (
	"context"
	"time"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm"
)

type ServiceTokenService struct {
	jwtSvc JwtSvc
	mapper ServiceTokenMapper
}

func NewServiceTokenService(
	jwtSvc JwtSvc,
	mapper ServiceTokenMapper,
) *ServiceTokenService {
	return &ServiceTokenService{
		jwtSvc: jwtSvc,
		mapper: mapper,
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

	generated, str, err := s.jwtSvc.CreateServiceToken(ctx, &GenerateTokenRequest{
		TTL:  time.Until(req.Req.ExpiresAt.AsTime()),
		User: user,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate service token")
	}

	token := &database.ServiceToken{
		ID:        generated.ID,
		Name:      req.Req.Name,
		ExpiresAt: req.Req.ExpiresAt.AsTime(),
		CreatedAt: time.Now().UTC(),
	}

	if err = tx.Create(token).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create service token")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return &configurationv1.CreateServiceTokenResponse{
		ServiceToken: s.mapper.MapServiceToken(ctx, token),
		Token:        str,
	}, nil
}

func (s *ServiceTokenService) GetServiceTokens(
	ctx context.Context,
	req *configurationv1.GetServiceTokensRequest,
) (*configurationv1.GetServiceTokensResponse, error) {
	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly))

	var tokens []*database.ServiceToken

	query := db.Model(&database.ServiceToken{})
	if len(req.Ids) > 0 {
		query = query.Where("id IN ?", req.Ids)
	}

	if err := query.Find(&tokens).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get service tokens")
	}

	resp := &configurationv1.GetServiceTokensResponse{
		ServiceTokens: make([]*gomoneypbv1.ServiceToken, 0, len(tokens)),
	}

	for _, token := range tokens {
		resp.ServiceTokens = append(resp.ServiceTokens, s.mapper.MapServiceToken(ctx, token))
	}

	return resp, nil
}

func (s *ServiceTokenService) RevokeServiceToken(
	ctx context.Context,
	req *configurationv1.RevokeServiceTokenRequest,
) (*configurationv1.RevokeServiceTokenResponse, error) {
	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	var token database.ServiceToken
	if err := tx.Where("id = ?", req.Id).First(&token).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find service token")
	}

	token.DeletedAt = gorm.DeletedAt{Time: time.Now().UTC(), Valid: true}

	if err := tx.Save(&token).Error; err != nil {
		return nil, errors.Wrap(err, "failed to revoke service token")
	}

	if err := s.jwtSvc.RevokeServiceToken(ctx, token.ID, token.ExpiresAt); err != nil {
		return nil, errors.Wrap(err, "failed to revoke jwt token")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return &configurationv1.RevokeServiceTokenResponse{
		Id: req.Id,
	}, nil
}
