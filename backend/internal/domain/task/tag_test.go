package task

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTag(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name        string
		tagName     string
		colorHex    string
		expectError error
		checkTag    func(t *testing.T, tag *Tag)
	}{
		{
			name:        "valid latin tag",
			tagName:     "focus",
			colorHex:    "#6366f1",
			expectError: nil,
			checkTag: func(t *testing.T, tag *Tag) {
				assert.Equal(t, "focus", tag.Name)
				assert.Equal(t, "#6366f1", tag.ColorHex)
				assert.Equal(t, wsID, tag.WorkspaceID)
				assert.NotEqual(t, uuid.Nil, tag.ID)
			},
		},
		{
			name:        "valid cyrillic tag with @ and spaces",
			tagName:     "@дом и уют",
			colorHex:    "",
			expectError: nil,
			checkTag: func(t *testing.T, tag *Tag) {
				assert.Equal(t, "@дом и уют", tag.Name)
				assert.Equal(t, DefaultTagColor, tag.ColorHex)
			},
		},
		{
			name:        "error on empty tag name",
			tagName:     "   ",
			colorHex:    "matcha",
			expectError: ErrInvalidTagName,
		},
		{
			name:        "error on tag name exceeding 64 chars",
			tagName:     strings.Repeat("a", 65),
			colorHex:    "matcha",
			expectError: ErrInvalidTagName,
		},
		{
			name:        "error on newline in tag name",
			tagName:     "tag\nname",
			colorHex:    "matcha",
			expectError: ErrInvalidTagName,
		},
		{
			name:        "error on slash in tag name",
			tagName:     "tag/name",
			colorHex:    "matcha",
			expectError: ErrInvalidTagName,
		},
		{
			name:        "error on color exceeding 32 chars",
			tagName:     "valid",
			colorHex:    strings.Repeat("c", 33),
			expectError: ErrInvalidTagColor,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag, err := NewTag(wsID, tc.tagName, tc.colorHex)
			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
				assert.Nil(t, tag)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tag)
				if tc.checkTag != nil {
					tc.checkTag(t, tag)
				}
			}
		})
	}
}

func TestTask_TagOperations(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "Test task", "", PriorityHigh, nil, 0)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Empty(t, task.Tags)

	tag1, err := NewTag(wsID, "focus", "matcha")
	require.NoError(t, err)
	tag2, err := NewTag(wsID, "work", "sakura")
	require.NoError(t, err)

	// 1. Add tag1
	err = task.AddTag(tag1)
	require.NoError(t, err)
	assert.Len(t, task.Tags, 1)
	assert.Equal(t, tag1.ID, task.Tags[0].ID)

	// 2. Add duplicate tag1 -> error
	err = task.AddTag(tag1)
	require.ErrorIs(t, err, ErrTagAlreadyAttached)
	assert.Len(t, task.Tags, 1)

	// 3. Add nil tag -> error
	err = task.AddTag(nil)
	require.ErrorIs(t, err, ErrTagNotFound)

	// 4. Add tag2
	err = task.AddTag(tag2)
	require.NoError(t, err)
	assert.Len(t, task.Tags, 2)

	// 5. Remove tag1
	removed := task.RemoveTag(tag1.ID)
	assert.True(t, removed)
	assert.Len(t, task.Tags, 1)
	assert.Equal(t, tag2.ID, task.Tags[0].ID)

	// 6. Remove non-existent tag -> false
	removed = task.RemoveTag(uuid.New())
	assert.False(t, removed)
	assert.Len(t, task.Tags, 1)
}

func TestTask_SetFocusOrder(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "Focus test task", "", PriorityLow, nil, 0)
	require.NoError(t, err)

	// 1. Set valid focus order
	order := 1
	err = task.SetFocusOrder(&order)
	require.NoError(t, err)
	require.NotNil(t, task.FocusOrder)
	assert.Equal(t, 1, *task.FocusOrder)

	// 2. Clear focus order (nil)
	err = task.SetFocusOrder(nil)
	require.NoError(t, err)
	assert.Nil(t, task.FocusOrder)

	// 3. Negative focus order -> error
	invalidOrder := -1
	err = task.SetFocusOrder(&invalidOrder)
	require.ErrorIs(t, err, ErrInvalidFocusOrder)
}

