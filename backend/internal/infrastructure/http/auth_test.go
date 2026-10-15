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
	workspaceApp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

type mockWorkspaceService struct {
	registerFn func(ctx context.Context, input workspaceApp.RegisterInput) (*workspaceApp.RegisterOutput, error)
	loginFn    func(ctx context.Context, input workspaceApp.LoginInput) (*workspaceApp.LoginOutput, error)
	getMeFn    func(ctx context.Context, userID, workspaceID uuid.UUID) (*workspaceApp.GetProfileOutput, error)
}

func (m *mockWorkspaceService) Register(ctx context.Context, input workspaceApp.RegisterInput) (*workspaceApp.RegisterOutput, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, input)
	}
	return nil, nil
}

func (m *mockWorkspaceService) Login(ctx context.Context, input workspaceApp.LoginInput) (*workspaceApp.LoginOutput, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, input)
	}
	return nil, nil
}

func (m *mockWorkspaceService) GetMe(ctx context.Context, userID, workspaceID uuid.UUID) (*workspaceApp.GetProfileOutput, error) {
	if m.getMeFn != nil {
		return m.getMeFn(ctx, userID, workspaceID)
	}
	return nil, nil
}

func (m *mockWorkspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]*domainWorkspace.UserWorkspace, error) {
	return nil, nil
}

func (m *mockWorkspaceService) CreateWorkspace(ctx context.Context, userID uuid.UUID, name string) (*domainWorkspace.UserWorkspace, error) {
	return nil, nil
}

func (m *mockWorkspaceService) UpdateWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, name string) (*domainWorkspace.UserWorkspace, error) {
	return nil, nil
}

func (m *mockWorkspaceService) SwitchWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*workspaceApp.SwitchWorkspaceOutput, error) {
	return nil, nil
}

func (m *mockWorkspaceService) ListMembers(ctx context.Context, userID, workspaceID uuid.UUID) ([]*workspaceApp.WorkspaceMemberInfo, error) {
	return nil, nil
}

func (m *mockWorkspaceService) RemoveMember(ctx context.Context, actorID, workspaceID, targetUserID uuid.UUID) error {
	return nil
}

func (m *mockWorkspaceService) CreateInvite(ctx context.Context, actorID, workspaceID uuid.UUID) (*domainWorkspace.WorkspaceInvite, error) {
	return nil, nil
}

func (m *mockWorkspaceService) GetInvitePreview(ctx context.Context, token string) (*workspaceApp.InvitePreviewInfo, error) {
	return nil, nil
}

func (m *mockWorkspaceService) AcceptInvite(ctx context.Context, userID uuid.UUID, token string) (*workspaceApp.SwitchWorkspaceOutput, error) {
	return nil, nil
}

func TestJWTManager_GenerateAndParse(t *testing.T) {
	jwtManager := NewJWTManager("super-secret-key-12345678901234567")
	userID := uuid.New()
	wsID := uuid.New()

	tokenStr, err := jwtManager.GenerateToken(userID, wsID)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := jwtManager.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.WorkspaceID != wsID {
		t.Errorf("expected workspaceID %v, got %v", wsID, claims.WorkspaceID)
	}
}

func TestAuthMiddleware_CookieAndBearer(t *testing.T) {
	jwtManager := NewJWTManager("super-secret-key-12345678901234567")
	userID := uuid.New()
	wsID := uuid.New()
	tokenStr, _ := jwtManager.GenerateToken(userID, wsID)

	e := echo.New()
	middleware := AuthMiddleware(jwtManager)

	// Test 1: Missing credentials
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/inbox", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handlerCalled := false
	err := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized || handlerCalled {
		t.Errorf("expected 401 Unauthorized for missing token")
	}

	// Test 2: With Cookie
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/inbox", nil)
	req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: tokenStr})
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	_ = middleware(func(c echo.Context) error {
		handlerCalled = true
		uID, _ := GetUserIDFromContext(c)
		wID, _ := GetWorkspaceIDFromContext(c)
		if uID != userID || wID != wsID {
			t.Errorf("context IDs do not match claims")
		}
		return c.NoContent(http.StatusOK)
	})(c)
	if !handlerCalled {
		t.Errorf("expected handler to be called with valid cookie")
	}

	// Test 3: With Bearer Header
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/inbox", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	_ = middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})(c)
	if !handlerCalled {
		t.Errorf("expected handler to be called with Bearer header")
	}
}

func TestHandler_Register_Success(t *testing.T) {
	jwtManager := NewJWTManager("secret")
	userID := uuid.New()
	wsID := uuid.New()

	mockSvc := &mockWorkspaceService{
		registerFn: func(ctx context.Context, input workspaceApp.RegisterInput) (*workspaceApp.RegisterOutput, error) {
			return &workspaceApp.RegisterOutput{
				Token: "sample-token",
				User: &domainWorkspace.User{
					ID:          userID,
					Email:       input.Email,
					DisplayName: input.DisplayName,
					CreatedAt:   time.Now().UTC(),
				},
				Workspace: &domainWorkspace.UserWorkspace{
					ID:        wsID,
					Name:      "Личное",
					Type:      domainWorkspace.TypePersonal,
					Role:      domainWorkspace.RoleOwner,
					CreatedAt: time.Now().UTC(),
				},
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, mockSvc, nil, nil, jwtManager)

	payload := `{"email":"test@takt.dev","password":"password123","displayName":"Pavel"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rec.Code)
	}

	var resp generated.AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Token != "sample-token" || string(resp.User.Email) != "test@takt.dev" {
		t.Errorf("unexpected response content")
	}

	// Verify Set-Cookie header is present
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != AuthCookieName {
		t.Errorf("expected Set-Cookie header for %s", AuthCookieName)
	}
}
