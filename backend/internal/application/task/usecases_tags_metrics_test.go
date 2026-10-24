package task_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	taskapp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	domainTask "takt/internal/domain/task"
)

var errDBFailure = errors.New("database failure")

func TestService_ListWorkspaceTagsWithMetrics(t *testing.T) {
	wsID := uuid.New()
	tag1, _ := domainTask.NewTag(wsID, "Backend", "#3B82F6")
	tag2, _ := domainTask.NewTag(wsID, "Frontend", "#10B981")

	m1 := domainTask.NewTagWithMetrics(*tag1, 5)
	m2 := domainTask.NewTagWithMetrics(*tag2, 0)

	testCases := []struct {
		name         string
		workspaceID  uuid.UUID
		setupTagRepo func(m *mocks.MockTagRepository)
		checkOut     func(t *testing.T, metrics []*domainTask.TagWithMetrics)
		wantErr      error
	}{
		{
			name:        "успешное получение тегов с метриками",
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().
					ListWorkspaceTagsWithMetrics(mock.Anything, wsID).
					Return([]*domainTask.TagWithMetrics{&m1, &m2}, nil).
					Once()
			},
			checkOut: func(t *testing.T, metrics []*domainTask.TagWithMetrics) {
				require.Len(t, metrics, 2)
				assert.Equal(t, "Backend", metrics[0].Name)
				assert.Equal(t, int64(5), metrics[0].TaskCount)
				assert.Equal(t, "Frontend", metrics[1].Name)
				assert.Equal(t, int64(0), metrics[1].TaskCount)
			},
		},
		{
			name:        "ошибка репозитория при получении метрик",
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().
					ListWorkspaceTagsWithMetrics(mock.Anything, wsID).
					Return(nil, errDBFailure).
					Once()
			},
			wantErr: errDBFailure,
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
			res, err := svc.ListWorkspaceTagsWithMetrics(context.Background(), tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				if tc.checkOut != nil {
					tc.checkOut(t, res)
				}
			}
		})
	}
}

func TestService_DeleteUnusedTags(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name         string
		workspaceID  uuid.UUID
		setupTagRepo func(m *mocks.MockTagRepository)
		wantErr      error
	}{
		{
			name:        "успешное удаление неиспользуемых тегов",
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().
					DeleteUnusedTags(mock.Anything, wsID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name:        "ошибка БД при удалении неиспользуемых тегов",
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().
					DeleteUnusedTags(mock.Anything, wsID).
					Return(errDBFailure).
					Once()
			},
			wantErr: errDBFailure,
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
			err := svc.DeleteUnusedTags(context.Background(), tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
