package project

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProject(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name        string
		workspaceID uuid.UUID
		projName    string
		description string
		color       string
		wantErr     error
		verify      func(t *testing.T, p *Project)
	}{
		{
			name:        "valid project with default color",
			workspaceID: wsID,
			projName:    "Рефакторинг ядра",
			description: "Перевод на Clean Architecture",
			color:       "",
			wantErr:     nil,
			verify: func(t *testing.T, p *Project) {
				assert.Equal(t, "Рефакторинг ядра", p.Name)
				assert.Equal(t, "Перевод на Clean Architecture", p.Description)
				assert.Equal(t, DefaultColor, p.Color)
				assert.Equal(t, wsID, p.WorkspaceID)
				assert.NotEqual(t, uuid.Nil, p.ID)
				assert.False(t, p.CreatedAt.IsZero())
			},
		},
		{
			name:        "valid project with custom color",
			workspaceID: wsID,
			projName:    "Маркетинг",
			description: "",
			color:       "#10b981",
			wantErr:     nil,
			verify: func(t *testing.T, p *Project) {
				assert.Equal(t, "Маркетинг", p.Name)
				assert.Equal(t, "#10b981", p.Color)
			},
		},
		{
			name:        "error on empty project name",
			workspaceID: wsID,
			projName:    "   ",
			description: "",
			color:       "",
			wantErr:     ErrInvalidProjectName,
		},
		{
			name:        "error on project name exceeding 255 chars",
			workspaceID: wsID,
			projName:    strings.Repeat("a", 256),
			description: "",
			color:       "",
			wantErr:     ErrInvalidProjectName,
		},
		{
			name:        "error on color exceeding 32 chars",
			workspaceID: wsID,
			projName:    "Проект",
			description: "",
			color:       strings.Repeat("c", 33),
			wantErr:     ErrInvalidProjectColor,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProject(tc.workspaceID, tc.projName, tc.description, tc.color)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
				if tc.verify != nil {
					tc.verify(t, p)
				}
			}
		})
	}
}

func TestProject_Update(t *testing.T) {
	wsID := uuid.New()
	p, err := NewProject(wsID, "Старое имя", "Старое описание", "#000000")
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		err := p.Update("Новое имя", "Новое описание", "#ffffff")
		require.NoError(t, err)
		assert.Equal(t, "Новое имя", p.Name)
		assert.Equal(t, "Новое описание", p.Description)
		assert.Equal(t, "#ffffff", p.Color)
	})

	t.Run("empty name does not overwrite existing name", func(t *testing.T) {
		err := p.Update("", "Только описание", "")
		require.NoError(t, err)
		assert.Equal(t, "Новое имя", p.Name)
		assert.Equal(t, "Только описание", p.Description)
		assert.Equal(t, "#ffffff", p.Color)
	})

	t.Run("invalid name length returns error", func(t *testing.T) {
		err := p.Update(strings.Repeat("b", 256), "", "")
		require.ErrorIs(t, err, ErrInvalidProjectName)
	})

	t.Run("invalid color length returns error", func(t *testing.T) {
		err := p.Update("", "", strings.Repeat("x", 33))
		require.ErrorIs(t, err, ErrInvalidProjectColor)
	})
}

func TestBoard(t *testing.T) {
	projID := uuid.New()

	t.Run("create default board", func(t *testing.T) {
		b, err := NewDefaultBoard(projID)
		require.NoError(t, err)
		assert.Equal(t, DefaultBoardName, b.Name)
		assert.Equal(t, DefaultBoardPosition, b.Position)
		assert.Equal(t, projID, b.ProjectID)
		assert.NotEqual(t, uuid.Nil, b.ID)
	})

	t.Run("create custom board", func(t *testing.T) {
		b, err := NewBoard(projID, "Спринт 1")
		require.NoError(t, err)
		assert.Equal(t, "Спринт 1", b.Name)
		assert.Equal(t, DefaultBoardPosition, b.Position)
	})

	t.Run("error on empty board name", func(t *testing.T) {
		_, err := NewBoard(projID, "   ")
		require.ErrorIs(t, err, ErrInvalidBoardName)
	})

	t.Run("error on board name exceeding 255 chars", func(t *testing.T) {
		_, err := NewBoard(projID, strings.Repeat("a", 256))
		require.ErrorIs(t, err, ErrInvalidBoardName)
	})

	t.Run("update board name", func(t *testing.T) {
		b, err := NewBoard(projID, "Доска 1")
		require.NoError(t, err)

		err = b.Update("Доска 2")
		require.NoError(t, err)
		assert.Equal(t, "Доска 2", b.Name)

		err = b.Update("   ")
		require.ErrorIs(t, err, ErrInvalidBoardName)
	})

	t.Run("update board position", func(t *testing.T) {
		b, err := NewBoard(projID, "Доска 1")
		require.NoError(t, err)
		assert.Equal(t, DefaultBoardPosition, b.Position)

		b.UpdatePosition(2500.0)
		assert.Equal(t, 2500.0, b.Position)
	})
}

