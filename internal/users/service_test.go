package users

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
	"todo-app/pkg/email"
	"todo-app/pkg/locale"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUsersRepo := NewMockRepository(ctrl)
	mockEmailService := email.NewMockService(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUsersRepo, v, mockEmailService)
	ctx := context.Background()

	user := User{
		Email:     "test@test.com",
		Password:  "password",
		FirstName: "test",
		LastName:  "test",
	}

	mockUsersRepo.
		EXPECT().
		Create(ctx, gomock.Any()).
		Return(nil).
		Times(1)

	mockEmailService.
		EXPECT().
		SendVerificationEmail(user.Email, user.FirstName, gomock.Any()).
		Return(nil).
		Times(1)

	err := service.Create(ctx, &user)

	assert.NoError(t, err)

	ctrl.Finish()
}

func TestService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUsersRepo := NewMockRepository(ctrl)
	mockEmailService := email.NewMockService(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUsersRepo, v, mockEmailService)
	ctx := context.Background()

	pw, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	user := User{
		ID:        1,
		Email:     "test@test.com",
		Password:  string(pw),
		FirstName: "test",
		LastName:  "test",
	}

	t.Run("update successfully", func(t *testing.T) {
		updatedUser := User{
			ID:        1,
			Email:     "test1@test.com",
			Password:  "password",
			FirstName: "new_test",
			LastName:  "test",
		}

		mockUsersRepo.
			EXPECT().
			GetById(ctx, user.ID).
			Return(user, nil).
			Times(1)

		mockUsersRepo.
			EXPECT().
			Update(ctx, user.ID, gomock.Any()).
			Return(nil).
			Times(1)

		mockEmailService.
			EXPECT().
			SendVerificationEmail(updatedUser.Email, updatedUser.FirstName, gomock.Any()).
			Return(nil).
			Times(1)

		mockUsersRepo.
			EXPECT().
			GetById(ctx, user.ID).
			Return(updatedUser, nil).
			Times(1)

		u, err := service.Update(ctx, &updatedUser)

		assert.NoError(t, err)
		assert.Equal(t, updatedUser, u)
		assert.Equal(t, user.ID, u.ID)
		assert.Equal(t, user.LastName, u.LastName)
		assert.NotEqual(t, user.FirstName, u.FirstName)
		assert.NotEqual(t, user.Email, u.Email)

		ctrl.Finish()
	})

	t.Run("no updates found", func(t *testing.T) {
		updatedUser := User{
			ID:        1,
			Email:     "test@test.com",
			Password:  "password",
			FirstName: "test",
			LastName:  "test",
		}

		mockUsersRepo.
			EXPECT().
			GetById(ctx, user.ID).
			Return(user, nil).
			Times(1)

		u, err := service.Update(ctx, &updatedUser)

		assert.Error(t, err)
		assert.Equal(t, user.ID, updatedUser.ID)
		assert.Equal(t, user.Email, updatedUser.Email)
		assert.Equal(t, user.FirstName, updatedUser.FirstName)
		assert.Equal(t, user.LastName, updatedUser.LastName)
		assert.Equal(t, User{}, u)
		assert.Contains(t, err.Error(), locale.ErrorNotFoundUpdates)

		ctrl.Finish()
	})
}

func TestService_VerifyEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUsersRepo := NewMockRepository(ctrl)
	mockEmailService := email.NewMockService(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUsersRepo, v, mockEmailService)
	ctx := context.Background()

	user := User{
		ID:        1,
		Email:     "test@test.com",
		Password:  "password",
		FirstName: "test",
		LastName:  "test",
	}
	token := "verification_token"

	t.Run("send email verification successfully", func(t *testing.T) {
		expTime := time.Now().Add(time.Hour)
		user.EmailVerificationExpiry = &expTime

		mockUsersRepo.
			EXPECT().
			GetByEmailVerificationToken(ctx, token).
			Return(user, nil).
			Times(1)

		mockUsersRepo.
			EXPECT().
			VerifyEmail(ctx, user.ID).
			Return(nil).
			Times(1)

		err := service.VerifyEmail(ctx, token)

		assert.NoError(t, err)
		ctrl.Finish()
	})

	t.Run("token expired", func(t *testing.T) {
		expTime := time.Now().Add(-time.Hour)
		user.EmailVerificationExpiry = &expTime

		mockUsersRepo.
			EXPECT().
			GetByEmailVerificationToken(ctx, token).
			Return(user, nil).
			Times(1)

		err := service.VerifyEmail(ctx, token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has expired")
		ctrl.Finish()
	})

	t.Run("email verified", func(t *testing.T) {
		expTime := time.Now().Add(time.Hour)
		user.EmailVerificationExpiry = &expTime
		user.IsEmailVerified = true

		mockUsersRepo.
			EXPECT().
			GetByEmailVerificationToken(ctx, token).
			Return(user, nil).
			Times(1)

		err := service.VerifyEmail(ctx, token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already verified")
		ctrl.Finish()
	})
}
