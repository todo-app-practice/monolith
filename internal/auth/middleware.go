package auth

import (
	"net/http"
	"strings"
	e "todo-app/pkg/errors"
	"todo-app/pkg/locale"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// JWTMiddleware creates a middleware function for JWT authentication
func JWTMiddleware(authService Service, logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Warnw("missing authorization header")
				return c.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorMissingToken})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Warnw("invalid authorization header format")
				return c.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorInvalidToken})
			}

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				logger.Warnw("invalid token", "error", err.Error())
				return c.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorInvalidToken})
			}

			// Store user information in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)

			return next(c)
		}
	}
}

// GetUserIDFromContext extracts user ID from Echo context
func GetUserIDFromContext(c echo.Context) uint {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return 0
	}
	return userID
}

// GetUserEmailFromContext extracts user email from Echo context
func GetUserEmailFromContext(c echo.Context) string {
	email, ok := c.Get("user_email").(string)
	if !ok {
		return ""
	}
	return email
}
