package workspace_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	workspaceapp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
)

func TestService_ListWorkspaces(t *testing.T) {
	d := newMockDeps(t)
	svc := d.newService()
	userID := uuid.New()
	wsID := uuid.New()

	expected := []*domainWorkspace.UserWorkspace{
		{ID: wsID, OwnerID: userID, Name: "Workspace 1", Role: domainWorkspace.RoleOwner},
	}

	d.wsRepo.EXPECT().GetWorkspacesByUserID(mock.Anything, userID).Return(expected, nil).Once()

	res, err := svc.ListWorkspaces(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestService_CreateWorkspace(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()

		d.wsRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*workspace.Workspace")).Return(nil).Once()
		d.wsRepo.EXPECT().CreateMember(mock.Anything, mock.AnythingOfType("*workspace.WorkspaceMember")).Return(nil).Once()
		d.tasksCreator.EXPECT().CreateInitialInboxTasks(mock.Anything, mock.Anything).Return(nil).Once()

		res, err := svc.CreateWorkspace(context.Background(), userID, "New Team")
		require.NoError(t, err)
		assert.Equal(t, "New Team", res.Name)
		assert.Equal(t, domainWorkspace.RoleOwner, res.Role)
	})

	t.Run("empty name validation error", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		res, err := svc.CreateWorkspace(context.Background(), userID, "")
		assert.ErrorIs(t, err, domainWorkspace.ErrInvalidWorkspaceName)
		assert.Nil(t, res)
	})
}

func TestService_UpdateWorkspace(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success by owner", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID,
			UserID:      userID,
			Role:        domainWorkspace.RoleOwner,
		}, nil).Once()

		d.wsRepo.EXPECT().UpdateName(mock.Anything, wsID, userID, "Renamed WS").Return(&domainWorkspace.Workspace{
			ID:      wsID,
			OwnerID: userID,
			Name:    "Renamed WS",
		}, nil).Once()

		res, err := svc.UpdateWorkspace(context.Background(), userID, wsID, "Renamed WS")
		require.NoError(t, err)
		assert.Equal(t, "Renamed WS", res.Name)
	})

	t.Run("forbidden for non-owner", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID,
			UserID:      userID,
			Role:        domainWorkspace.RoleMember,
		}, nil).Once()

		res, err := svc.UpdateWorkspace(context.Background(), userID, wsID, "Renamed WS")
		assert.ErrorIs(t, err, domainWorkspace.ErrAccessDenied)
		assert.Nil(t, res)
	})
}

func TestService_SwitchWorkspace(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID,
			UserID:      userID,
			Role:        domainWorkspace.RoleMember,
		}, nil).Once()

		d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(&domainWorkspace.Workspace{
			ID:      wsID,
			OwnerID: uuid.New(),
			Name:    "Target WS",
		}, nil).Once()

		d.tokenGen.EXPECT().GenerateToken(userID, wsID).Return("new-jwt-token", nil).Once()

		res, err := svc.SwitchWorkspace(context.Background(), userID, wsID)
		require.NoError(t, err)
		assert.Equal(t, "new-jwt-token", res.Token)
		assert.Equal(t, wsID, res.Workspace.ID)
	})

	t.Run("not a member", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(nil, domainWorkspace.ErrAccessDenied).Once()

		res, err := svc.SwitchWorkspace(context.Background(), userID, wsID)
		assert.ErrorIs(t, err, domainWorkspace.ErrAccessDenied)
		assert.Nil(t, res)
	})
}

func TestService_ListMembers(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID,
			UserID:      userID,
			Role:        domainWorkspace.RoleMember,
		}, nil).Once()

		expected := []*workspaceapp.WorkspaceMemberInfo{
			{ID: uuid.New(), UserID: userID, DisplayName: "User 1", Role: domainWorkspace.RoleMember},
		}
		d.wsRepo.EXPECT().ListMembers(mock.Anything, wsID).Return(expected, nil).Once()

		res, err := svc.ListMembers(context.Background(), userID, wsID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("not a member", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, userID).Return(nil, domainWorkspace.ErrAccessDenied).Once()

		res, err := svc.ListMembers(context.Background(), userID, wsID)
		assert.ErrorIs(t, err, domainWorkspace.ErrAccessDenied)
		assert.Nil(t, res)
	})
}

func TestService_RemoveMember(t *testing.T) {
	ownerID := uuid.New()
	memberID := uuid.New()
	otherMemberID := uuid.New()
	wsID := uuid.New()

	t.Run("owner removes member", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, ownerID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: ownerID, Role: domainWorkspace.RoleOwner,
		}, nil).Once()
		d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(&domainWorkspace.Workspace{
			ID: wsID, OwnerID: ownerID,
		}, nil).Once()
		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, memberID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: memberID, Role: domainWorkspace.RoleMember,
		}, nil).Once()
		d.wsRepo.EXPECT().DeleteMember(mock.Anything, wsID, memberID).Return(nil).Once()

		err := svc.RemoveMember(context.Background(), ownerID, wsID, memberID)
		require.NoError(t, err)
	})

	t.Run("member leaves own workspace", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, memberID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: memberID, Role: domainWorkspace.RoleMember,
		}, nil).Once()
		d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(&domainWorkspace.Workspace{
			ID: wsID, OwnerID: ownerID,
		}, nil).Once()
		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, memberID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: memberID, Role: domainWorkspace.RoleMember,
		}, nil).Once()
		d.wsRepo.EXPECT().DeleteMember(mock.Anything, wsID, memberID).Return(nil).Once()

		err := svc.RemoveMember(context.Background(), memberID, wsID, memberID)
		require.NoError(t, err)
	})

	t.Run("cannot remove owner", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, ownerID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: ownerID, Role: domainWorkspace.RoleOwner,
		}, nil).Once()
		d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(&domainWorkspace.Workspace{
			ID: wsID, OwnerID: ownerID,
		}, nil).Once()

		err := svc.RemoveMember(context.Background(), ownerID, wsID, ownerID)
		assert.ErrorIs(t, err, domainWorkspace.ErrCannotRemoveOwner)
	})

	t.Run("non-owner cannot remove other member", func(t *testing.T) {
		d := newMockDeps(t)
		svc := d.newService()

		d.wsRepo.EXPECT().GetMember(mock.Anything, wsID, memberID).Return(&domainWorkspace.WorkspaceMember{
			WorkspaceID: wsID, UserID: memberID, Role: domainWorkspace.RoleMember,
		}, nil).Once()
		d.wsRepo.EXPECT().GetByID(mock.Anything, wsID).Return(&domainWorkspace.Workspace{
			ID: wsID, OwnerID: ownerID,
		}, nil).Once()

		err := svc.RemoveMember(context.Background(), memberID, wsID, otherMemberID)
		assert.ErrorIs(t, err, domainWorkspace.ErrAccessDenied)
	})
}
