package auth

import (
	"net/http"
	"strings"
	e "todo-app/pkg/errors"
	"todo-app/pkg/handlers"
	"todo-app/pkg/locale"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type endpointHandler struct {
	logger  *zap.SugaredLogger
	service Service
	e       *echo.Echo
}

func GetEndpointHandler(
	logger *zap.SugaredLogger,
	service Service,
	e *echo.Echo,
) handlers.EndpointHandler {
	return &endpointHandler{
		logger:  logger,
		service: service,
		e:       e,
	}
}

func (h *endpointHandler) AddEndpoints() {
	var endpoints = []handlers.Endpoint{
		{
			Method:  http.MethodPost,
			Path:    "/login",
			Handler: h.login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/logout",
			Handler: h.logout,
		},
		{
			Method:  http.MethodPost,
			Path:    "/refresh",
			Handler: h.refresh,
		},
	}

	for _, endpoint := range endpoints {
		handlers.Method(h.e, endpoint.Method, endpoint.Path, endpoint.Handler)
	}
}

// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @ID login
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} errors.ResponseError "Bad Request"
// @Failure 401 {object} errors.ResponseError "Unauthorized"
// @Router /login [post]
func (h *endpointHandler) login(ctx echo.Context) error {
	h.logger.Infow("user login attempt...")

	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		h.logger.Warnw("could not bind login request", "error", err.Error())
		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidBody})
	}

	response, err := h.service.Login(ctx.Request().Context(), req)
	if err != nil {
		h.logger.Warnw("login failed", "error", err.Error())
		return ctx.JSON(http.StatusUnauthorized, e.ResponseError{Message: err.Error()})
	}

	h.logger.Infow("user logged in successfully", "user_id", response.User.ID)
	return ctx.JSON(http.StatusOK, response)
}

// @Summary User logout
// @Description Logout user and revoke refresh tokens
// @Tags auth
// @ID logout
// @Security BearerAuth
// @Produce json
// @Success 200 {string} string "Successfully logged out"
// @Failure 400 {object} errors.ResponseError "Bad Request"
// @Failure 401 {object} errors.ResponseError "Unauthorized"
// @Router /logout [post]
func (h *endpointHandler) logout(ctx echo.Context) error {
	h.logger.Infow("user logout...")

	// Extract token from Authorization header
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ctx.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorMissingToken})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return ctx.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorInvalidToken})
	}

	err := h.service.Logout(ctx.Request().Context(), tokenString)
	if err != nil {
		h.logger.Warnw("logout failed", "error", err.Error())
		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: err.Error()})
	}

	h.logger.Infow("user logged out successfully")
	return ctx.JSON(http.StatusOK, map[string]string{"message": "Successfully logged out"})
}

// @Summary Refresh JWT token
// @Description Refresh JWT token using refresh token
// @Tags auth
// @ID refresh
// @Accept json
// @Produce json
// @Param refresh_token body map[string]string true "Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} errors.ResponseError "Bad Request"
// @Failure 401 {object} errors.ResponseError "Unauthorized"
// @Router /refresh [post]
func (h *endpointHandler) refresh(ctx echo.Context) error {
	h.logger.Infow("token refresh attempt...")

	var req map[string]string
	if err := ctx.Bind(&req); err != nil {
		h.logger.Warnw("could not bind refresh request", "error", err.Error())
		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidBody})
	}

	refreshToken, exists := req["refresh_token"]
	if !exists || refreshToken == "" {
		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorMissingRefreshToken})
	}

	response, err := h.service.RefreshToken(ctx.Request().Context(), refreshToken)
	if err != nil {
		h.logger.Warnw("token refresh failed", "error", err.Error())
		return ctx.JSON(http.StatusUnauthorized, e.ResponseError{Message: err.Error()})
	}

	h.logger.Infow("token refreshed successfully")
	return ctx.JSON(http.StatusOK, response)
}
