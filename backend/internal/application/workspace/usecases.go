package workspace

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	domainWorkspace "takt/internal/domain/workspace"
)

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type RegisterOutput struct {
	Token     string
	User      *domainWorkspace.User
	Workspace *domainWorkspace.UserWorkspace
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token     string
	User      *domainWorkspace.User
	Workspace *domainWorkspace.UserWorkspace
}

type GetProfileOutput struct {
	User      *domainWorkspace.User
	Workspace *domainWorkspace.UserWorkspace
}

type UseCases interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	GetMe(ctx context.Context, userID, workspaceID uuid.UUID) (*GetProfileOutput, error)
	ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]*domainWorkspace.UserWorkspace, error)
	CreateWorkspace(ctx context.Context, userID uuid.UUID, name string) (*domainWorkspace.UserWorkspace, error)
	UpdateWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, name string) (*domainWorkspace.UserWorkspace, error)
	SwitchWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*SwitchWorkspaceOutput, error)
	ListMembers(ctx context.Context, userID, workspaceID uuid.UUID) ([]*WorkspaceMemberInfo, error)
	RemoveMember(ctx context.Context, actorID, workspaceID, targetUserID uuid.UUID) error
	CreateInvite(ctx context.Context, actorID, workspaceID uuid.UUID) (*domainWorkspace.WorkspaceInvite, error)
	GetInvitePreview(ctx context.Context, token string) (*InvitePreviewInfo, error)
	AcceptInvite(ctx context.Context, userID uuid.UUID, token string) (*SwitchWorkspaceOutput, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, displayName *string) (*domainWorkspace.User, error)
	UploadUserAvatar(ctx context.Context, userID uuid.UUID, filename, contentType string, reader io.Reader, size int64) (*domainWorkspace.User, error)
}

type Service struct {
	userRepo         UserRepository
	wsRepo           WorkspaceRepository
	hasher           PasswordHasher
	tokenGen         TokenGenerator
	tasksCreator     InitialTasksCreator
	transactor       Transactor
	automationSeeder AutomationSeeder
	avatarStorage    AvatarStorage
}

// SetAvatarStorage sets the optional avatar storage service.
func (s *Service) SetAvatarStorage(storage AvatarStorage) {
	s.avatarStorage = storage
}

func NewService(
	userRepo UserRepository,
	wsRepo WorkspaceRepository,
	hasher PasswordHasher,
	tokenGen TokenGenerator,
	tasksCreator InitialTasksCreator,
	transactor Transactor,
	automationSeeder ...AutomationSeeder,
) *Service {
	var seeder AutomationSeeder
	if len(automationSeeder) > 0 {
		seeder = automationSeeder[0]
	}
	return &Service{
		userRepo:         userRepo,
		wsRepo:           wsRepo,
		hasher:           hasher,
		tokenGen:         tokenGen,
		tasksCreator:     tasksCreator,
		transactor:       transactor,
		automationSeeder: seeder,
	}
}

