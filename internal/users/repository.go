package users

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
	GetById(ctx context.Context, id uint) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
}

type repository struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func GetRepository(logger *zap.SugaredLogger, db *gorm.DB) Repository {
	return &repository{
		logger: logger,
		db:     db,
	}
}

func (r *repository) GetById(ctx context.Context, id uint) (User, error) {
	var user User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		r.logger.Errorw("failed to find user by id", "id", id, "error", result.Error)

		return User{}, result.Error
	}

	return user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (User, error) {
	var user User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		r.logger.Errorw("failed to find user by email", "email", email, "error", result.Error)

		return User{}, result.Error
	}

	return user, nil
}

func (r *repository) Create(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		r.logger.Errorw("failed to create user", "error", result.Error)

		return result.Error
	}

	return nil
}

func (r *repository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		r.logger.Errorw("failed to update user", "id", id, "error", result.Error)

		return result.Error
	}

	return nil
}
