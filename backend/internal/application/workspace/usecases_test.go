package workspace

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	domainWorkspace "takt/internal/domain/workspace"
)

type mockUserRepo struct {
	usersByID    map[uuid.UUID]*domainWorkspace.User
	usersByEmail map[string]*domainWorkspace.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:    make(map[uuid.UUID]*domainWorkspace.User),
		usersByEmail: make(map[string]*domainWorkspace.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domainWorkspace.User) error {
	m.usersByID[u.ID] = u
	m.usersByEmail[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domainWorkspace.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, domainWorkspace.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainWorkspace.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domainWorkspace.ErrUserNotFound
	}
	return u, nil
}

type mockWorkspaceRepo struct {
	workspacesByID map[uuid.UUID]*domainWorkspace.Workspace
	members        map[string]*domainWorkspace.WorkspaceMember
}

func newMockWorkspaceRepo() *mockWorkspaceRepo {
	return &mockWorkspaceRepo{
		workspacesByID: make(map[uuid.UUID]*domainWorkspace.Workspace),
		members:        make(map[string]*domainWorkspace.WorkspaceMember),
	}
}

func (m *mockWorkspaceRepo) Create(ctx context.Context, ws *domainWorkspace.Workspace) error {
	m.workspacesByID[ws.ID] = ws
	return nil
}

func (m *mockWorkspaceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainWorkspace.Workspace, error) {
	ws, ok := m.workspacesByID[id]
	if !ok {
		return nil, domainWorkspace.ErrWorkspaceNotFound
	}
	return ws, nil
}

func (m *mockWorkspaceRepo) CreateMember(ctx context.Context, member *domainWorkspace.WorkspaceMember) error {
	key := member.WorkspaceID.String() + ":" + member.UserID.String()
	m.members[key] = member
	return nil
}

func (m *mockWorkspaceRepo) GetMember(ctx context.Context, wsID, userID uuid.UUID) (*domainWorkspace.WorkspaceMember, error) {
	key := wsID.String() + ":" + userID.String()
	mem, ok := m.members[key]
	if !ok {
		return nil, domainWorkspace.ErrAccessDenied
	}
	return mem, nil
}

func (m *mockWorkspaceRepo) GetWorkspacesByUserID(ctx context.Context, userID uuid.UUID) ([]*domainWorkspace.UserWorkspace, error) {
	var list []*domainWorkspace.UserWorkspace
	for _, ws := range m.workspacesByID {
		key := ws.ID.String() + ":" + userID.String()
		if mem, ok := m.members[key]; ok {
			list = append(list, &domainWorkspace.UserWorkspace{
				ID:        ws.ID,
				Name:      ws.Name,
				Type:      ws.Type,
				Role:      mem.Role,
				CreatedAt: ws.CreatedAt,
			})
		}
	}
	return list, nil
}

func (m *mockWorkspaceRepo) GetPersonalWorkspaceByUserID(ctx context.Context, userID uuid.UUID) (*domainWorkspace.UserWorkspace, error) {
	list, err := m.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, ws := range list {
		if ws.Type == domainWorkspace.TypePersonal {
			return ws, nil
		}
	}
	return nil, domainWorkspace.ErrWorkspaceNotFound
}

type mockHasher struct{}

func (h *mockHasher) Hash(p string) (string, error) {
	return "hash_" + p, nil
}

func (h *mockHasher) Compare(hashed, p string) error {
	if hashed == "hash_"+p {
		return nil
	}
	return errors.New("mismatch")
}

type mockTokenGen struct{}

func (t *mockTokenGen) GenerateToken(userID, wsID uuid.UUID) (string, error) {
	return "jwt_token_" + userID.String(), nil
}

type mockTasksCreator struct {
	createdWorkspaces []uuid.UUID
}

func (c *mockTasksCreator) CreateInitialInboxTasks(ctx context.Context, wsID uuid.UUID) error {
	c.createdWorkspaces = append(c.createdWorkspaces, wsID)
	return nil
}

func TestService_Register_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	wsRepo := newMockWorkspaceRepo()
	hasher := &mockHasher{}
	tokenGen := &mockTokenGen{}
	tasksCreator := &mockTasksCreator{}

	svc := NewService(userRepo, wsRepo, hasher, tokenGen, tasksCreator, nil)

	out, err := svc.Register(context.Background(), RegisterInput{
		Email:       "pavel@takt.dev",
		Password:    "password123",
		DisplayName: "Pavel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.User.Email != "pavel@takt.dev" {
		t.Errorf("expected email pavel@takt.dev, got %s", out.User.Email)
	}
	if out.Workspace.Name != "Личное" {
		t.Errorf("expected personal workspace 'Личное', got %s", out.Workspace.Name)
	}
	if out.Workspace.Role != domainWorkspace.RoleOwner {
		t.Errorf("expected owner role, got %s", out.Workspace.Role)
	}
	if len(tasksCreator.createdWorkspaces) != 1 {
		t.Errorf("expected initial tasks to be created for 1 workspace, got %d", len(tasksCreator.createdWorkspaces))
	}
	if out.Token == "" {
		t.Errorf("expected non-empty token")
	}
}

func TestService_Register_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	wsRepo := newMockWorkspaceRepo()
	svc := NewService(userRepo, wsRepo, &mockHasher{}, &mockTokenGen{}, &mockTasksCreator{}, nil)

	_, err := svc.Register(context.Background(), RegisterInput{
		Email:       "user@example.com",
		Password:    "secret123",
		DisplayName: "First",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Register(context.Background(), RegisterInput{
		Email:       "user@example.com",
		Password:    "secret123",
		DisplayName: "Second",
	})
	if !errors.Is(err, domainWorkspace.ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestService_Register_ShortPassword(t *testing.T) {
	svc := NewService(newMockUserRepo(), newMockWorkspaceRepo(), &mockHasher{}, &mockTokenGen{}, nil, nil)

	_, err := svc.Register(context.Background(), RegisterInput{
		Email:       "user@example.com",
		Password:    "123",
		DisplayName: "Short Pass",
	})
	if !errors.Is(err, domainWorkspace.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestService_Login_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	wsRepo := newMockWorkspaceRepo()
	svc := NewService(userRepo, wsRepo, &mockHasher{}, &mockTokenGen{}, &mockTasksCreator{}, nil)

	_, err := svc.Register(context.Background(), RegisterInput{
		Email:       "login@takt.dev",
		Password:    "secret123",
		DisplayName: "Login User",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	loginOut, err := svc.Login(context.Background(), LoginInput{
		Email:    "login@takt.dev",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}

	if loginOut.Token == "" {
		t.Errorf("expected token on login")
	}
	if loginOut.Workspace.Name != "Личное" {
		t.Errorf("expected workspace 'Личное', got %s", loginOut.Workspace.Name)
	}
}

func TestService_Login_InvalidPassword(t *testing.T) {
	userRepo := newMockUserRepo()
	wsRepo := newMockWorkspaceRepo()
	svc := NewService(userRepo, wsRepo, &mockHasher{}, &mockTokenGen{}, nil, nil)

	_, _ = svc.Register(context.Background(), RegisterInput{
		Email:       "test@takt.dev",
		Password:    "correct_password",
		DisplayName: "Test",
	})

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "test@takt.dev",
		Password: "wrong_password",
	})
	if !errors.Is(err, domainWorkspace.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
