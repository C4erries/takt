package project_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	projectapp "takt/internal/application/project"
	domainProject "takt/internal/domain/project"
)

func TestService_CreateBoard(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()

	existingProj := &domainProject.Project{ID: projID, WorkspaceID: wsID, Name: "Проект"}

	testCases := []struct {
		name      string
		projID    uuid.UUID
		wsID      uuid.UUID
		boardName string
		setupMock func(d mockDeps)
		wantErr   error
		verify    func(t *testing.T, b *domainProject.Board)
	}{
		{
			name:      "successful board creation with 3 default columns in transaction",
			projID:    projID,
			wsID:      wsID,
			boardName: "Спринт 1",
			setupMock: func(d mockDeps) {
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
				d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				).Once()
				d.boardRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(b *domainProject.Board) bool {
					return b.Name == "Спринт 1" && b.ProjectID == projID
				})).Return(nil).Once()
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == domainProject.DefaultColumnNameTodo && c.Position == domainProject.DefaultPositionTodo
				})).Return(nil).Once()
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == domainProject.DefaultColumnNameInProgress && c.Position == domainProject.DefaultPositionInProgress
				})).Return(nil).Once()
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == domainProject.DefaultColumnNameDone && c.Position == domainProject.DefaultPositionDone
				})).Return(nil).Once()
			},
			wantErr: nil,
			verify: func(t *testing.T, b *domainProject.Board) {
				assert.Equal(t, "Спринт 1", b.Name)
				assert.Equal(t, projID, b.ProjectID)
			},
		},
		{
			name:      "project not found error",
			projID:    projID,
			wsID:      wsID,
			boardName: "Спринт 2",
			setupMock: func(d mockDeps) {
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(nil, domainProject.ErrProjectNotFound).Once()
			},
			wantErr: domainProject.ErrProjectNotFound,
		},
		{
			name:      "invalid board name error",
			projID:    projID,
			wsID:      wsID,
			boardName: "   ",
			setupMock: func(d mockDeps) {
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
			},
			wantErr: domainProject.ErrInvalidBoardName,
		},
		{
			name:      "transactor error propagates",
			projID:    projID,
			wsID:      wsID,
			boardName: "Спринт 3",
			setupMock: func(d mockDeps) {
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
				d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).Return(errDB).Once()
			},
			wantErr: errDB,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := newMockDeps(t)
			tc.setupMock(deps)
			svc := deps.newService()

			b, err := svc.CreateBoard(context.Background(), tc.projID, tc.wsID, tc.boardName)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, b)
				if tc.verify != nil {
					tc.verify(t, b)
				}
			}
		})
	}
}

