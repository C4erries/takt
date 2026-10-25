package workspace_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"takt/internal/application/workspace/mocks"
	domainWorkspace "takt/internal/domain/workspace"
)

func TestService_UpdateUserProfile(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()

	testUser := &domainWorkspace.User{
		ID:          userID,
		Email:       "user@takt.dev",
		DisplayName: "Old Name",
		CreatedAt:   now,
	}

	t.Run("successful display name update", func(t *testing.T) {
		deps := newMockDeps(t)
		svc := deps.newService()

		newName := "New Name"
		deps.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(testUser, nil).Once()
		deps.userRepo.EXPECT().UpdateProfile(mock.Anything, userID, &newName, (*string)(nil)).
			Return(&domainWorkspace.User{
				ID:          userID,
				Email:       "user@takt.dev",
				DisplayName: newName,
				CreatedAt:   now,
			}, nil).Once()

		updated, err := svc.UpdateUserProfile(context.Background(), userID, &newName)
		require.NoError(t, err)
		assert.Equal(t, "New Name", updated.DisplayName)
	})

	t.Run("user not found", func(t *testing.T) {
		deps := newMockDeps(t)
		svc := deps.newService()

		newName := "New Name"
		deps.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, domainWorkspace.ErrUserNotFound).Once()

		updated, err := svc.UpdateUserProfile(context.Background(), userID, &newName)
		require.ErrorIs(t, err, domainWorkspace.ErrUserNotFound)
		assert.Nil(t, updated)
	})

	t.Run("invalid display name fails validation", func(t *testing.T) {
		deps := newMockDeps(t)
		svc := deps.newService()

		emptyName := ""
		deps.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(testUser, nil).Once()

		updated, err := svc.UpdateUserProfile(context.Background(), userID, &emptyName)
		require.ErrorIs(t, err, domainWorkspace.ErrInvalidDisplayName)
		assert.Nil(t, updated)
	})
}

func TestService_UploadUserAvatar(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()

	testUser := &domainWorkspace.User{
		ID:          userID,
		Email:       "user@takt.dev",
		DisplayName: "Pavel",
		CreatedAt:   now,
	}

	t.Run("successful png avatar upload", func(t *testing.T) {
		deps := newMockDeps(t)
		mockStorage := mocks.NewMockAvatarStorage(t)
		svc := deps.newService()
		svc.SetAvatarStorage(mockStorage)

		imageData := []byte("fake-png-data")
		reader := bytes.NewReader(imageData)

		deps.userRepo.EXPECT().GetByID(mock.Anything, userID).Return(testUser, nil).Once()
		mockStorage.EXPECT().PutObject(mock.Anything, mock.MatchedBy(func(key string) bool {
			return len(key) > 0
		}), reader, int64(len(imageData)), "image/png").Return(nil).Once()

		deps.userRepo.EXPECT().UpdateProfile(mock.Anything, userID, (*string)(nil), mock.MatchedBy(func(url *string) bool {
			return url != nil && len(*url) > 0
		})).Return(&domainWorkspace.User{
			ID:          userID,
			Email:       "user@takt.dev",
			DisplayName: "Pavel",
			AvatarURL:   func() *string { s := "/api/v1/users/me/avatar/avatars/test.png"; return &s }(),
			CreatedAt:   now,
		}, nil).Once()

		user, err := svc.UploadUserAvatar(context.Background(), userID, "avatar.png", "image/png", reader, int64(len(imageData)))
		require.NoError(t, err)
		assert.NotNil(t, user.AvatarURL)
	})

	t.Run("file size exceeds limit (> 5MB)", func(t *testing.T) {
		deps := newMockDeps(t)
		svc := deps.newService()

		user, err := svc.UploadUserAvatar(
			context.Background(),
			userID,
			"huge.png",
			"image/png",
			bytes.NewReader([]byte("x")),
			6*1024*1024,
		)
		require.ErrorIs(t, err, domainWorkspace.ErrAvatarTooLarge)
		assert.Nil(t, user)
	})

	t.Run("invalid image format rejects", func(t *testing.T) {
		deps := newMockDeps(t)
		svc := deps.newService()

		user, err := svc.UploadUserAvatar(
			context.Background(),
			userID,
			"file.pdf",
			"application/pdf",
			bytes.NewReader([]byte("%PDF")),
			100,
		)
		require.ErrorIs(t, err, domainWorkspace.ErrInvalidAvatarFormat)
		assert.Nil(t, user)
	})
}
