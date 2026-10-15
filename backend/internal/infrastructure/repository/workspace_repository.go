package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	workspaceApp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/db/sqlc"
)

type WorkspaceRepository struct {
	queries *sqlc.Queries
}

func NewWorkspaceRepository(queries *sqlc.Queries) *WorkspaceRepository {
	return &WorkspaceRepository{queries: queries}
}

func (r *WorkspaceRepository) Create(ctx context.Context, ws *domainWorkspace.Workspace) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		OwnerID: uuidToPg(ws.OwnerID),
		Name:    ws.Name,
		Type:    string(ws.Type),
	})
	if err != nil {
		return err
	}

	ws.ID = pgToUUID(row.ID)
	ws.CreatedAt = pgToTime(row.CreatedAt)
	return nil
}

func (r *WorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainWorkspace.Workspace, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetWorkspaceByID(ctx, uuidToPg(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrWorkspaceNotFound
		}
		return nil, err
	}

	return &domainWorkspace.Workspace{
		ID:        pgToUUID(row.ID),
		OwnerID:   pgToUUID(row.OwnerID),
		Name:      row.Name,
		Type:      domainWorkspace.WorkspaceType(row.Type),
		CreatedAt: pgToTime(row.CreatedAt),
	}, nil
}

func (r *WorkspaceRepository) UpdateName(ctx context.Context, id, ownerID uuid.UUID, name string) (*domainWorkspace.Workspace, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateWorkspaceName(ctx, sqlc.UpdateWorkspaceNameParams{
		ID:      uuidToPg(id),
		Name:    name,
		OwnerID: uuidToPg(ownerID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrWorkspaceNotFound
		}
		return nil, err
	}

	return &domainWorkspace.Workspace{
		ID:        pgToUUID(row.ID),
		OwnerID:   pgToUUID(row.OwnerID),
		Name:      row.Name,
		Type:      domainWorkspace.WorkspaceType(row.Type),
		CreatedAt: pgToTime(row.CreatedAt),
	}, nil
}

func (r *WorkspaceRepository) CreateMember(ctx context.Context, member *domainWorkspace.WorkspaceMember) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberParams{
		WorkspaceID: uuidToPg(member.WorkspaceID),
		UserID:      uuidToPg(member.UserID),
		Role:        string(member.Role),
	})
	if err != nil {
		return err
	}

	member.ID = pgToUUID(row.ID)
	member.CreatedAt = pgToTime(row.CreatedAt)
	return nil
}

func (r *WorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*domainWorkspace.WorkspaceMember, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: uuidToPg(workspaceID),
		UserID:      uuidToPg(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrAccessDenied
		}
		return nil, err
	}

	return &domainWorkspace.WorkspaceMember{
		ID:          pgToUUID(row.ID),
		WorkspaceID: pgToUUID(row.WorkspaceID),
		UserID:      pgToUUID(row.UserID),
		Role:        domainWorkspace.MemberRole(row.Role),
		CreatedAt:   pgToTime(row.CreatedAt),
	}, nil
}

func (r *WorkspaceRepository) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*workspaceApp.WorkspaceMemberInfo, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListWorkspaceMembers(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}

	members := make([]*workspaceApp.WorkspaceMemberInfo, 0, len(rows))
	for _, row := range rows {
		members = append(members, &workspaceApp.WorkspaceMemberInfo{
			ID:          pgToUUID(row.ID),
			UserID:      pgToUUID(row.UserID),
			Email:       row.Email,
			DisplayName: row.DisplayName,
			Role:        domainWorkspace.MemberRole(row.Role),
			CreatedAt:   pgToTime(row.CreatedAt),
		})
	}
	return members, nil
}

func (r *WorkspaceRepository) DeleteMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteWorkspaceMember(ctx, sqlc.DeleteWorkspaceMemberParams{
		WorkspaceID: uuidToPg(workspaceID),
		UserID:      uuidToPg(userID),
	})
}

