package auth

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"todo-app/internal/users"
	"todo-app/pkg/locale"
)

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAuthRepo := NewMockRepository(ctrl)
	mockUserRepo := users.NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUserRepo, mockAuthRepo, v)
	ctx := context.Background()

	password := "test"
	userPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := users.User{
		ID:        1,
		Email:     "test@test.com",
		FirstName: "test",
		LastName:  "test",
		Password:  string(userPassword),
	}

	t.Run("successful login", func(t *testing.T) {
		mockUserRepo.
			EXPECT().
			GetByEmail(ctx, user.Email).
			Return(user, nil).
			Times(1)

		mockAuthRepo.
			EXPECT().
			SaveRefreshToken(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		loginRequest := LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		response, err := service.Login(ctx, loginRequest)
		assert.NoError(t, err)
		assert.Equal(t, user.FirstName, response.User.FirstName)
		assert.Equal(t, user.LastName, response.User.LastName)
		assert.Equal(t, user.Email, response.User.Email)

		ctrl.Finish()
	})

	t.Run("failed login", func(t *testing.T) {
		mockUserRepo.
			EXPECT().
			GetByEmail(ctx, user.Email).
			Return(user, nil).
			Times(1)

		loginRequest := LoginRequest{
			Email:    user.Email,
			Password: "wrong_password",
		}

		response, err := service.Login(ctx, loginRequest)
		assert.Error(t, err)
		assert.Equal(t, response, LoginResponse{})
		assert.Equal(t, err, errors.New(locale.ErrorInvalidCredentials))

		ctrl.Finish()
	})
}
