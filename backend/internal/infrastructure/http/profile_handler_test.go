package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	workspaceMocks "takt/internal/application/workspace/mocks"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

func TestHandler_UpdateUserProfile(t *testing.T) {
	userID := uuid.New()
	newName := "New Display Name"

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		user := &domainWorkspace.User{
			ID:          userID,
			Email:       "user@takt.dev",
			DisplayName: newName,
			CreatedAt:   time.Now().UTC(),
		}
		mockSvc.EXPECT().UpdateUserProfile(mock.Anything, userID, &newName).Return(user, nil).Once()

		body := []byte(`{"displayName":"New Display Name"}`)
		c, rec := setupTestContext(http.MethodPatch, "/api/v1/users/me", body, userID, uuid.Nil)
		err := h.UpdateUserProfile(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp generated.User
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, newName, *resp.DisplayName)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		c, rec := setupTestContext(http.MethodPatch, "/api/v1/users/me", []byte(`{}`), uuid.Nil, uuid.Nil)
		err := h.UpdateUserProfile(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("validation error on invalid display name", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		invalidName := ""
		mockSvc.EXPECT().UpdateUserProfile(mock.Anything, userID, &invalidName).
			Return(nil, domainWorkspace.ErrInvalidDisplayName).Once()

		body := []byte(`{"displayName":""}`)
		c, rec := setupTestContext(http.MethodPatch, "/api/v1/users/me", body, userID, uuid.Nil)
		err := h.UpdateUserProfile(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().UpdateUserProfile(mock.Anything, userID, &newName).
			Return(nil, domainWorkspace.ErrUserNotFound).Once()

		body := []byte(`{"displayName":"New Display Name"}`)
		c, rec := setupTestContext(http.MethodPatch, "/api/v1/users/me", body, userID, uuid.Nil)
		err := h.UpdateUserProfile(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandler_UploadUserAvatar(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		avatarURL := "/api/v1/users/me/avatar/avatars/user.png"
		user := &domainWorkspace.User{
			ID:          userID,
			Email:       "user@takt.dev",
			DisplayName: "Pavel",
			AvatarURL:   &avatarURL,
			CreatedAt:   time.Now().UTC(),
		}

		mockSvc.EXPECT().UploadUserAvatar(mock.Anything, userID, "avatar.png", "image/png", mock.Anything, int64(4)).
			Return(user, nil).Once()

		// Create multipart form data with explicit Content-Type
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Disposition", `form-data; name="avatar"; filename="avatar.png"`)
		partHeader.Set("Content-Type", "image/png")
		part, err := writer.CreatePart(partHeader)
		require.NoError(t, err)
		_, err = part.Write([]byte("fake"))
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		reqCtx := context.WithValue(req.Context(), UserIDContextKey, userID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req.WithContext(reqCtx), rec)
		c.Set(string(UserIDContextKey), userID)

		err = h.UploadUserAvatar(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp generated.User
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, avatarURL, *resp.AvatarUrl)
	})

	t.Run("missing avatar form field", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		require.NoError(t, writer.Close())

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		reqCtx := context.WithValue(req.Context(), UserIDContextKey, userID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req.WithContext(reqCtx), rec)
		c.Set(string(UserIDContextKey), userID)

		err := h.UploadUserAvatar(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.9.7", false, mockSvc, nil, nil, nil)

		c, rec := setupTestContext(http.MethodPost, "/api/v1/users/me/avatar", nil, uuid.Nil, uuid.Nil)
		err := h.UploadUserAvatar(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
