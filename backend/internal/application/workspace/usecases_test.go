package workspace_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	workspaceapp "takt/internal/application/workspace"
	"takt/internal/application/workspace/mocks"
	domainWorkspace "takt/internal/domain/workspace"
)

var errTestDB = errors.New("database connection failed")

type mockDeps struct {
	userRepo     *mocks.MockUserRepository
	wsRepo       *mocks.MockWorkspaceRepository
	hasher       *mocks.MockPasswordHasher
	tokenGen     *mocks.MockTokenGenerator
	tasksCreator *mocks.MockInitialTasksCreator
	transactor   *mocks.MockTransactor
}

func newMockDeps(t *testing.T) mockDeps {
	return mockDeps{
		userRepo:     mocks.NewMockUserRepository(t),
		wsRepo:       mocks.NewMockWorkspaceRepository(t),
		hasher:       mocks.NewMockPasswordHasher(t),
		tokenGen:     mocks.NewMockTokenGenerator(t),
		tasksCreator: mocks.NewMockInitialTasksCreator(t),
		transactor:   mocks.NewMockTransactor(t),
	}
}

func (d mockDeps) newService() *workspaceapp.Service {
	return workspaceapp.NewService(
		d.userRepo, d.wsRepo, d.hasher, d.tokenGen, d.tasksCreator, d.transactor,
	)
}

func TestService_Register(t *testing.T) {
	testCases := []struct {
		name      string
		input     workspaceapp.RegisterInput
		setupMock func(d mockDeps)
		checkOut  func(t *testing.T, out *workspaceapp.RegisterOutput)
		wantErr   error
	}{
		{
			name:  "успешная регистрация пользователя и создание личного воркспейса",
			input: workspaceapp.RegisterInput{Email: "pavel@takt.dev", Password: "password123", DisplayName: "Pavel"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "pavel@takt.dev").Return(nil, domainWorkspace.ErrUserNotFound).Once()
				d.hasher.EXPECT().Hash("password123").Return("hash_password123", nil).Once()
				d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				d.userRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *domainWorkspace.User) bool {
					return u.Email == "pavel@takt.dev" && u.DisplayName == "Pavel" && u.PasswordHash == "hash_password123"
				})).Return(nil).Once()
				d.wsRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ws *domainWorkspace.Workspace) bool {
					return ws.Name == "Личное" && ws.Type == domainWorkspace.TypePersonal
				})).Return(nil).Once()
				d.wsRepo.EXPECT().CreateMember(mock.Anything, mock.MatchedBy(func(m *domainWorkspace.WorkspaceMember) bool {
					return m.Role == domainWorkspace.RoleOwner
				})).Return(nil).Once()
				d.tasksCreator.EXPECT().CreateInitialInboxTasks(mock.Anything, mock.Anything).Return(nil).Once()
				d.tokenGen.EXPECT().GenerateToken(mock.Anything, mock.Anything).Return("jwt_register_token", nil).Once()
			},
			checkOut: func(t *testing.T, out *workspaceapp.RegisterOutput) {
				assert.Equal(t, "jwt_register_token", out.Token)
				assert.Equal(t, "pavel@takt.dev", out.User.Email)
				assert.Equal(t, "Личное", out.Workspace.Name)
				assert.Equal(t, domainWorkspace.RoleOwner, out.Workspace.Role)
				assert.Equal(t, domainWorkspace.TypePersonal, out.Workspace.Type)
			},
		},
		{
			name:  "ошибка: дубликат email",
			input: workspaceapp.RegisterInput{Email: "user@example.com", Password: "secret123", DisplayName: "Second"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "user@example.com").Return(&domainWorkspace.User{ID: uuid.New(), Email: "user@example.com"}, nil).Once()
			},
			wantErr: domainWorkspace.ErrUserAlreadyExists,
		},
		{
			name:      "ошибка: короткий пароль",
			input:     workspaceapp.RegisterInput{Email: "user@example.com", Password: "123", DisplayName: "Short Pass"},
			setupMock: func(d mockDeps) {},
			wantErr:   domainWorkspace.ErrInvalidPassword,
		},
		{
			name:  "ошибка: сбой транзакции при создании воркспейса",
			input: workspaceapp.RegisterInput{Email: "error@takt.dev", Password: "password123", DisplayName: "Error User"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "error@takt.dev").Return(nil, domainWorkspace.ErrUserNotFound).Once()
				d.hasher.EXPECT().Hash("password123").Return("hash_pass", nil).Once()
				d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				d.userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newMockDeps(t)
			tc.setupMock(deps)

			svc := deps.newService()
			out, err := svc.Register(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, out)
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				if tc.checkOut != nil {
					tc.checkOut(t, out)
				}
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	userID, wsID := uuid.New(), uuid.New()
	user := &domainWorkspace.User{ID: userID, Email: "login@takt.dev", PasswordHash: "hashed_pass"}
	personalWs := &domainWorkspace.UserWorkspace{
		ID: wsID, Name: "Личное", Type: domainWorkspace.TypePersonal, Role: domainWorkspace.RoleOwner, CreatedAt: time.Now(),
	}

	testCases := []struct {
		name      string
		input     workspaceapp.LoginInput
		setupMock func(d mockDeps)
		checkOut  func(t *testing.T, out *workspaceapp.LoginOutput)
		wantErr   error
	}{
		{
			name:  "успешный вход в систему с личным воркспейсом",
			input: workspaceapp.LoginInput{Email: "login@takt.dev", Password: "correct_password"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "login@takt.dev").Return(user, nil).Once()
				d.hasher.EXPECT().Compare("hashed_pass", "correct_password").Return(nil).Once()
				d.wsRepo.EXPECT().GetPersonalWorkspaceByUserID(mock.Anything, userID).Return(personalWs, nil).Once()
				d.tokenGen.EXPECT().GenerateToken(userID, wsID).Return("jwt_login_token", nil).Once()
			},
			checkOut: func(t *testing.T, out *workspaceapp.LoginOutput) {
				assert.Equal(t, "jwt_login_token", out.Token)
				assert.Equal(t, "login@takt.dev", out.User.Email)
				assert.Equal(t, "Личное", out.Workspace.Name)
			},
		},
		{
			name:  "успешный вход с fallback на первый доступный воркспейс",
			input: workspaceapp.LoginInput{Email: "login@takt.dev", Password: "correct_password"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "login@takt.dev").Return(user, nil).Once()
				d.hasher.EXPECT().Compare("hashed_pass", "correct_password").Return(nil).Once()
				d.wsRepo.EXPECT().GetPersonalWorkspaceByUserID(mock.Anything, userID).Return(nil, domainWorkspace.ErrWorkspaceNotFound).Once()
				fallbackWs := &domainWorkspace.UserWorkspace{ID: wsID, Name: "Команда", Type: domainWorkspace.TypeShared, Role: domainWorkspace.RoleMember}
				d.wsRepo.EXPECT().GetWorkspacesByUserID(mock.Anything, userID).Return([]*domainWorkspace.UserWorkspace{fallbackWs}, nil).Once()
				d.tokenGen.EXPECT().GenerateToken(userID, wsID).Return("jwt_team_token", nil).Once()
			},
			checkOut: func(t *testing.T, out *workspaceapp.LoginOutput) {
				assert.Equal(t, "jwt_team_token", out.Token)
				assert.Equal(t, "Команда", out.Workspace.Name)
			},
		},
		{
			name:  "ошибка: пользователь не найден",
			input: workspaceapp.LoginInput{Email: "unknown@takt.dev", Password: "any_password"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "unknown@takt.dev").Return(nil, domainWorkspace.ErrUserNotFound).Once()
			},
			wantErr: domainWorkspace.ErrInvalidCredentials,
		},
		{
			name:  "ошибка: неверный пароль",
			input: workspaceapp.LoginInput{Email: "login@takt.dev", Password: "wrong_password"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "login@takt.dev").Return(user, nil).Once()
				d.hasher.EXPECT().Compare("hashed_pass", "wrong_password").Return(errors.New("mismatch")).Once()
			},
			wantErr: domainWorkspace.ErrInvalidCredentials,
		},
		{
			name:  "ошибка: у пользователя нет доступных воркспейсов",
			input: workspaceapp.LoginInput{Email: "login@takt.dev", Password: "correct_password"},
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByEmail(mock.Anything, "login@takt.dev").Return(user, nil).Once()
				d.hasher.EXPECT().Compare("hashed_pass", "correct_password").Return(nil).Once()
				d.wsRepo.EXPECT().GetPersonalWorkspaceByUserID(mock.Anything, userID).Return(nil, domainWorkspace.ErrWorkspaceNotFound).Once()
				d.wsRepo.EXPECT().GetWorkspacesByUserID(mock.Anything, userID).Return(nil, domainWorkspace.ErrWorkspaceNotFound).Once()
			},
			wantErr: domainWorkspace.ErrWorkspaceNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newMockDeps(t)
			tc.setupMock(deps)

			svc := deps.newService()
			out, err := svc.Login(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, out)
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				if tc.checkOut != nil {
					tc.checkOut(t, out)
				}
			}
		})
	}
}

