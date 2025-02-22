package users

import (
	"context"
	"errors"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"net/mail"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Login(
	ctx context.Context,
	req *usersv1.LoginRequest,
) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	if req.Login == "" {
		return nil, errors.New("email is required")
	}

	parsedEmail, err := mail.ParseAddress(req.Login)
	if err != nil {
		return nil, errors.New("invalid email")
	}

	var client *database.Client
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

	return client, nil
}

func (s *Service) isPasswordValid(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)

	return err == nil
}
