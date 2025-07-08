package users

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"todo-app/pkg/locale"
)

type Service interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) (User, error)
}

type service struct {
	logger     *zap.SugaredLogger
	repository Repository
	validator  *validator.Validate
}

func GetService(logger *zap.SugaredLogger, repo Repository, validator *validator.Validate) Service {
	return &service{
		logger:     logger,
		repository: repo,
		validator:  validator,
	}
}

func (s *service) Create(ctx context.Context, user *User) error {
	if err := s.validator.Struct(user); err != nil {
		return err
	}

	return s.repository.Create(ctx, user)
}

func (s *service) Update(ctx context.Context, user *User) (User, error) {
	actualUser, err := s.repository.GetById(ctx, user.ID)
	if err != nil {
		return User{}, err
	}

	updates := map[string]interface{}{}

	if actualUser.Email != user.Email {
		updates["email"] = user.Email
	}
	if actualUser.FirstName != user.FirstName {
		updates["first_name"] = user.FirstName
	}
	if actualUser.LastName != user.LastName {
		updates["last_name"] = user.LastName
	}
	if bcrypt.CompareHashAndPassword([]byte(actualUser.Password), []byte(user.Password)) != nil {
		updates["password"] = user.Password
	}

	if len(updates) == 0 {
		return User{}, errors.New(locale.ErrorNotFoundUpdates)
	}

	err = s.repository.Update(ctx, user.ID, updates)
	if err != nil {
		return User{}, err
	}

	updatedUser, err := s.repository.GetById(ctx, user.ID)
	if err != nil {
		return User{}, errors.New(locale.ErrorNotFoundRecord)
	}

	updatedUser.Password = ""
	return updatedUser, nil
}