func TestService_UpdateColumn(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	boardID1 := uuid.New()
	boardID2 := uuid.New()
	colID := uuid.New()

	existingProj := &domainProject.Project{ID: projID, WorkspaceID: wsID, Name: "Проект"}
	existingBoard1 := &domainProject.Board{ID: boardID1, ProjectID: projID, Name: "Доска 1"}
	existingBoard2 := &domainProject.Board{ID: boardID2, ProjectID: projID, Name: "Доска 2"}
	foreignProjID := uuid.New()
	foreignBoard := &domainProject.Board{ID: uuid.New(), ProjectID: foreignProjID, Name: "Чужая доска"}

	testCases := []struct {
		name      string
		input     projectapp.UpdateColumnInput
		setupMock func(d mockDeps)
		wantErr   error
		verify    func(t *testing.T, col *domainProject.Column)
	}{
		{
			name: "update column name and category without moving board",
			input: projectapp.UpdateColumnInput{
				WorkspaceID:    wsID,
				ColumnID:       colID,
				Name:           "Новое имя",
				StatusCategory: domainProject.StatusCategoryInProgress,
			},
			setupMock: func(d mockDeps) {
				col := &domainProject.Column{ID: colID, BoardID: boardID1, Name: "Старое", Position: 1000.0, StatusCategory: domainProject.StatusCategoryTodo}
				d.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(col, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(existingBoard1, nil).Once()
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
				d.columnRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == "Новое имя" && c.StatusCategory == domainProject.StatusCategoryInProgress
				})).Return(nil).Once()
			},
			wantErr: nil,
			verify: func(t *testing.T, col *domainProject.Column) {
				assert.Equal(t, "Новое имя", col.Name)
				assert.Equal(t, domainProject.StatusCategoryInProgress, col.StatusCategory)
				assert.Equal(t, boardID1, col.BoardID)
			},
		},
		{
			name: "move column to another board in same project",
			input: projectapp.UpdateColumnInput{
				WorkspaceID: wsID,
				ColumnID:    colID,
				BoardID:     &boardID2,
				Position:    2500.0,
			},
			setupMock: func(d mockDeps) {
				col := &domainProject.Column{ID: colID, BoardID: boardID1, Name: "Колонка", Position: 1000.0, StatusCategory: domainProject.StatusCategoryTodo}
				d.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(col, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(existingBoard1, nil).Once()
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID2).Return(existingBoard2, nil).Once()
				d.columnRepo.EXPECT().MoveToBoard(mock.Anything, colID, boardID2, 2500.0).Return(nil).Once()
			},
			wantErr: nil,
			verify: func(t *testing.T, col *domainProject.Column) {
				assert.Equal(t, boardID2, col.BoardID)
				assert.Equal(t, 2500.0, col.Position)
			},
		},
		{
			name: "error on moving column to foreign project board",
			input: projectapp.UpdateColumnInput{
				WorkspaceID: wsID,
				ColumnID:    colID,
				BoardID:     &foreignBoard.ID,
			},
			setupMock: func(d mockDeps) {
				col := &domainProject.Column{ID: colID, BoardID: boardID1, Name: "Колонка", Position: 1000.0, StatusCategory: domainProject.StatusCategoryTodo}
				d.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(col, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(existingBoard1, nil).Once()
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, foreignBoard.ID).Return(foreignBoard, nil).Once()
			},
			wantErr: domainProject.ErrAccessDenied,
		},
		{
			name: "error on nil board id",
			input: projectapp.UpdateColumnInput{
				WorkspaceID: wsID,
				ColumnID:    colID,
				BoardID:     func() *uuid.UUID { id := uuid.Nil; return &id }(),
			},
			setupMock: func(d mockDeps) {
				col := &domainProject.Column{ID: colID, BoardID: boardID1, Name: "Колонка", Position: 1000.0, StatusCategory: domainProject.StatusCategoryTodo}
				d.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(col, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(existingBoard1, nil).Once()
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
			},
			wantErr: domainProject.ErrInvalidBoardID,
		},
		{
			name: "error on target board not found",
			input: projectapp.UpdateColumnInput{
				WorkspaceID: wsID,
				ColumnID:    colID,
				BoardID:     &boardID2,
			},
			setupMock: func(d mockDeps) {
				col := &domainProject.Column{ID: colID, BoardID: boardID1, Name: "Колонка", Position: 1000.0, StatusCategory: domainProject.StatusCategoryTodo}
				d.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(col, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(existingBoard1, nil).Once()
				d.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID2).Return(nil, domainProject.ErrBoardNotFound).Once()
			},
			wantErr: domainProject.ErrBoardNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := newMockDeps(t)
			tc.setupMock(deps)
			svc := deps.newService()

			col, err := svc.UpdateColumn(context.Background(), tc.input)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
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

func TestService_Boards_CRUD(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	boardID1 := uuid.New()
	boardID2 := uuid.New()

	existingProj := &domainProject.Project{ID: projID, WorkspaceID: wsID, Name: "Проект"}
	board1 := &domainProject.Board{ID: boardID1, ProjectID: projID, Name: "Доска 1"}
	board2 := &domainProject.Board{ID: boardID2, ProjectID: projID, Name: "Доска 2"}

	t.Run("UpdateBoard успешно обновляет имя и позицию доски", func(t *testing.T) {
		deps := newMockDeps(t)
		bCopy := &domainProject.Board{ID: boardID1, ProjectID: projID, Name: "Доска 1", Position: 1000.0}
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(bCopy, nil).Once()
		deps.boardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(b *domainProject.Board) bool {
			return b.Name == "Новое название" && b.Position == 2500.0
		})).Return(nil).Once()

		svc := deps.newService()
		newName := "Новое название"
		newPos := 2500.0
		updated, err := svc.UpdateBoard(context.Background(), projectapp.UpdateBoardInput{
			WorkspaceID: wsID,
			ProjectID:   projID,
			BoardID:     boardID1,
			Name:        &newName,
			Position:    &newPos,
		})
		require.NoError(t, err)
		assert.Equal(t, "Новое название", updated.Name)
		assert.Equal(t, 2500.0, updated.Position)
	})

	t.Run("UpdateBoard обновляет только позицию", func(t *testing.T) {
		deps := newMockDeps(t)
		bCopy := &domainProject.Board{ID: boardID1, ProjectID: projID, Name: "Доска 1", Position: 1000.0}
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(bCopy, nil).Once()
		deps.boardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(b *domainProject.Board) bool {
			return b.Name == "Доска 1" && b.Position == 3000.0
		})).Return(nil).Once()

		svc := deps.newService()
		newPos := 3000.0
		updated, err := svc.UpdateBoard(context.Background(), projectapp.UpdateBoardInput{
			WorkspaceID: wsID,
			ProjectID:   projID,
			BoardID:     boardID1,
			Position:    &newPos,
		})
		require.NoError(t, err)
		assert.Equal(t, "Доска 1", updated.Name)
		assert.Equal(t, 3000.0, updated.Position)
	})

	t.Run("DeleteBoard отклоняется с ErrLastBoard, если доска единственная", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(board1, nil).Once()
		deps.boardRepo.EXPECT().ListByProject(mock.Anything, projID).Return([]*domainProject.Board{board1}, nil).Once()

		svc := deps.newService()
		err := svc.DeleteBoard(context.Background(), projID, boardID1, wsID)
		require.ErrorIs(t, err, domainProject.ErrLastBoard)
	})

	t.Run("DeleteBoard сбрасывает задачи в Inbox и удаляет доску", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID1).Return(board1, nil).Once()
		deps.boardRepo.EXPECT().ListByProject(mock.Anything, projID).Return([]*domainProject.Board{board1, board2}, nil).Once()
		deps.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		).Once()
		deps.taskRepo.EXPECT().ResetBoardTasksToInbox(mock.Anything, boardID1, wsID).Return(nil).Once()
		deps.boardRepo.EXPECT().Delete(mock.Anything, boardID1).Return(nil).Once()

		svc := deps.newService()
		err := svc.DeleteBoard(context.Background(), projID, boardID1, wsID)
		require.NoError(t, err)
	})
}
