package workspace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	domainWorkspace "takt/internal/domain/workspace"
)

// ListWorkspaces returns all workspaces the user has access to.
func (s *Service) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]*domainWorkspace.UserWorkspace, error) {
	return s.wsRepo.GetWorkspacesByUserID(ctx, userID)
}

// CreateWorkspace creates a new unified workspace, assigns user as owner, and seeds initial tasks and default automations.
func (s *Service) CreateWorkspace(ctx context.Context, userID uuid.UUID, name string) (*domainWorkspace.UserWorkspace, error) {
	ws, err := domainWorkspace.NewWorkspace(userID, name)
	if err != nil {
		return nil, err
	}

	var member *domainWorkspace.WorkspaceMember
	err = s.runInTx(ctx, func(txCtx context.Context) error {
		if err := s.wsRepo.Create(txCtx, ws); err != nil {
			return err
		}

		var memErr error
		member, memErr = domainWorkspace.NewWorkspaceMember(ws.ID, userID, domainWorkspace.RoleOwner)
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

	return &domainWorkspace.UserWorkspace{
		ID:        ws.ID,
		OwnerID:   ws.OwnerID,
		Name:      ws.Name,
		Type:      ws.Type,
		Role:      member.Role,
		CreatedAt: ws.CreatedAt,
	}, nil
}

// UpdateWorkspace renames a workspace (only owner can rename).
func (s *Service) UpdateWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, name string) (*domainWorkspace.UserWorkspace, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 255 {
		return nil, domainWorkspace.ErrInvalidWorkspaceName
	}

	member, err := s.wsRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, domainWorkspace.ErrAccessDenied
	}
	if member.Role != domainWorkspace.RoleOwner {
		return nil, domainWorkspace.ErrAccessDenied
	}

	ws, err := s.wsRepo.UpdateName(ctx, workspaceID, userID, trimmedName)
	if err != nil {
		return nil, err
	}

	return &domainWorkspace.UserWorkspace{
		ID:        ws.ID,
		OwnerID:   ws.OwnerID,
		Name:      ws.Name,
		Type:      ws.Type,
		Role:      member.Role,
		CreatedAt: ws.CreatedAt,
	}, nil
}

// SwitchWorkspace validates membership in target workspace, generates a new JWT token, and returns workspace info.
func (s *Service) SwitchWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*SwitchWorkspaceOutput, error) {
	member, err := s.wsRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, domainWorkspace.ErrAccessDenied
	}

	ws, err := s.wsRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenGen.GenerateToken(userID, workspaceID)
	if err != nil {
		return nil, err
	}

	return &SwitchWorkspaceOutput{
		Token: token,
		Workspace: &domainWorkspace.UserWorkspace{
			ID:        ws.ID,
			OwnerID:   ws.OwnerID,
			Name:      ws.Name,
			Type:      ws.Type,
			Role:      member.Role,
			CreatedAt: ws.CreatedAt,
		},
	}, nil
}

// ListMembers lists all members of the workspace if the user is a member.
func (s *Service) ListMembers(ctx context.Context, userID, workspaceID uuid.UUID) ([]*WorkspaceMemberInfo, error) {
	if _, err := s.wsRepo.GetMember(ctx, workspaceID, userID); err != nil {
		return nil, domainWorkspace.ErrAccessDenied
	}

	return s.wsRepo.ListMembers(ctx, workspaceID)
}

// RemoveMember removes a member from a workspace, or allows a member to leave.
// Owner cannot be removed. Non-owners can only remove themselves.
func (s *Service) RemoveMember(ctx context.Context, actorID, workspaceID, targetUserID uuid.UUID) error {
	actorMember, err := s.wsRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return domainWorkspace.ErrAccessDenied
	}

	ws, err := s.wsRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if targetUserID == ws.OwnerID {
		return domainWorkspace.ErrCannotRemoveOwner
	}

	if actorID != targetUserID && actorMember.Role != domainWorkspace.RoleOwner {
		return domainWorkspace.ErrAccessDenied
	}

	if _, err := s.wsRepo.GetMember(ctx, workspaceID, targetUserID); err != nil {
		return domainWorkspace.ErrWorkspaceNotFound
	}

	return s.wsRepo.DeleteMember(ctx, workspaceID, targetUserID)
}

// CreateInvite creates or retrieves an active invite link for a workspace (only owner can create).
func (s *Service) CreateInvite(ctx context.Context, actorID, workspaceID uuid.UUID) (*domainWorkspace.WorkspaceInvite, error) {
	actorMember, err := s.wsRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, domainWorkspace.ErrAccessDenied
	}
	if actorMember.Role != domainWorkspace.RoleOwner {
		return nil, domainWorkspace.ErrAccessDenied
	}

	active, err := s.wsRepo.GetActiveInvite(ctx, workspaceID)
	if err == nil && active != nil && !active.IsExpired() {
		return active, nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	invite, err := domainWorkspace.NewWorkspaceInvite(workspaceID, actorID, token, &expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.wsRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

// GetInvitePreview retrieves public preview information for an invite token.
func (s *Service) GetInvitePreview(ctx context.Context, token string) (*InvitePreviewInfo, error) {
	invite, preview, err := s.wsRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, domainWorkspace.ErrInviteNotFound
	}
	if invite.IsExpired() {
		return nil, domainWorkspace.ErrInviteExpired
	}

	return preview, nil
}

// AcceptInvite accepts an invite by token, adds the authenticated user as a member, and returns the switched workspace.
func (s *Service) AcceptInvite(ctx context.Context, userID uuid.UUID, token string) (*SwitchWorkspaceOutput, error) {
	invite, _, err := s.wsRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, domainWorkspace.ErrInviteNotFound
	}
	if invite.IsExpired() {
		return nil, domainWorkspace.ErrInviteExpired
	}

	existing, _ := s.wsRepo.GetMember(ctx, invite.WorkspaceID, userID)
	if existing != nil {
		return nil, domainWorkspace.ErrAlreadyMember
	}

	newMember, err := domainWorkspace.NewWorkspaceMember(invite.WorkspaceID, userID, domainWorkspace.RoleMember)
	if err != nil {
		return nil, err
	}

	if err := s.wsRepo.CreateMember(ctx, newMember); err != nil {
		return nil, err
	}

	ws, err := s.wsRepo.GetByID(ctx, invite.WorkspaceID)
	if err != nil {
		return nil, err
	}

	jwtToken, err := s.tokenGen.GenerateToken(userID, invite.WorkspaceID)
	if err != nil {
		return nil, err
	}

	return &SwitchWorkspaceOutput{
		Token: jwtToken,
		Workspace: &domainWorkspace.UserWorkspace{
			ID:        ws.ID,
			OwnerID:   ws.OwnerID,
			Name:      ws.Name,
			Type:      ws.Type,
			Role:      domainWorkspace.RoleMember,
			CreatedAt: ws.CreatedAt,
		},
	}, nil
}
