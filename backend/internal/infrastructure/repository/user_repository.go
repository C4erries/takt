package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/db/sqlc"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, u *domainWorkspace.User) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
	})
	if err != nil {
		return err
	}

	u.ID = pgToUUID(row.ID)
	u.CreatedAt = pgToTime(row.CreatedAt)
	u.AvatarURL = pgToTextPtr(row.AvatarUrl)
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domainWorkspace.User, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrUserNotFound
		}
		return nil, err
	}

	return &domainWorkspace.User{
		ID:           pgToUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		AvatarURL:    pgToTextPtr(row.AvatarUrl),
		CreatedAt:    pgToTime(row.CreatedAt),
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainWorkspace.User, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetUserByID(ctx, uuidToPg(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrUserNotFound
		}
		return nil, err
	}

	return &domainWorkspace.User{
		ID:           pgToUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		AvatarURL:    pgToTextPtr(row.AvatarUrl),
		CreatedAt:    pgToTime(row.CreatedAt),
	}, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, displayName *string, avatarURL *string) (*domainWorkspace.User, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:          uuidToPg(id),
		DisplayName: textPtrToPg(displayName),
		AvatarUrl:   textPtrToPg(avatarURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWorkspace.ErrUserNotFound
		}
		return nil, err
	}

	return &domainWorkspace.User{
		ID:           pgToUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		AvatarURL:    pgToTextPtr(row.AvatarUrl),
		CreatedAt:    pgToTime(row.CreatedAt),
	}, nil
}
