package task_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/task"
)

func TestNewTaskActivity_Table(t *testing.T) {
	t.Parallel()

	validTaskID := uuid.New()
	validUserID := uuid.New()

	tests := []struct {
		name         string
		taskID       uuid.UUID
		userID       *uuid.UUID
		activityType string
		payload      map[string]interface{}
		wantErr      error
		check        func(t *testing.T, act *task.TaskActivity)
	}{
		{
			name:         "success with full payload and user id",
			taskID:       validTaskID,
			userID:       &validUserID,
			activityType: task.ActivityColumnMoved,
			payload: map[string]interface{}{
				"from_column": "Todo",
				"to_column":   "In Progress",
			},
			wantErr: nil,
			check: func(t *testing.T, act *task.TaskActivity) {
				assert.NotEqual(t, uuid.Nil, act.ID)
				assert.Equal(t, validTaskID, act.TaskID)
				require.NotNil(t, act.UserID)
				assert.Equal(t, validUserID, *act.UserID)
				assert.Equal(t, task.ActivityColumnMoved, act.ActivityType)
				assert.Equal(t, "Todo", act.Payload["from_column"])
				assert.WithinDuration(t, time.Now().UTC(), act.CreatedAt, 2*time.Second)
			},
		},
		{
			name:         "success with nil user id for system/automation action",
			taskID:       validTaskID,
			userID:       nil,
			activityType: task.ActivityStatusChanged,
			payload:      map[string]interface{}{"is_completed": true},
			wantErr:      nil,
			check: func(t *testing.T, act *task.TaskActivity) {
				assert.Nil(t, act.UserID)
				assert.Equal(t, task.ActivityStatusChanged, act.ActivityType)
			},
		},
		{
			name:         "success with nil payload initializes empty map",
			taskID:       validTaskID,
			userID:       &validUserID,
			activityType: task.ActivityDueDateCleared,
			payload:      nil,
			wantErr:      nil,
			check: func(t *testing.T, act *task.TaskActivity) {
				require.NotNil(t, act.Payload)
				assert.Empty(t, act.Payload)
			},
		},
		{
			name:         "fail on nil task id",
			taskID:       uuid.Nil,
			userID:       &validUserID,
			activityType: task.ActivityPriorityChanged,
			payload:      nil,
			wantErr:      task.ErrInvalidTaskID,
			check:        nil,
		},
		{
			name:         "fail on empty activity type",
			taskID:       validTaskID,
			userID:       &validUserID,
			activityType: "",
			payload:      nil,
			wantErr:      task.ErrInvalidActivityType,
			check:        nil,
		},
		{
			name:         "fail on unknown activity type",
			taskID:       validTaskID,
			userID:       &validUserID,
			activityType: "random_unknown_action",
			payload:      nil,
			wantErr:      task.ErrInvalidActivityType,
			check:        nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			act, err := task.NewTaskActivity(tc.taskID, tc.userID, tc.activityType, tc.payload)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				assert.Nil(t, act)
			} else {
				require.NoError(t, err)
				require.NotNil(t, act)
				if tc.check != nil {
					tc.check(t, act)
				}
			}
		})
	}
}

func TestIsValidActivityType_AllSupportedTypes(t *testing.T) {
	t.Parallel()

	expectedTypes := []string{
		task.ActivityColumnMoved,
		task.ActivityStatusChanged,
		task.ActivityPriorityChanged,
		task.ActivityDueDateSet,
		task.ActivityDueDateCleared,
		task.ActivityTagAdded,
		task.ActivityTagRemoved,
		task.ActivityTitleChanged,
		task.ActivitySubtaskAdded,
		task.ActivityTaskArchived,
		task.ActivityTaskUnarchived,
	}

	for _, typ := range expectedTypes {
		assert.True(t, task.IsValidActivityType(typ), "type %s should be valid", typ)
	}

	assert.False(t, task.IsValidActivityType("invalid_type"))
	assert.False(t, task.IsValidActivityType(""))
}
