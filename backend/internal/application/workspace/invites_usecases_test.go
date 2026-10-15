package workspace_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	workspaceapp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
)

func TestService_CreateInvite(t *testing.T) {
	ownerID := uuid.New()
	memberID := uuid.New()
	wsID := uuid.New()

	t.Run("returns active unexpired invite", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, ownerID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: ownerID, Role: domainWorkspace.RoleOwner,
		}, nil).Once()

		future := time.Now().UTC().Add(24 * time.Hour)
		active := &domainWorkspace.WorkspaceInvite{
			ID:          uuid.New(),
			WorkspaceID: wsID,
			InvitedByID: ownerID,
			Token:       "active-token",
			ExpiresAt:   &future,
		}
		d.wsRepo.EXPECT().GetActiveInvite(mock.Anything, wsID).Return(active, nil).Once()

		invite, err := svc.CreateInvite(context.Background(), ownerID, wsID)
		require.NoError(t, err)
		assert.Equal(t, "active-token", invite.Token)
	})

	t.Run("creates new invite when no active invite", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, ownerID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: ownerID, Role: domainWorkspace.RoleOwner,
		}, nil).Once()

		d.wsRepo.EXPECT().GetActiveInvite(mock.Anything, wsID).Return(nil, nil).Once()
		d.wsRepo.EXPECT().CreateInvite(mock.Anything, mock.AnythingOfType("*workspace.WorkspaceInvite")).Return(nil).Once()

		invite, err := svc.CreateInvite(context.Background(), ownerID, wsID)
		require.NoError(t, err)
		assert.NotEmpty(t, invite.Token)
		assert.Equal(t, 64, len(invite.Token))
	})

	t.Run("non-owner forbidden", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, memberID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: memberID, Role: domainWorkspace.RoleMember,
		}, nil).Once()

		invite, err := svc.CreateInvite(context.Background(), memberID, wsID)
		assert.ErrorIs(t, err, domainWorkspace.ErrAccessDenied)
		assert.Nil(t, invite)
	})
}

func TestService_GetInvitePreview(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		future := time.Now().UTC().Add(24 * time.Hour)
		invite := &domainWorkspace.WorkspaceInvite{
			Token:     "token123",
			ExpiresAt: &future,
		}
		preview := &workspaceapp.InvitePreviewInfo{
			Token:         "token123",
			WorkspaceName: "Team",
			InvitedByName: "Owner",
			ExpiresAt:     &future,
		}

		d.wsRepo.EXPECT().GetInviteByToken(mock.Anything, "token123").Return(invite, preview, nil).Once()

		res, err := svc.GetInvitePreview(context.Background(), "token123")
		require.NoError(t, err)
		assert.Equal(t, "Team", res.WorkspaceName)
	})

	t.Run("expired invite", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		past := time.Now().UTC().Add(-24 * time.Hour)
		invite := &domainWorkspace.WorkspaceInvite{
			Token:     "token123",
			ExpiresAt: &past,
		}

		d.wsRepo.EXPECT().GetInviteByToken(mock.Anything, "token123").Return(invite, nil, nil).Once()

		res, err := svc.GetInvitePreview(context.Background(), "token123")
		assert.ErrorIs(t, err, domainWorkspace.ErrInviteExpired)
		assert.Nil(t, res)
	})
}

func TestService_AcceptInvite(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		future := time.Now().UTC().Add(24 * time.Hour)
		invite := &domainWorkspace.WorkspaceInvite{
			WorkspaceID: wsID,
			Token:       "token123",
			ExpiresAt:   &future,
		}

		d.wsRepo.EXPECT().GetInviteByToken(mock.Anything, "token123").Return(invite, nil, nil).Once()
		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(nil, domainWorkspace.ErrAccessDenied).Once()
		d.wsRepo.EXPECT().CreateMember(mock.Anything, mock.AnythingOfType("*workspace.WorkspaceMember")).Return(nil).Once()
		d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(&domainWorkspace.Workspace{
			ID: wsID, Name: "Workspace",
		}, nil).Once()
		d.tokenGen.EXPECT().GenerateToken(userID, wsID).Return("new-token", nil).Once()

		res, err := svc.AcceptInvite(context.Background(), userID, "token123")
		require.NoError(t, err)
		assert.Equal(t, "new-token", res.Token)
		assert.Equal(t, wsID, res.Workspace.ID)
	})

	t.Run("already a member", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		future := time.Now().UTC().Add(24 * time.Hour)
		invite := &domainWorkspace.WorkspaceInvite{
			WorkspaceID: wsID,
			Token:       "token123",
			ExpiresAt:   &future,
		}

		d.wsRepo.EXPECT().GetInviteByToken(mock.Anything, "token123").Return(invite, nil, nil).Once()
		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: userID, Role: domainWorkspace.RoleMember,
		}, nil).Once()

		res, err := svc.AcceptInvite(context.Background(), userID, "token123")
		assert.ErrorIs(t, err, domainWorkspace.ErrAlreadyMember)
		assert.Nil(t, res)
	})
}
