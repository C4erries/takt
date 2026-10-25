package project_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	projectapp "takt/internal/application/project"
	"takt/internal/application/project/mocks"
	"takt/internal/domain/events"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
)

var errDB = errors.New("database error")

type mockDeps struct {
	projectRepo   *mocks.MockProjectRepository
	boardRepo     *mocks.MockBoardRepository
	columnRepo    *mocks.MockColumnRepository
	taskRepo      *mocks.MockTaskMoveRepository
	accessChecker *mocks.MockWorkspaceAccessChecker
	transactor    *mocks.MockTransactor
}

func newMockDeps(t *testing.T) mockDeps {
	return mockDeps{
		projectRepo:   mocks.NewMockProjectRepository(t),
		boardRepo:     mocks.NewMockBoardRepository(t),
		columnRepo:    mocks.NewMockColumnRepository(t),
		taskRepo:      mocks.NewMockTaskMoveRepository(t),
		accessChecker: mocks.NewMockWorkspaceAccessChecker(t),
		transactor:    mocks.NewMockTransactor(t),
	}
}

func (d mockDeps) newService() *projectapp.Service {
	return projectapp.NewService(
		d.projectRepo,
		d.boardRepo,
		d.columnRepo,
		d.taskRepo,
		d.accessChecker,
		d.transactor,
	)
}

func TestService_CreateProject(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name      string
		input     projectapp.CreateProjectInput
		setupMock func(d mockDeps)
		wantErr   error
	}{
		{
			name: "успешное создание проекта с дефолтной доской и 3 GTD колонками в транзакции",
			input: projectapp.CreateProjectInput{
				WorkspaceID: wsID,
				Name:        "Редизайн UI",
				Description: "Новый интерфейс",
				Color:       "#6366f1",
			},
			setupMock: func(d mockDeps) {
				d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				).Once()
				d.projectRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *domainProject.Project) bool {
					return p.Name == "Редизайн UI" && p.WorkspaceID == wsID
				})).Return(nil).Once()
				d.boardRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(b *domainProject.Board) bool {
					return b.Name == "Основная"
				})).Return(nil).Once()
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == "К выполнению" && c.Position == 1000.0
				})).Return(nil).Once()
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == "В работе" && c.Position == 2000.0
				})).Return(nil).Once()
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
					return c.Name == "Готово" && c.Position == 3000.0
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "ошибка валидации имени проекта",
			input: projectapp.CreateProjectInput{
				WorkspaceID: wsID,
				Name:        "   ",
			},
			setupMock: func(d mockDeps) {
				d.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				).Once()
			},
			wantErr: domainProject.ErrInvalidProjectName,
		},
		{
			name: "сбой транзакции БД при сохранении проекта",
			input: projectapp.CreateProjectInput{
				WorkspaceID: wsID,
				Name:        "Проект",
			},
			setupMock: func(d mockDeps) {
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

			p, err := svc.CreateProject(context.Background(), tc.input)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.Equal(t, tc.input.Name, p.Name)
			}
		})
	}
}

func TestService_ProjectQueriesAndUpdates(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	existingProj := &domainProject.Project{
		ID:          projID,
		WorkspaceID: wsID,
		Name:        "Старое имя",
		Color:       "#000000",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	t.Run("GetProjects возвращает список проектов", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.projectRepo.EXPECT().ListByWorkspace(mock.Anything, wsID).Return([]*domainProject.Project{existingProj}, nil).Once()

		svc := deps.newService()
		list, err := svc.GetProjects(context.Background(), wsID)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("GetProject возвращает проект по ID", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()

		svc := deps.newService()
		p, err := svc.GetProject(context.Background(), projID, wsID)
		require.NoError(t, err)
		assert.Equal(t, existingProj.ID, p.ID)
	})

	t.Run("UpdateProject успешно обновляет данные", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.projectRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p *domainProject.Project) bool {
			return p.Name == "Новое имя" && p.Color == "#10b981"
		})).Return(nil).Once()

		svc := deps.newService()
		p, err := svc.UpdateProject(context.Background(), projectapp.UpdateProjectInput{
			ProjectID:   projID,
			WorkspaceID: wsID,
			Name:        "Новое имя",
			Color:       "#10b981",
		})
		require.NoError(t, err)
		assert.Equal(t, "Новое имя", p.Name)
	})

	t.Run("DeleteProject разлинковывает задачи в Inbox и удаляет проект в транзакции", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.transactor.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		).Once()
		deps.taskRepo.EXPECT().ResetProjectTasksToInbox(mock.Anything, projID, wsID).Return(nil).Once()
		deps.projectRepo.EXPECT().Delete(mock.Anything, projID, wsID).Return(nil).Once()

		svc := deps.newService()
		err := svc.DeleteProject(context.Background(), projID, wsID)
		require.NoError(t, err)
	})
}

