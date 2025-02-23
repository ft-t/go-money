package users

import (
	"context"
	"errors"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
	"github.com/ft-t/go-money/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"net/mail"
)

type Service struct {
	cfg *ServiceConfig
}

type ServiceConfig struct {
	JwtSvc JwtSvc
}

func NewService(
	cfg *ServiceConfig,
) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) Login(
	ctx context.Context,
	req *usersv1.LoginRequest,
) (*usersv1.LoginResponse, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	if req.Login == "" {
		return nil, errors.New("email is required")
	}

	parsedEmail, err := mail.ParseAddress(req.Login)
	if err != nil {
		return nil, errors.New("invalid email")
	}

	var client *database.User
	if err = db.Where("email = ?", parsedEmail.Address).
		First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}

		return nil, err
	}
	if client.Password == "" {
		return nil, errors.New("password not set")
	}

	if !s.isPasswordValid(client.Password, []byte(req.Password)) {
		return nil, errors.New("password is invalid")
	}

	token, err := s.cfg.JwtSvc.GenerateToken(ctx, client)
	if err != nil {
		return nil, err
	}

	return &usersv1.LoginResponse{
		Token: token,
	}, nil
}

func (s *Service) isPasswordValid(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)

	return err == nil
}
