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
				assert.Nil(t, tag.ProjectID)
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

func TestNewScopedTag(t *testing.T) {
	wsID := uuid.New()
	projectID := uuid.New()

	t.Run("creates tag with valid project scope", func(t *testing.T) {
		tag, err := NewScopedTag(wsID, &projectID, "backend-dev", "#3B82F6")
		require.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, wsID, tag.WorkspaceID)
		require.NotNil(t, tag.ProjectID)
		assert.Equal(t, projectID, *tag.ProjectID)
		assert.Equal(t, "backend-dev", tag.Name)
		assert.Equal(t, "#3B82F6", tag.ColorHex)
	})

	t.Run("creates global tag when projectID is nil", func(t *testing.T) {
		tag, err := NewScopedTag(wsID, nil, "global-tag", "")
		require.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, wsID, tag.WorkspaceID)
		assert.Nil(t, tag.ProjectID)
		assert.Equal(t, "global-tag", tag.Name)
		assert.Equal(t, DefaultTagColor, tag.ColorHex)
	})

	t.Run("creates global tag when projectID is uuid.Nil", func(t *testing.T) {
		nilUUID := uuid.Nil
		tag, err := NewScopedTag(wsID, &nilUUID, "global-nil-tag", "")
		require.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, wsID, tag.WorkspaceID)
		assert.Nil(t, tag.ProjectID)
	})

	t.Run("returns error on invalid name", func(t *testing.T) {
		tag, err := NewScopedTag(wsID, &projectID, "   ", "")
		require.ErrorIs(t, err, ErrInvalidTagName)
		assert.Nil(t, tag)
	})

	t.Run("returns error on invalid color length", func(t *testing.T) {
		tag, err := NewScopedTag(wsID, &projectID, "tag", strings.Repeat("a", 33))
		require.ErrorIs(t, err, ErrInvalidTagColor)
		assert.Nil(t, tag)
	})
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

func TestTag_SetName(t *testing.T) {
	tag, err := NewTag(uuid.New(), "initial", "#ffffff")
	require.NoError(t, err)

	t.Run("valid name", func(t *testing.T) {
		err := tag.SetName("updated-name")
		require.NoError(t, err)
		assert.Equal(t, "updated-name", tag.Name)
	})

	t.Run("invalid empty name", func(t *testing.T) {
		err := tag.SetName("   ")
		require.ErrorIs(t, err, ErrInvalidTagName)
	})

	t.Run("invalid character in name", func(t *testing.T) {
		err := tag.SetName("invalid/name")
		require.ErrorIs(t, err, ErrInvalidTagName)
	})
}

func TestTag_SetColorHex(t *testing.T) {
	tag, err := NewTag(uuid.New(), "tag", "#ffffff")
	require.NoError(t, err)

	validCases := []string{
		"#123",
		"#abc",
		"#ABC",
		"#123456",
		"#abcdef",
		"#ABCDEF",
		"#1a2B3c",
		"  #6366f1  ",
	}

	for _, color := range validCases {
		t.Run("valid_"+color, func(t *testing.T) {
			err := tag.SetColorHex(color)
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(color), tag.ColorHex)
		})
	}

	invalidCases := []string{
		"",
		"   ",
		"123456",
		"#12",
		"#1234",
		"#12345",
		"#1234567",
		"#ggg",
		"#12345z",
		"red",
		"#",
	}

	for _, color := range invalidCases {
		t.Run("invalid_"+color, func(t *testing.T) {
			err := tag.SetColorHex(color)
			require.ErrorIs(t, err, ErrInvalidTagColor)
		})
	}
}

func TestNewTagWithMetrics(t *testing.T) {
	wsID := uuid.New()
	baseTag, err := NewTag(wsID, "frontend", "#3b82f6")
	require.NoError(t, err)
	require.NotNil(t, baseTag)

	testCases := []struct {
		name          string
		inputTag      Tag
		inputCount    int64
		expectedCount int64
	}{
		{
			name:          "positive task count",
			inputTag:      *baseTag,
			inputCount:    5,
			expectedCount: 5,
		},
		{
			name:          "zero task count",
			inputTag:      *baseTag,
			inputCount:    0,
			expectedCount: 0,
		},
		{
			name:          "negative count clamped to zero",
			inputTag:      *baseTag,
			inputCount:    -3,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewTagWithMetrics(tc.inputTag, tc.inputCount)
			assert.Equal(t, tc.expectedCount, result.TaskCount)
			assert.Equal(t, tc.inputTag.ID, result.ID)
			assert.Equal(t, tc.inputTag.WorkspaceID, result.WorkspaceID)
			assert.Equal(t, tc.inputTag.Name, result.Name)
			assert.Equal(t, tc.inputTag.ColorHex, result.ColorHex)
			assert.Equal(t, tc.inputTag, result.Tag)
		})
	}
}



