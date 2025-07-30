package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"todo-app/pkg/locale"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	localErr "todo-app/pkg/errors"
)

func TestHandler_Create(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("success create user", func(t *testing.T) {
		userData := `{"firstName":"John","lastName":"Doe","email":"john.doe@example.com","password":"password123"}`
		user := User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
		}

		req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			Create(ctx.Request().Context(), &user).
			Return(nil).
			Times(1)

		if assert.NoError(t, h.create(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var responseUser User
			err := json.Unmarshal(rec.Body.Bytes(), &responseUser)
			assert.NoError(t, err)

			assert.Equal(t, user.FirstName, responseUser.FirstName)
			assert.Equal(t, user.LastName, responseUser.LastName)
			assert.Equal(t, user.Email, responseUser.Email)
			assert.Equal(t, "", responseUser.Password) // Password should be cleared
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		userData := `{"firstName":123,"lastName":456,"email":"invalid","password":789}`

		req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, h.create(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorCouldNotReadUser, responseError.Message)
		}
	})

	t.Run("malformed json body", func(t *testing.T) {
		userData := `{invalid json}`

		req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, h.create(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorCouldNotReadUser, responseError.Message)
		}
	})

	t.Run("service error", func(t *testing.T) {
		userData := `{"firstName":"John","lastName":"Doe","email":"john.doe@example.com","password":"password123"}`
		user := User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
		}

		req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			Create(ctx.Request().Context(), &user).
			Return(errors.New("email already exists")).
			Times(1)

		if assert.NoError(t, h.create(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidUser, responseError.Message)
			assert.Equal(t, "email already exists", responseError.Details)
		}
	})

	t.Run("empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		user := User{} // Empty user struct expected

		mockService.
			EXPECT().
			Create(ctx.Request().Context(), &user).
			Return(errors.New("validation failed")).
			Times(1)

		if assert.NoError(t, h.create(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidUser, responseError.Message)
		}
	})
}

func TestHandler_Update(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("success update user", func(t *testing.T) {
		userData := `{"firstName":"Jane","lastName":"Smith","email":"jane.smith@example.com"}`
		inputUser := User{
			ID:        1,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
		}
		updatedUser := User{
			ID:        1,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
			Password:  "hashedpassword",
		}

		req := httptest.NewRequest(http.MethodPut, "/user/1", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/user/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			Update(ctx.Request().Context(), &inputUser).
			Return(updatedUser, nil).
			Times(1)

		if assert.NoError(t, h.update(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var responseUser User
			err := json.Unmarshal(rec.Body.Bytes(), &responseUser)
			assert.NoError(t, err)

			assert.Equal(t, updatedUser.ID, responseUser.ID)
			assert.Equal(t, updatedUser.FirstName, responseUser.FirstName)
			assert.Equal(t, updatedUser.LastName, responseUser.LastName)
			assert.Equal(t, updatedUser.Email, responseUser.Email)
			assert.Equal(t, "", responseUser.Password) // Password should be cleared
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		userData := `{"firstName":"Jane","lastName":"Smith","email":"jane.smith@example.com"}`

		req := httptest.NewRequest(http.MethodPut, "/user/abc", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/user/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("abc")

		if assert.NoError(t, h.update(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidID, responseError.Message)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		userData := `{"firstName":123,"lastName":456}`

		req := httptest.NewRequest(http.MethodPut, "/user/1", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/user/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		if assert.NoError(t, h.update(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorCouldNotReadUser, responseError.Message)
		}
	})

	t.Run("service error", func(t *testing.T) {
		userData := `{"firstName":"Jane","lastName":"Smith","email":"jane.smith@example.com"}`
		inputUser := User{
			ID:        1,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
		}

		req := httptest.NewRequest(http.MethodPut, "/user/1", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/user/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			Update(ctx.Request().Context(), &inputUser).
			Return(User{}, errors.New("user not found")).
			Times(1)

		if assert.NoError(t, h.update(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, "user not found", responseError.Message)
		}
	})

	t.Run("malformed json body", func(t *testing.T) {
		userData := `{invalid json}`

		req := httptest.NewRequest(http.MethodPut, "/user/1", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/user/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		if assert.NoError(t, h.update(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorCouldNotReadUser, responseError.Message)
		}
	})

	t.Run("zero id", func(t *testing.T) {
		userData := `{"firstName":"Jane","lastName":"Smith","email":"jane.smith@example.com"}`

		req := httptest.NewRequest(http.MethodPut, "/user/0", strings.NewReader(userData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/user/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("0")

		inputUser := User{
			ID:        0,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
		}

		mockService.
			EXPECT().
			Update(ctx.Request().Context(), &inputUser).
			Return(User{}, errors.New("invalid user id")).
			Times(1)

		if assert.NoError(t, h.update(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, "invalid user id", responseError.Message)
		}
	})
}

func TestHandler_VerifyEmail(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("success verify email", func(t *testing.T) {
		token := "valid-verification-token"

		req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			VerifyEmail(ctx.Request().Context(), token).
			Return(nil).
			Times(1)

		if assert.NoError(t, h.verifyEmail(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "Email Verified!")
			assert.Contains(t, rec.Body.String(), "Email verified successfully")
			assert.Contains(t, rec.Body.String(), "✅")
		}
	})

	t.Run("missing verification token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/verify-email", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, h.verifyEmail(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "Error")
			assert.Contains(t, rec.Body.String(), "Verification token is required")
			assert.Contains(t, rec.Body.String(), "❌")
		}
	})

	t.Run("empty verification token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/verify-email?token=", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, h.verifyEmail(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "Error")
			assert.Contains(t, rec.Body.String(), "Verification token is required")
			assert.Contains(t, rec.Body.String(), "❌")
		}
	})

	t.Run("invalid verification token", func(t *testing.T) {
		token := "invalid-token"

		req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			VerifyEmail(ctx.Request().Context(), token).
			Return(errors.New("token is invalid or expired")).
			Times(1)

		if assert.NoError(t, h.verifyEmail(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "Verification Failed")
			assert.Contains(t, rec.Body.String(), "token is invalid or expired")
			assert.Contains(t, rec.Body.String(), "❌")
		}
	})

	t.Run("service error", func(t *testing.T) {
		token := "some-token"

		req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			VerifyEmail(ctx.Request().Context(), token).
			Return(errors.New("database connection failed")).
			Times(1)

		if assert.NoError(t, h.verifyEmail(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "Verification Failed")
			assert.Contains(t, rec.Body.String(), "database connection failed")
			assert.Contains(t, rec.Body.String(), "❌")
		}
	})

	t.Run("token with special characters", func(t *testing.T) {
		token := "token-with-special-chars-123-!@#"

		req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			VerifyEmail(ctx.Request().Context(), token).
			Return(nil).
			Times(1)

		if assert.NoError(t, h.verifyEmail(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Body.String(), "Email Verified!")
			assert.Contains(t, rec.Body.String(), "Email verified successfully")
			assert.Contains(t, rec.Body.String(), "✅")
		}
	})
}