func TestService_GetMe(t *testing.T) {
	userID, wsID := uuid.New(), uuid.New()
	user := &domainWorkspace.User{ID: userID, Email: "me@takt.dev"}
	member := &domainWorkspace.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: domainWorkspace.RoleOwner}
	ws := &domainWorkspace.Workspace{ID: wsID, Name: "Личное", Type: domainWorkspace.TypePersonal}

	testCases := []struct {
		name      string
		userID    uuid.UUID
		wsID      uuid.UUID
		setupMock func(d mockDeps)
		checkOut  func(t *testing.T, out *workspaceapp.GetProfileOutput)
		wantErr   error
	}{
		{
			name:   "успешное получение профиля",
			userID: userID,
			wsID:   wsID,
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil).Once()
				d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(member, nil).Once()
				d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(ws, nil).Once()
			},
			checkOut: func(t *testing.T, out *workspaceapp.GetProfileOutput) {
				assert.Equal(t, "me@takt.dev", out.User.Email)
				assert.Equal(t, "Личное", out.Workspace.Name)
				assert.Equal(t, domainWorkspace.RoleOwner, out.Workspace.Role)
			},
		},
		{
			name:   "ошибка: пользователь не найден",
			userID: userID,
			wsID:   wsID,
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, domainWorkspace.ErrUserNotFound).Once()
			},
			wantErr: domainWorkspace.ErrUserNotFound,
		},
		{
			name:   "ошибка: доступ запрещен (не участник)",
			userID: userID,
			wsID:   wsID,
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil).Once()
				d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(nil, domainWorkspace.ErrAccessDenied).Once()
			},
			wantErr: domainWorkspace.ErrAccessDenied,
		},
		{
			name:   "ошибка: воркспейс не найден",
			userID: userID,
			wsID:   wsID,
			setupMock: func(d mockDeps) {
				d.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil).Once()
				d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(member, nil).Once()
				d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(nil, domainWorkspace.ErrWorkspaceNotFound).Once()
			},
			wantErr: domainWorkspace.ErrWorkspaceNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newMockDeps(t)
			tc.setupMock(deps)

			svc := deps.newService()
			out, err := svc.GetMe(context.Background(), tc.userID, tc.wsID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, out)
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				if tc.checkOut != nil {
					tc.checkOut(t, out)
				}
			}
		})
	}
}
