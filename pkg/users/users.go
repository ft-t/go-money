package users

import (
	"context"
	"errors"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
	"github.com/ft-t/go-money/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

func (s *Service) ShouldCreateAdmin(ctx context.Context) (bool, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeReadonly))

	var count int64
	if err := db.Model(&database.User{}).Count(&count).Error; err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *Service) Create(
	ctx context.Context,
	req *usersv1.CreateRequest,
) (*usersv1.CreateResponse, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	shouldCreate, err := s.ShouldCreateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if !shouldCreate {
		return nil, errors.New("admin already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 5)
	if err != nil {
		return nil, err
	}

	targetUser := &database.User{
		Login:    req.Login,
		Password: string(hashedPassword),
	}

	if err = db.Create(targetUser).Error; err != nil {
		return nil, err
	}

	return &usersv1.CreateResponse{
		Id: targetUser.ID,
	}, nil
}

func (s *Service) Login(
	ctx context.Context,
	req *usersv1.LoginRequest,
) (*usersv1.LoginResponse, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeReadonly))

	if req.Login == "" {
		return nil, errors.New("login is required")
	}

	var user *database.User
	if err := db.Where("login = ?", req.Login).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}

		return nil, err
	}
	if !s.isPasswordValid(user.Password, []byte(req.Password)) {
		return nil, errors.New("password is invalid")
	}

	token, err := s.cfg.JwtSvc.GenerateToken(ctx, user)
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