func TestService_ColumnsAndSafetyGuard(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	boardID := uuid.New()
	colID := uuid.New()

	existingProj := &domainProject.Project{ID: projID, WorkspaceID: wsID, Name: "Проект"}
	existingBoard := &domainProject.Board{ID: boardID, ProjectID: projID, Name: "Основная"}
	existingCol := &domainProject.Column{
		ID:             colID,
		BoardID:        boardID,
		Name:           "К выполнению",
		Position:       1000.0,
		StatusCategory: domainProject.StatusCategoryTodo,
	}

	t.Run("CreateColumn успешно создает колонку", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
			return c.Name == "Тестирование" && c.Position == 2500.0
		})).Return(nil).Once()

		svc := deps.newService()
		col, err := svc.CreateColumn(context.Background(), projectapp.CreateColumnInput{
			WorkspaceID:    wsID,
			BoardID:        boardID,
			Name:           "Тестирование",
			Position:       2500.0,
			StatusCategory: domainProject.StatusCategoryInProgress,
		})
		require.NoError(t, err)
		assert.Equal(t, "Тестирование", col.Name)
	})

	t.Run("DeleteColumn отклоняется с ErrColumnNotEmpty, если в колонке есть задачи", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.columnRepo.EXPECT().CountTasksInColumn(mock.Anything, colID).Return(3, nil).Once()

		svc := deps.newService()
		err := svc.DeleteColumn(context.Background(), boardID, colID, wsID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domainProject.ErrColumnNotEmpty)
	})

	t.Run("DeleteColumn успешно удаляет пустую колонку", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.columnRepo.EXPECT().CountTasksInColumn(mock.Anything, colID).Return(0, nil).Once()
		deps.columnRepo.EXPECT().Delete(mock.Anything, colID).Return(nil).Once()

		svc := deps.newService()
		err := svc.DeleteColumn(context.Background(), boardID, colID, wsID)
		require.NoError(t, err)
	})
}

func TestService_MoveTask(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	boardID := uuid.New()
	fromColID := uuid.New()
	colID := uuid.New()
	taskID := uuid.New()

	existingProj := &domainProject.Project{ID: projID, WorkspaceID: wsID, Name: "Проект"}
	existingBoard := &domainProject.Board{ID: boardID, ProjectID: projID, Name: "Основная"}
	fromCol := &domainProject.Column{ID: fromColID, BoardID: boardID, Name: "Входящие", Position: 1000.0}
	existingCol := &domainProject.Column{ID: colID, BoardID: boardID, Name: "В работе", Position: 2000.0}

	currentTask := &domainTask.Task{
		ID:          taskID,
		WorkspaceID: wsID,
		ProjectID:   &projID,
		ColumnID:    &fromColID,
		Title:       "Сделать DND",
		Position:    1000.0,
	}

	movedTask := &domainTask.Task{
		ID:          taskID,
		WorkspaceID: wsID,
		ProjectID:   &projID,
		ColumnID:    &colID,
		Title:       "Сделать DND",
		Position:    1500.0,
	}

	t.Run("успешное перемещение задачи с заполнением названий колонок и публикацией события", func(t *testing.T) {
		deps := newMockDeps(t)
		mockPub := mocks.NewMockEventPublisher(t)
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.taskRepo.EXPECT().GetTaskByID(mock.Anything, taskID, wsID).Return(currentTask, nil).Once()
		deps.columnRepo.EXPECT().GetByID(mock.Anything, fromColID).Return(fromCol, nil).Once()
		deps.taskRepo.EXPECT().MoveTaskToColumnAndPosition(
			mock.Anything,
			taskID,
			wsID,
			&colID,
			&projID,
			1500.0,
		).Return(movedTask, nil).Once()

		mockPub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(evt events.Event) bool {
			e, ok := evt.(events.TaskMovedEvent)
			if !ok {
				return false
			}
			return e.TaskID == taskID &&
				e.FromColumnName == "Входящие" &&
				e.ToColumnName == "В работе" &&
				e.PreviousColumnID != nil && *e.PreviousColumnID == fromColID &&
				e.ColumnID != nil && *e.ColumnID == colID
		})).Return(nil).Once()

		svc := deps.newService()
		svc.SetEventPublisher(mockPub)
		res, err := svc.MoveTask(context.Background(), projectapp.MoveTaskInput{
			WorkspaceID: wsID,
			TaskID:      taskID,
			ColumnID:    colID,
			Position:    1500.0,
		})
		require.NoError(t, err)
		assert.Equal(t, movedTask.Position, res.Position)
		assert.Equal(t, &colID, res.ColumnID)
	})

	t.Run("успешное перемещение задачи без предыдущей колонки", func(t *testing.T) {
		deps := newMockDeps(t)
		mockPub := mocks.NewMockEventPublisher(t)
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		taskWithoutCol := &domainTask.Task{
			ID:          taskID,
			WorkspaceID: wsID,
			ProjectID:   &projID,
			ColumnID:    nil,
			Position:    1000.0,
		}
		deps.taskRepo.EXPECT().GetTaskByID(mock.Anything, taskID, wsID).Return(taskWithoutCol, nil).Once()
		deps.taskRepo.EXPECT().MoveTaskToColumnAndPosition(
			mock.Anything,
			taskID,
			wsID,
			&colID,
			&projID,
			1500.0,
		).Return(movedTask, nil).Once()

		mockPub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(evt events.Event) bool {
			e, ok := evt.(events.TaskMovedEvent)
			if !ok {
				return false
			}
			return e.TaskID == taskID &&
				e.FromColumnName == "" &&
				e.ToColumnName == "В работе" &&
				e.PreviousColumnID == nil &&
				e.ColumnID != nil && *e.ColumnID == colID
		})).Return(nil).Once()

		svc := deps.newService()
		svc.SetEventPublisher(mockPub)
		res, err := svc.MoveTask(context.Background(), projectapp.MoveTaskInput{
			WorkspaceID: wsID,
			TaskID:      taskID,
			ColumnID:    colID,
			Position:    1500.0,
		})
		require.NoError(t, err)
		assert.Equal(t, movedTask.Position, res.Position)
	})

	t.Run("ошибка если задача не найдена в репозитории", func(t *testing.T) {
		deps := newMockDeps(t)
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.taskRepo.EXPECT().GetTaskByID(mock.Anything, taskID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()

		svc := deps.newService()
		_, err := svc.MoveTask(context.Background(), projectapp.MoveTaskInput{
			WorkspaceID: wsID,
			TaskID:      taskID,
			ColumnID:    colID,
			Position:    1500.0,
		})
		require.ErrorIs(t, err, domainTask.ErrTaskNotFound)
	})
}

