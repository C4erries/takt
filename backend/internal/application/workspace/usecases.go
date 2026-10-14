package workspace

import (
	"context"
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
}

type Service struct {
	userRepo     UserRepository
	wsRepo       WorkspaceRepository
	hasher       PasswordHasher
	tokenGen     TokenGenerator
	tasksCreator     InitialTasksCreator
	transactor       Transactor
	automationSeeder AutomationSeeder
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