func (r *WorkspaceRepository) GetWorkspacesByUserID(ctx context.Context, userID uuid.UUID) ([]*domainWorkspace.UserWorkspace, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.GetWorkspacesByUserID(ctx, uuidToPg(userID))
	if err != nil {
		return nil, err
	}

	res := make([]*domainWorkspace.UserWorkspace, 0, len(rows))
	for _, row := range rows {
		res = append(res, &domainWorkspace.UserWorkspace{
			ID:        pgToUUID(row.ID),
			OwnerID:   pgToUUID(row.OwnerID),
			Name:      row.Name,
			Type:      domainWorkspace.WorkspaceType(row.Type),
			Role:      domainWorkspace.MemberRole(row.Role),
			CreatedAt: pgToTime(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *WorkspaceRepository) GetPersonalWorkspaceByUserID(ctx context.Context, userID uuid.UUID) (*domainWorkspace.UserWorkspace, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetPersonalWorkspaceByUserID(ctx, uuidToPg(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrWorkspaceNotFound
		}
		return nil, err
	}

	return &domainWorkspace.UserWorkspace{
		ID:        pgToUUID(row.ID),
		OwnerID:   pgToUUID(row.OwnerID),
		Name:      row.Name,
		Type:      domainWorkspace.WorkspaceType(row.Type),
		Role:      domainWorkspace.MemberRole(row.Role),
		CreatedAt: pgToTime(row.CreatedAt),
	}, nil
}

func (r *WorkspaceRepository) CreateInvite(ctx context.Context, invite *domainWorkspace.WorkspaceInvite) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateWorkspaceInvite(ctx, sqlc.CreateWorkspaceInviteParams{
		WorkspaceID: uuidToPg(invite.WorkspaceID),
		InvitedByID: uuidToPg(invite.InvitedByID),
		Token:       invite.Token,
		ExpiresAt:   timePtrToPg(invite.ExpiresAt),
	})
	if err != nil {
		return err
	}

	invite.ID = pgToUUID(row.ID)
	invite.CreatedAt = pgToTime(row.CreatedAt)
	return nil
}

func (r *WorkspaceRepository) GetInviteByToken(ctx context.Context, token string) (*domainWorkspace.WorkspaceInvite, *workspaceApp.InvitePreviewInfo, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetWorkspaceInviteByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domainWorkspace.ErrInviteNotFound
		}
		return nil, nil, err
	}

	invite := &domainWorkspace.WorkspaceInvite{
		ID:          pgToUUID(row.ID),
		WorkspaceID: pgToUUID(row.WorkspaceID),
		InvitedByID: pgToUUID(row.InvitedByID),
		Token:       row.Token,
		ExpiresAt:   pgToTimePtr(row.ExpiresAt),
		CreatedAt:   pgToTime(row.CreatedAt),
	}

	preview := &workspaceApp.InvitePreviewInfo{
		Token:         row.Token,
		WorkspaceName: row.WorkspaceName,
		InvitedByName: row.InvitedByName,
		ExpiresAt:     invite.ExpiresAt,
	}

	return invite, preview, nil
}

func (r *WorkspaceRepository) GetActiveInvite(ctx context.Context, workspaceID uuid.UUID) (*domainWorkspace.WorkspaceInvite, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetActiveWorkspaceInvite(ctx, uuidToPg(workspaceID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domainWorkspace.WorkspaceInvite{
		ID:          pgToUUID(row.ID),
		WorkspaceID: pgToUUID(row.WorkspaceID),
		InvitedByID: pgToUUID(row.InvitedByID),
		Token:       row.Token,
		ExpiresAt:   pgToTimePtr(row.ExpiresAt),
		CreatedAt:   pgToTime(row.CreatedAt),
	}, nil
}

func (r *WorkspaceRepository) DeleteInvite(ctx context.Context, id uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteWorkspaceInvite(ctx, uuidToPg(id))
}

func (r *WorkspaceRepository) DeleteInviteByToken(ctx context.Context, token string) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteWorkspaceInviteByToken(ctx, token)
}
