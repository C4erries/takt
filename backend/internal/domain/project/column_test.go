package project

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewColumn(t *testing.T) {
	boardID := uuid.New()

	testCases := []struct {
		name           string
		boardID        uuid.UUID
		colName        string
		position       float64
		statusCategory StatusCategory
		wantErr        error
		verify         func(t *testing.T, col *Column)
	}{
		{
			name:           "valid column with status and position",
			boardID:        boardID,
			colName:        "В работе",
			position:       2000.0,
			statusCategory: StatusCategoryInProgress,
			wantErr:        nil,
			verify: func(t *testing.T, col *Column) {
				assert.Equal(t, "В работе", col.Name)
				assert.Equal(t, 2000.0, col.Position)
				assert.Equal(t, StatusCategoryInProgress, col.StatusCategory)
				assert.Equal(t, boardID, col.BoardID)
				assert.NotEqual(t, uuid.Nil, col.ID)
			},
		},
		{
			name:           "default status category when empty",
			boardID:        boardID,
			colName:        "Новая",
			position:       1000.0,
			statusCategory: "",
			wantErr:        nil,
			verify: func(t *testing.T, col *Column) {
				assert.Equal(t, StatusCategoryTodo, col.StatusCategory)
			},
		},
		{
			name:           "error on empty column name",
			boardID:        boardID,
			colName:        "   ",
			position:       1000.0,
			statusCategory: StatusCategoryTodo,
			wantErr:        ErrInvalidColumnName,
		},
		{
			name:           "error on column name exceeding 255 chars",
			boardID:        boardID,
			colName:        strings.Repeat("c", 256),
			position:       1000.0,
			statusCategory: StatusCategoryTodo,
			wantErr:        ErrInvalidColumnName,
		},
		{
			name:           "error on invalid status category",
			boardID:        boardID,
			colName:        "Колонка",
			position:       1000.0,
			statusCategory: StatusCategory("unknown"),
			wantErr:        ErrInvalidStatusCategory,
		},
		{
			name:           "error on zero or negative position",
			boardID:        boardID,
			colName:        "Колонка",
			position:       0.0,
			statusCategory: StatusCategoryTodo,
			wantErr:        ErrInvalidPosition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			col, err := NewColumn(tc.boardID, tc.colName, tc.position, tc.statusCategory)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, col)
			} else {
				require.NoError(t, err)
				require.NotNil(t, col)
				if tc.verify != nil {
					tc.verify(t, col)
				}
			}
		})
	}
}

func TestNewDefaultColumns(t *testing.T) {
	boardID := uuid.New()
	columns := NewDefaultColumns(boardID)

	require.Len(t, columns, 3)

	assert.Equal(t, DefaultColumnNameTodo, columns[0].Name)
	assert.Equal(t, DefaultPositionTodo, columns[0].Position)
	assert.Equal(t, StatusCategoryTodo, columns[0].StatusCategory)
	assert.Equal(t, boardID, columns[0].BoardID)

	assert.Equal(t, DefaultColumnNameInProgress, columns[1].Name)
	assert.Equal(t, DefaultPositionInProgress, columns[1].Position)
	assert.Equal(t, StatusCategoryInProgress, columns[1].StatusCategory)
	assert.Equal(t, boardID, columns[1].BoardID)

	assert.Equal(t, DefaultColumnNameDone, columns[2].Name)
	assert.Equal(t, DefaultPositionDone, columns[2].Position)
	assert.Equal(t, StatusCategoryDone, columns[2].StatusCategory)
	assert.Equal(t, boardID, columns[2].BoardID)
}

func TestColumn_Update(t *testing.T) {
	boardID := uuid.New()
	col, err := NewColumn(boardID, "К выполнению", 1000.0, StatusCategoryTodo)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		err := col.Update("В процессе", 1500.0, StatusCategoryInProgress)
		require.NoError(t, err)
		assert.Equal(t, "В процессе", col.Name)
		assert.Equal(t, 1500.0, col.Position)
		assert.Equal(t, StatusCategoryInProgress, col.StatusCategory)
	})

	t.Run("invalid status category in update", func(t *testing.T) {
		err := col.Update("", 0, StatusCategory("invalid"))
		require.ErrorIs(t, err, ErrInvalidStatusCategory)
	})

	t.Run("invalid name length in update", func(t *testing.T) {
		err := col.Update(strings.Repeat("x", 256), 0, "")
		require.ErrorIs(t, err, ErrInvalidColumnName)
	})
}

func TestColumn_MoveToBoard(t *testing.T) {
	initialBoardID := uuid.New()
	targetBoardID := uuid.New()

	testCases := []struct {
		name        string
		newBoardID  uuid.UUID
		newPosition float64
		wantErr     error
		verify      func(t *testing.T, col *Column)
	}{
		{
			name:        "successful move to target board",
			newBoardID:  targetBoardID,
			newPosition: 5000.0,
			wantErr:     nil,
			verify: func(t *testing.T, col *Column) {
				assert.Equal(t, targetBoardID, col.BoardID)
				assert.Equal(t, 5000.0, col.Position)
			},
		},
		{
			name:        "error on nil board ID",
			newBoardID:  uuid.Nil,
			newPosition: 5000.0,
			wantErr:     ErrInvalidBoardID,
			verify: func(t *testing.T, col *Column) {
				assert.Equal(t, initialBoardID, col.BoardID)
			},
		},
		{
			name:        "error on zero position",
			newBoardID:  targetBoardID,
			newPosition: 0.0,
			wantErr:     ErrInvalidPosition,
			verify: func(t *testing.T, col *Column) {
				assert.Equal(t, initialBoardID, col.BoardID)
				assert.Equal(t, 1000.0, col.Position)
			},
		},
		{
			name:        "error on negative position",
			newBoardID:  targetBoardID,
			newPosition: -100.0,
			wantErr:     ErrInvalidPosition,
			verify: func(t *testing.T, col *Column) {
				assert.Equal(t, initialBoardID, col.BoardID)
				assert.Equal(t, 1000.0, col.Position)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			col, err := NewColumn(initialBoardID, "К выполнению", 1000.0, StatusCategoryTodo)
			require.NoError(t, err)

			err = col.MoveToBoard(tc.newBoardID, tc.newPosition)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			if tc.verify != nil {
				tc.verify(t, col)
			}
		})
	}
}