func (s *Service) runInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if s.transactor != nil {
		return s.transactor.WithinTx(ctx, fn)
	}
	return fn(ctx)
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if len(strings.TrimSpace(input.Password)) < 6 {
		return nil, domainWorkspace.ErrInvalidPassword
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, domainWorkspace.ErrUserAlreadyExists
	}

	hashedPassword, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user, err := domainWorkspace.NewUser(input.Email, hashedPassword, input.DisplayName)
	if err != nil {
		return nil, err
	}

	var (
		ws     *domainWorkspace.Workspace
		member *domainWorkspace.WorkspaceMember
	)

	err = s.runInTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return err
		}
		var wsErr error
		ws, wsErr = domainWorkspace.NewPersonalWorkspace(user.ID, "Личное")
		if wsErr != nil {
			return wsErr
		}
		if err := s.wsRepo.Create(txCtx, ws); err != nil {
			return err
		}
		var memErr error
		member, memErr = domainWorkspace.NewWorkspaceMember(ws.ID, user.ID, domainWorkspace.RoleOwner)
		if memErr != nil {
			return memErr
		}
		if err := s.wsRepo.CreateMember(txCtx, member); err != nil {
			return err
		}
		if s.tasksCreator != nil {
			if err := s.tasksCreator.CreateInitialInboxTasks(txCtx, ws.ID); err != nil {
				return err
			}
		}
		if s.automationSeeder != nil {
			if err := s.automationSeeder.SeedDefaultAutomations(txCtx, ws.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	token, err := s.tokenGen.GenerateToken(user.ID, ws.ID)
	if err != nil {
		return nil, err
	}

	userWs := &domainWorkspace.UserWorkspace{
		ID:        ws.ID,
		Name:      ws.Name,
		Type:      ws.Type,
		Role:      member.Role,
		CreatedAt: ws.CreatedAt,
	}

	return &RegisterOutput{
		Token:     token,
		User:      user,
		Workspace: userWs,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil || user == nil {
		return nil, domainWorkspace.ErrInvalidCredentials
	}

	if err := s.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, domainWorkspace.ErrInvalidCredentials
	}

	personalWs, err := s.wsRepo.GetPersonalWorkspaceByUserID(ctx, user.ID)
	if err != nil || personalWs == nil {
		// Fallback to first accessible workspace
		allWs, errWs := s.wsRepo.GetWorkspacesByUserID(ctx, user.ID)
		if errWs != nil || len(allWs) == 0 {
			return nil, domainWorkspace.ErrWorkspaceNotFound
		}
		personalWs = allWs[0]
	}

	token, err := s.tokenGen.GenerateToken(user.ID, personalWs.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:     token,
		User:      user,
		Workspace: personalWs,
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID, workspaceID uuid.UUID) (*GetProfileOutput, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domainWorkspace.ErrUserNotFound
	}

	member, err := s.wsRepo.GetMember(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, domainWorkspace.ErrAccessDenied
	}

	ws, err := s.wsRepo.GetByID(ctx, workspaceID)
	if err != nil || ws == nil {
		return nil, domainWorkspace.ErrWorkspaceNotFound
	}

	userWs := &domainWorkspace.UserWorkspace{
		ID:        ws.ID,
		Name:      ws.Name,
		Type:      ws.Type,
		Role:      member.Role,
		CreatedAt: ws.CreatedAt,
	}

	return &GetProfileOutput{
		User:      user,
		Workspace: userWs,
	}, nil
}

func (s *Service) UpdateUserProfile(ctx context.Context, userID uuid.UUID, displayName *string) (*domainWorkspace.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domainWorkspace.ErrUserNotFound
	}

	if err := user.UpdateProfile(displayName); err != nil {
		return nil, err
	}

	return s.userRepo.UpdateProfile(ctx, userID, &user.DisplayName, nil)
}

func (s *Service) UploadUserAvatar(
	ctx context.Context,
	userID uuid.UUID,
	filename, contentType string,
	reader io.Reader,
	size int64,
) (*domainWorkspace.User, error) {
	if size > 5*1024*1024 {
		return nil, domainWorkspace.ErrAvatarTooLarge
	}

	cType := strings.ToLower(strings.TrimSpace(contentType))
	var ext string
	switch {
	case strings.Contains(cType, "jpeg") || strings.Contains(cType, "jpg"):
		ext = ".jpg"
	case strings.Contains(cType, "png"):
		ext = ".png"
	case strings.Contains(cType, "webp"):
		ext = ".webp"
	default:
		// Fallback check on filename extension
		fExt := strings.ToLower(filepath.Ext(filename))
		if fExt == ".jpg" || fExt == ".jpeg" {
			ext = ".jpg"
		} else if fExt == ".png" {
			ext = ".png"
		} else if fExt == ".webp" {
			ext = ".webp"
		} else {
			return nil, domainWorkspace.ErrInvalidAvatarFormat
		}
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domainWorkspace.ErrUserNotFound
	}

	key := fmt.Sprintf("avatars/%s/%s%s", userID.String(), uuid.New().String(), ext)
	if s.avatarStorage != nil {
		if err := s.avatarStorage.PutObject(ctx, key, reader, size, contentType); err != nil {
			return nil, fmt.Errorf("failed to store avatar: %w", err)
		}
	}

	avatarURL := fmt.Sprintf("/api/v1/users/me/avatar/%s", key)
	user.SetAvatar(avatarURL)

	return s.userRepo.UpdateProfile(ctx, userID, nil, user.AvatarURL)
}
