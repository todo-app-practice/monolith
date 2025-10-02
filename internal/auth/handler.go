package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
			Path:    "/auth/login",
			Handler: h.login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/logout",
			Handler: h.logout,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/refresh",
			Handler: h.refresh,
		},
		{
			Method:  http.MethodGet,
			Path:    "/auth/google/login",
			Handler: h.googleLogin,
		},
		{
			Method:  http.MethodGet,
			Path:    "/auth/google/callback",
			Handler: h.googleCallback,
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

// @Summary Google OAuth login
// @Description Redirects user to Google OAuth for authentication
// @Tags auth
// @ID google-login
// @Produce json
// @Success 302 {string} string "Redirect to Google OAuth"
// @Router /auth/google/login [get]
func (h *endpointHandler) googleLogin(ctx echo.Context) error {
	url := h.service.GoogleLogin(ctx.Request().Context(), "state-string") // State should be random
	return ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// @Summary Google OAuth callback
// @Description Handles the callback from Google OAuth and redirects to frontend with tokens
// @Tags auth
// @ID google-callback
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string false "State parameter for CSRF protection"
// @Success 302 {string} string "Redirect to frontend with tokens"
// @Failure 400 {string} string "Invalid authorization code"
// @Router /auth/google/callback [get]
func (h *endpointHandler) googleCallback(ctx echo.Context) error {
	code := ctx.QueryParam("code")
	// You should also verify the 'state' query param here against the one you stored
	response, err := h.service.GoogleCallback(ctx.Request().Context(), code)
	if err != nil {
		h.logger.Warnw("google callback failed", "error", err.Error())
		// Redirect to a frontend error page
		return ctx.Redirect(http.StatusTemporaryRedirect, "/login?error=google-failed")
	}

	// On success, we need to send the tokens and user data to the frontend.
	userJSON, err := json.Marshal(response.User)
	if err != nil {
		h.logger.Errorw("failed to marshal user data", "error", err)
		return ctx.Redirect(http.StatusTemporaryRedirect, "/login?error=internal-error")
	}
	userBase64 := base64.URLEncoding.EncodeToString(userJSON)

	frontendURL := "http://local.todo.com" // This should be an env var in a real app
	redirectURL := fmt.Sprintf(
		"%s/login/success?token=%s&refresh=%s&user=%s",
		frontendURL,
		response.Token,
		response.Refresh,
		userBase64,
	)
	return ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
