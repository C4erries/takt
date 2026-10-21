package task_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	taskapp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	domainTask "takt/internal/domain/task"
)

func TestService_UpdateTag(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()
	errDB := errors.New("db error")

	newName := "updated-tag"
	newColor := "#10B981"
	invalidName := "invalid/name"
	invalidColor := "not-a-hex"

	testCases := []struct {
		name         string
		input        taskapp.UpdateTagInput
		setupTagRepo func(m *mocks.MockTagRepository)
		checkOut     func(t *testing.T, tag *domainTask.Tag)
		wantErr      error
	}{
		{
			name: "успешное обновление имени и цвета",
			input: taskapp.UpdateTagInput{
				TagID:       tagID,
				WorkspaceID: wsID,
				Name:        &newName,
				ColorHex:    &newColor,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				existing := &domainTask.Tag{
					ID:          tagID,
					WorkspaceID: wsID,
					Name:        "old-name",
					ColorHex:    "#000000",
					CreatedAt:   time.Now().UTC(),
				}
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(existing, nil).Once()
				m.EXPECT().UpdateTag(mock.Anything, mock.MatchedBy(func(tag *domainTask.Tag) bool {
					return tag.ID == tagID && tag.Name == newName && tag.ColorHex == newColor
				})).Return(nil).Once()
			},
			checkOut: func(t *testing.T, tag *domainTask.Tag) {
				assert.Equal(t, newName, tag.Name)
				assert.Equal(t, newColor, tag.ColorHex)
			},
		},
		{
			name: "успешное обновление только цвета",
			input: taskapp.UpdateTagInput{
				TagID:       tagID,
				WorkspaceID: wsID,
				ColorHex:    &newColor,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				existing := &domainTask.Tag{
					ID:          tagID,
					WorkspaceID: wsID,
					Name:        "keep-name",
					ColorHex:    "#000000",
					CreatedAt:   time.Now().UTC(),
				}
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(existing, nil).Once()
				m.EXPECT().UpdateTag(mock.Anything, mock.MatchedBy(func(tag *domainTask.Tag) bool {
					return tag.ID == tagID && tag.Name == "keep-name" && tag.ColorHex == newColor
				})).Return(nil).Once()
			},
			checkOut: func(t *testing.T, tag *domainTask.Tag) {
				assert.Equal(t, "keep-name", tag.Name)
				assert.Equal(t, newColor, tag.ColorHex)
			},
		},
		{
			name: "тег не найден",
			input: taskapp.UpdateTagInput{
				TagID:       tagID,
				WorkspaceID: wsID,
				Name:        &newName,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(nil, domainTask.ErrTagNotFound).Once()
			},
			wantErr: domainTask.ErrTagNotFound,
		},
		{
			name: "ошибка валидации имени",
			input: taskapp.UpdateTagInput{
				TagID:       tagID,
				WorkspaceID: wsID,
				Name:        &invalidName,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				existing := &domainTask.Tag{
					ID:          tagID,
					WorkspaceID: wsID,
					Name:        "old-name",
					ColorHex:    "#000000",
					CreatedAt:   time.Now().UTC(),
				}
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(existing, nil).Once()
			},
			wantErr: domainTask.ErrInvalidTagName,
		},
		{
			name: "ошибка валидации цвета HEX",
			input: taskapp.UpdateTagInput{
				TagID:       tagID,
				WorkspaceID: wsID,
				ColorHex:    &invalidColor,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				existing := &domainTask.Tag{
					ID:          tagID,
					WorkspaceID: wsID,
					Name:        "old-name",
					ColorHex:    "#000000",
					CreatedAt:   time.Now().UTC(),
				}
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(existing, nil).Once()
			},
			wantErr: domainTask.ErrInvalidTagColor,
		},
		{
			name: "ошибка базы данных при обновлении тега",
			input: taskapp.UpdateTagInput{
				TagID:       tagID,
				WorkspaceID: wsID,
				ColorHex:    &newColor,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				existing := &domainTask.Tag{
					ID:          tagID,
					WorkspaceID: wsID,
					Name:        "old-name",
					ColorHex:    "#000000",
					CreatedAt:   time.Now().UTC(),
				}
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(existing, nil).Once()
				m.EXPECT().UpdateTag(mock.Anything, mock.Anything).Return(errDB).Once()
			},
			wantErr: errDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(nil, tagRepo, nil)
			tag, err := svc.UpdateTag(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, tag)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tag)
				if tc.checkOut != nil {
					tc.checkOut(t, tag)
				}
			}
		})
	}
}
