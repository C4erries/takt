package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	workspaceApp "takt/internal/application/workspace"
	workspaceMocks "takt/internal/application/workspace/mocks"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

func setupTestContext(method, path string, body []byte, userID, wsID uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if userID != uuid.Nil {
		c.Set(string(UserIDContextKey), userID)
		reqCtx := context.WithValue(req.Context(), UserIDContextKey, userID)
		if wsID != uuid.Nil {
			c.Set(string(WorkspaceIDContextKey), wsID)
			reqCtx = context.WithValue(reqCtx, WorkspaceIDContextKey, wsID)
		}
		c.SetRequest(req.WithContext(reqCtx))
	}

	return c, rec
}

func TestHandler_ListWorkspaces(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		expected := []*domainWorkspace.UserWorkspace{
			{ID: wsID, OwnerID: userID, Name: "Workspace 1", Role: domainWorkspace.RoleOwner, CreatedAt: time.Now().UTC()},
		}
		mockSvc.EXPECT().ListWorkspaces(mock.Anything, userID).Return(expected, nil).Once()

		c, rec := setupTestContext(http.MethodGet, "/api/v1/workspaces", nil, userID, wsID)
		err := h.ListWorkspaces(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []generated.Workspace
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Len(t, resp, 1)
		assert.Equal(t, "Workspace 1", resp[0].Name)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		c, rec := setupTestContext(http.MethodGet, "/api/v1/workspaces", nil, uuid.Nil, uuid.Nil)
		err := h.ListWorkspaces(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestHandler_CreateWorkspace(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		created := &domainWorkspace.UserWorkspace{
			ID: wsID, OwnerID: userID, Name: "New WS", Role: domainWorkspace.RoleOwner, CreatedAt: time.Now().UTC(),
		}
		mockSvc.EXPECT().CreateWorkspace(mock.Anything, userID, "New WS").Return(created, nil).Once()

		c, rec := setupTestContext(http.MethodPost, "/api/v1/workspaces", []byte(`{"name":"New WS"}`), userID, wsID)
		err := h.CreateWorkspace(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().CreateWorkspace(mock.Anything, userID, "").Return(nil, domainWorkspace.ErrInvalidWorkspaceName).Once()

		c, rec := setupTestContext(http.MethodPost, "/api/v1/workspaces", []byte(`{"name":""}`), userID, wsID)
		err := h.CreateWorkspace(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandler_UpdateWorkspace(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		updated := &domainWorkspace.UserWorkspace{
			ID: wsID, OwnerID: userID, Name: "Renamed", Role: domainWorkspace.RoleOwner, CreatedAt: time.Now().UTC(),
		}
		mockSvc.EXPECT().UpdateWorkspace(mock.Anything, userID, wsID, "Renamed").Return(updated, nil).Once()

		c, rec := setupTestContext(http.MethodPatch, "/api/v1/workspaces/"+wsID.String(), []byte(`{"name":"Renamed"}`), userID, wsID)
		err := h.UpdateWorkspace(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().UpdateWorkspace(mock.Anything, userID, wsID, "Renamed").Return(nil, domainWorkspace.ErrAccessDenied).Once()

		c, rec := setupTestContext(http.MethodPatch, "/api/v1/workspaces/"+wsID.String(), []byte(`{"name":"Renamed"}`), userID, wsID)
		err := h.UpdateWorkspace(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestHandler_SwitchWorkspace(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		out := &workspaceApp.SwitchWorkspaceOutput{
			Token: "new-jwt-token",
			Workspace: &domainWorkspace.UserWorkspace{
				ID: wsID, OwnerID: userID, Name: "Target WS", Role: domainWorkspace.RoleMember, CreatedAt: time.Now().UTC(),
			},
		}
		mockSvc.EXPECT().SwitchWorkspace(mock.Anything, userID, wsID).Return(out, nil).Once()

		c, rec := setupTestContext(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/switch", nil, userID, wsID)
		err := h.SwitchWorkspace(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp generated.WorkspaceSwitchResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "new-jwt-token", resp.Token)
		assert.Equal(t, wsID, resp.Workspace.Id)
	})
}

func TestHandler_ListWorkspaceMembers(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		expected := []*workspaceApp.WorkspaceMemberInfo{
			{ID: uuid.New(), UserID: userID, Email: "user@takt.dev", DisplayName: "Pavel", Role: domainWorkspace.RoleOwner, CreatedAt: time.Now().UTC()},
		}
		mockSvc.EXPECT().ListMembers(mock.Anything, userID, wsID).Return(expected, nil).Once()

		c, rec := setupTestContext(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/members", nil, userID, wsID)
		err := h.ListWorkspaceMembers(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []generated.WorkspaceMember
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Len(t, resp, 1)
		assert.Equal(t, "Pavel", resp[0].DisplayName)
	})
}

func TestHandler_RemoveWorkspaceMember(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()
	targetID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().RemoveMember(mock.Anything, userID, wsID, targetID).Return(nil).Once()

		c, rec := setupTestContext(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/members/"+targetID.String(), nil, userID, wsID)
		err := h.RemoveWorkspaceMember(c, wsID, targetID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("cannot remove owner", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().RemoveMember(mock.Anything, userID, wsID, targetID).Return(domainWorkspace.ErrCannotRemoveOwner).Once()

		c, rec := setupTestContext(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/members/"+targetID.String(), nil, userID, wsID)
		err := h.RemoveWorkspaceMember(c, wsID, targetID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandler_CreateWorkspaceInvite(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		future := time.Now().UTC().Add(7 * 24 * time.Hour)
		invite := &domainWorkspace.WorkspaceInvite{
			ID:          uuid.New(),
			WorkspaceID: wsID,
			InvitedByID: userID,
			Token:       "test-invite-token",
			ExpiresAt:   &future,
			CreatedAt:   time.Now().UTC(),
		}
		mockSvc.EXPECT().CreateInvite(mock.Anything, userID, wsID).Return(invite, nil).Once()

		c, rec := setupTestContext(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/invites", nil, userID, wsID)
		err := h.CreateWorkspaceInvite(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp generated.WorkspaceInvite
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "test-invite-token", resp.Token)
		assert.Equal(t, "/invite/test-invite-token", resp.InviteUrl)
	})
}

func TestHandler_GetInvitePreview(t *testing.T) {
	t.Run("success public endpoint", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		preview := &workspaceApp.InvitePreviewInfo{
			Token:         "valid-token",
			WorkspaceName: "Team",
			InvitedByName: "Pavel",
		}
		mockSvc.EXPECT().GetInvitePreview(mock.Anything, "valid-token").Return(preview, nil).Once()

		c, rec := setupTestContext(http.MethodGet, "/api/v1/invites/valid-token", nil, uuid.Nil, uuid.Nil)
		err := h.GetInvitePreview(c, "valid-token")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp generated.InvitePreview
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Team", resp.WorkspaceName)
		assert.Equal(t, "Pavel", resp.InvitedByName)
	})

	t.Run("expired invite", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().GetInvitePreview(mock.Anything, "expired-token").Return(nil, domainWorkspace.ErrInviteExpired).Once()

		c, rec := setupTestContext(http.MethodGet, "/api/v1/invites/expired-token", nil, uuid.Nil, uuid.Nil)
		err := h.GetInvitePreview(c, "expired-token")
		require.NoError(t, err)
		assert.Equal(t, http.StatusGone, rec.Code)
	})
}

func TestHandler_AcceptInvite(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		out := &workspaceApp.SwitchWorkspaceOutput{
			Token: "accepted-token",
			Workspace: &domainWorkspace.UserWorkspace{
				ID: wsID, OwnerID: uuid.New(), Name: "Joined WS", Role: domainWorkspace.RoleMember, CreatedAt: time.Now().UTC(),
			},
		}
		mockSvc.EXPECT().AcceptInvite(mock.Anything, userID, "valid-token").Return(out, nil).Once()

		c, rec := setupTestContext(http.MethodPost, "/api/v1/invites/valid-token/accept", nil, userID, wsID)
		err := h.AcceptInvite(c, "valid-token")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp generated.WorkspaceSwitchResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "accepted-token", resp.Token)
	})

	t.Run("already a member", func(t *testing.T) {
		mockSvc := workspaceMocks.NewMockUseCases(t)
		h := NewHandler(nil, "0.4.1", false, mockSvc, nil, nil, nil)

		mockSvc.EXPECT().AcceptInvite(mock.Anything, userID, "token123").Return(nil, domainWorkspace.ErrAlreadyMember).Once()

		c, rec := setupTestContext(http.MethodPost, "/api/v1/invites/token123/accept", nil, userID, wsID)
		err := h.AcceptInvite(c, "token123")
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
