package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"todo-app/internal/users"
	"todo-app/pkg/locale"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleOauth2 "google.golang.org/api/oauth2/v2"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (LoginResponse, error)
	GoogleLogin(ctx context.Context, state string) string
	GoogleCallback(ctx context.Context, code string) (*LoginResponse, error)
}

type service struct {
	logger          *zap.SugaredLogger
	userRepository  users.Repository
	authRepository  Repository
	validator       *validator.Validate
	jwtSecret       []byte
	tokenExpiration time.Duration
	googleOauth     *oauth2.Config
}

func GetService(
	logger *zap.SugaredLogger,
	userRepo users.Repository,
	authRepo Repository,
	validator *validator.Validate,
) Service {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test"
	}

	// Diagnostic logging to ensure Google credentials are loaded
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if googleClientID == "" || googleClientSecret == "" {
		logger.Warnw("Google OAuth credentials are not set in the environment. Google SSO will fail. Please check your .env file and ensure GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are correct.")
	} else {
		logger.Infow("Google OAuth credentials loaded successfully.")
	}

	googleOauth := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &service{
		logger:          logger,
		userRepository:  userRepo,
		authRepository:  authRepo,
		validator:       validator,
		jwtSecret:       []byte(jwtSecret),
		tokenExpiration: 20 * time.Minute,
		googleOauth:     googleOauth,
	}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	if err := s.validator.Struct(req); err != nil {
		return LoginResponse{}, err
	}

	user, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warnw("user not found", "email", req.Email)
		return LoginResponse{}, errors.New(locale.ErrorInvalidCredentials)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warnw("invalid password", "email", req.Email)
		return LoginResponse{}, errors.New(locale.ErrorInvalidCredentials)
	}

	if user.IsEmailVerified == false {
		s.logger.Warnw("email not verified", "email", req.Email)
		return LoginResponse{}, errors.New(locale.ErrorEmailUnverified)
	}

	expiresAt := time.Now().Add(s.tokenExpiration)
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "todo-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Errorw("failed to sign token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		s.logger.Errorw("failed to generate refresh token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	refreshTokenRecord := RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IsRevoked: false,
	}

	if err := s.authRepository.SaveRefreshToken(ctx, &refreshTokenRecord); err != nil {
		s.logger.Errorw("failed to store refresh token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	return LoginResponse{
		Token:     tokenString,
		Refresh:   refreshToken,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	claims, err := s.ValidateToken(token)
	if err != nil {
		return err
	}

	return s.authRepository.RevokeRefreshTokensByUserID(ctx, claims.UserID)
}

func (s *service) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (LoginResponse, error) {
	// Validate refresh token
	tokenRecord, err := s.authRepository.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return LoginResponse{}, errors.New(locale.ErrorInvalidToken)
	}

	if tokenRecord.IsRevoked || time.Now().After(tokenRecord.ExpiresAt) {
		return LoginResponse{}, errors.New(locale.ErrorInvalidToken)
	}

	user, err := s.userRepository.GetById(ctx, tokenRecord.UserID)
	if err != nil {
		return LoginResponse{}, errors.New(locale.ErrorUserNotFound)
	}

	// Generate new JWT token
	expiresAt := time.Now().Add(s.tokenExpiration)
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "todo-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Errorw("failed to sign token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	return LoginResponse{
		Token:     tokenString,
		Refresh:   refreshToken,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}, nil
}

func (s *service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *service) GoogleLogin(ctx context.Context, state string) string {
	// Generate a random state string for CSRF protection
	b := make([]byte, 16)
	rand.Read(b)
	oauthState := base64.URLEncoding.EncodeToString(b)

	// In a real app, you would store this state in a short-lived cookie or session
	// to verify it in the callback. For this example, we'll just pass it through.

	url := s.googleOauth.AuthCodeURL(oauthState)
	return url
}

func (s *service) GoogleCallback(ctx context.Context, code string) (*LoginResponse, error) {
	token, err := s.googleOauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("could not get token from google: %w", err)
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not get user info from google: %w", err)
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read user info response: %w", err)
	}

	var userInfo googleOauth2.Userinfo
	if err := json.Unmarshal(contents, &userInfo); err != nil {
		return nil, fmt.Errorf("could not unmarshal user info: %w", err)
	}

	// Check if user exists. If GetByEmail returns any error, we assume the user doesn't exist and create one.
	user, err := s.userRepository.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		// User not found, create a new one
		newUser := users.User{
			Email:           userInfo.Email,
			FirstName:       userInfo.GivenName,
			LastName:        userInfo.FamilyName,
			IsEmailVerified: true, // Email from Google is considered verified
		}
		if createErr := s.userRepository.Create(ctx, &newUser); createErr != nil {
			return nil, fmt.Errorf("could not create new user: %w", createErr)
		}
		user = newUser // Assign the newly created user
	}

	// Generate JWT for the user
	expiresAt := time.Now().Add(s.tokenExpiration)
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "todo-app",
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Errorw("failed to sign token", "error", err)
		return nil, errors.New(locale.ErrorInternalServer)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		s.logger.Errorw("failed to generate refresh token", "error", err)
		return nil, errors.New(locale.ErrorInternalServer)
	}

	refreshTokenRecord := RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IsRevoked: false,
	}

	if err := s.authRepository.SaveRefreshToken(ctx, &refreshTokenRecord); err != nil {
		s.logger.Errorw("failed to store refresh token", "error", err)
		return nil, errors.New(locale.ErrorInternalServer)
	}

	return &LoginResponse{
		Token:     tokenString,
		Refresh:   refreshToken,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}, nil
}
