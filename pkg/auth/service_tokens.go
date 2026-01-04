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
		ExpiresAt: req.Req.ExpiresAt.AsTime(),
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
		pbToken := &gomoneypbv1.ServiceToken{
			Id:        token.ID,
			Name:      token.Name,
			CreatedAt: timestamppb.New(token.CreatedAt),
			ExpiresAt: timestamppb.New(token.ExpiresAt),
		}

		if token.DeletedAt.Valid {
			pbToken.DeletedAt = timestamppb.New(token.DeletedAt.Time)
		}

		resp.ServiceTokens = append(resp.ServiceTokens, pbToken)
	}

	return resp, nil
}

func (s *ServiceTokenService) RevokeServiceToken(
	ctx context.Context,
	req *configurationv1.RevokeServiceTokenRequest,
) (*configurationv1.RevokeServiceTokenResponse, error) {
	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	result := db.Delete(&database.ServiceToken{}, "id = ?", req.Id)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to revoke service token")
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("service token not found")
	}

	return &configurationv1.RevokeServiceTokenResponse{
		Id: req.Id,
	}, nil
}

func (s *ServiceTokenService) IsRevoked(
	ctx context.Context,
	tokenID string,
) (bool, error) {
	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly))

	var token database.ServiceToken
	err := db.Unscoped().Where("id = ?", tokenID).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, errors.Wrap(err, "failed to check token revocation")
	}

	return token.DeletedAt.Valid, nil
}