func TestService_UpdateColumn_Inbox(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	boardID := uuid.New()
	colID := uuid.New()

	existingProj := &domainProject.Project{ID: projID, WorkspaceID: wsID, Name: "Проект"}
	existingBoard := &domainProject.Board{ID: boardID, ProjectID: projID, Name: "Основная"}

	t.Run("установка isInbox=true сбрасывает другие колонки на доске", func(t *testing.T) {
		existingCol := &domainProject.Column{
			ID:             colID,
			BoardID:        boardID,
			Name:           "Входящие",
			Position:       1000.0,
			StatusCategory: domainProject.StatusCategoryTodo,
			IsInbox:        false,
		}

		deps := newMockDeps(t)
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.columnRepo.EXPECT().ResetBoardInboxColumns(mock.Anything, boardID).Return(nil).Once()
		deps.columnRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
			return c.ID == colID && c.IsInbox == true
		})).Return(nil).Once()

		svc := deps.newService()
		isInbox := true
		col, err := svc.UpdateColumn(context.Background(), projectapp.UpdateColumnInput{
			WorkspaceID: wsID,
			ColumnID:    colID,
			IsInbox:     &isInbox,
		})
		require.NoError(t, err)
		assert.True(t, col.IsInbox)
	})

	t.Run("снятие isInbox=false обновляет колонку без сброса", func(t *testing.T) {
		existingCol := &domainProject.Column{
			ID:             colID,
			BoardID:        boardID,
			Name:           "Входящие",
			Position:       1000.0,
			StatusCategory: domainProject.StatusCategoryTodo,
			IsInbox:        true,
		}

		deps := newMockDeps(t)
		deps.columnRepo.EXPECT().GetByID(mock.Anything, colID).Return(existingCol, nil).Once()
		deps.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil).Once()
		deps.projectRepo.EXPECT().GetByID(mock.Anything, projID, wsID).Return(existingProj, nil).Once()
		deps.columnRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *domainProject.Column) bool {
			return c.ID == colID && c.IsInbox == false
		})).Return(nil).Once()

		svc := deps.newService()
		isInbox := false
		col, err := svc.UpdateColumn(context.Background(), projectapp.UpdateColumnInput{
			WorkspaceID: wsID,
			ColumnID:    colID,
			IsInbox:     &isInbox,
		})
		require.NoError(t, err)
		assert.False(t, col.IsInbox)
	})
}

