package auth

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (RefreshToken, error)
	RevokeRefreshTokensByUserID(ctx context.Context, userID uint) error
	DeleteExpiredTokens(ctx context.Context) error
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

func (r *repository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		r.logger.Errorw("failed to create refresh token", "error", result.Error)

		return result.Error
	}

	return nil
}

func (r *repository) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	var refreshToken RefreshToken
	result := r.db.WithContext(ctx).Where("token = ? AND is_revoked = ?", token, false).First(&refreshToken)
	if result.Error != nil {
		r.logger.Errorw("failed to find refresh token", "error", result.Error)

		return RefreshToken{}, result.Error
	}

	return refreshToken, nil
}

func (r *repository) RevokeRefreshTokensByUserID(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).Model(&RefreshToken{}).Where("user_id = ?", userID).Update("is_revoked", true)
	if result.Error != nil {
		r.logger.Errorw("failed to revoke refresh tokens", "user_id", userID, "error", result.Error)

		return result.Error
	}

	return nil
}

func (r *repository) DeleteExpiredTokens(ctx context.Context) error {
	result := r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&RefreshToken{})
	if result.Error != nil {
		r.logger.Errorw("failed to delete expired tokens", "error", result.Error)

		return result.Error
	}

	return nil
}
