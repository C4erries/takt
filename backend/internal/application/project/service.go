package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"takt/internal/domain/events"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
)

type Service struct {
	projectRepo   ProjectRepository
	boardRepo     BoardRepository
	columnRepo    ColumnRepository
	taskRepo      TaskMoveRepository
	accessChecker WorkspaceAccessChecker
	transactor    Transactor
	eventBus      EventPublisher
}

func (s *Service) SetEventPublisher(pub EventPublisher) {
	s.eventBus = pub
}

func NewService(
	projectRepo ProjectRepository,
	boardRepo BoardRepository,
	columnRepo ColumnRepository,
	taskRepo TaskMoveRepository,
	accessChecker WorkspaceAccessChecker,
	transactor Transactor,
) *Service {
	return &Service{
		projectRepo:   projectRepo,
		boardRepo:     boardRepo,
		columnRepo:    columnRepo,
		taskRepo:      taskRepo,
		accessChecker: accessChecker,
		transactor:    transactor,
	}
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (*domainProject.Project, error) {
	var createdProject *domainProject.Project

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		p, err := domainProject.NewProject(input.WorkspaceID, input.Name, input.Description, input.Color)
		if err != nil {
			return err
		}

		if err := s.projectRepo.Create(txCtx, p); err != nil {
			return err
		}

		board, err := domainProject.NewDefaultBoard(p.ID)
		if err != nil {
			return err
		}

		if err := s.boardRepo.Create(txCtx, board); err != nil {
			return err
		}

		cols := domainProject.NewDefaultColumns(board.ID)
		for _, col := range cols {
			if err := s.columnRepo.Create(txCtx, col); err != nil {
				return err
			}
		}

		createdProject = p
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdProject, nil
}

func (s *Service) GetProjects(ctx context.Context, workspaceID uuid.UUID) ([]*domainProject.Project, error) {
	return s.projectRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *Service) GetProject(ctx context.Context, projectID, workspaceID uuid.UUID) (*domainProject.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID, workspaceID)
}

func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (*domainProject.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, input.ProjectID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	if err := p.Update(input.Name, input.Description, input.Color); err != nil {
		return nil, err
	}

	if err := s.projectRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *Service) DeleteProject(ctx context.Context, projectID, workspaceID uuid.UUID) error {
	if _, err := s.projectRepo.GetByID(ctx, projectID, workspaceID); err != nil {
		return err
	}

	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.taskRepo.ResetProjectTasksToInbox(txCtx, projectID, workspaceID); err != nil {
			return err
		}
		return s.projectRepo.Delete(txCtx, projectID, workspaceID)
	})
}

func (s *Service) GetBoardColumns(ctx context.Context, boardID, workspaceID uuid.UUID) ([]*domainProject.Column, error) {
	b, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if _, err := s.projectRepo.GetByID(ctx, b.ProjectID, workspaceID); err != nil {
		return nil, err
	}

	return s.columnRepo.ListByBoard(ctx, boardID)
}

func (s *Service) CreateColumn(ctx context.Context, input CreateColumnInput) (*domainProject.Column, error) {
	b, err := s.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		return nil, err
	}

	if _, err := s.projectRepo.GetByID(ctx, b.ProjectID, input.WorkspaceID); err != nil {
		return nil, err
	}

	pos := input.Position
	if pos <= 0 {
		cols, err := s.columnRepo.ListByBoard(ctx, input.BoardID)
		if err == nil && len(cols) > 0 {
			pos = cols[len(cols)-1].Position + 1000.0
		} else {
			pos = 1000.0
		}
	}

	col, err := domainProject.NewColumn(input.BoardID, input.Name, pos, input.StatusCategory)
	if err != nil {
		return nil, err
	}

	if err := s.columnRepo.Create(ctx, col); err != nil {
		return nil, err
	}

	return col, nil
}

func (s *Service) UpdateColumn(ctx context.Context, input UpdateColumnInput) (*domainProject.Column, error) {
	col, err := s.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		return nil, err
	}

	b, err := s.boardRepo.GetByID(ctx, col.BoardID)
	if err != nil {
		return nil, err
	}

	if _, err := s.projectRepo.GetByID(ctx, b.ProjectID, input.WorkspaceID); err != nil {
		return nil, err
	}

	isMoved := false
	if input.BoardID != nil && *input.BoardID != col.BoardID {
		isMoved = true
		if *input.BoardID == uuid.Nil {
			return nil, domainProject.ErrInvalidBoardID
		}
		targetBoard, err := s.boardRepo.GetByID(ctx, *input.BoardID)
		if err != nil {
			return nil, err
		}
		if targetBoard.ProjectID != b.ProjectID {
			return nil, domainProject.ErrAccessDenied
		}

		newPos := col.Position
		if input.Position > 0 {
			newPos = input.Position
		}

		if err := col.MoveToBoard(*input.BoardID, newPos); err != nil {
			return nil, err
		}

		if err := s.columnRepo.MoveToBoard(ctx, col.ID, *input.BoardID, newPos); err != nil {
			return nil, err
		}
	}

	posForUpdate := input.Position
	if isMoved {
		posForUpdate = 0
	}

	inboxChanged := false
	if input.IsInbox != nil {
		inboxChanged = true
		if *input.IsInbox {
			if err := s.columnRepo.ResetBoardInboxColumns(ctx, col.BoardID); err != nil {
				return nil, err
			}
		}
		col.SetInbox(*input.IsInbox)
	}

	if input.Name != "" || posForUpdate > 0 || input.StatusCategory != "" {
		if err := col.Update(input.Name, posForUpdate, input.StatusCategory); err != nil {
			return nil, err
		}
	}

	if input.Name != "" || posForUpdate > 0 || input.StatusCategory != "" || inboxChanged {
		if err := s.columnRepo.Update(ctx, col); err != nil {
			return nil, err
		}
	}

	return col, nil
}

func (s *Service) DeleteColumn(ctx context.Context, boardID, columnID, workspaceID uuid.UUID) error {
	if boardID == uuid.Nil {
		col, err := s.columnRepo.GetByID(ctx, columnID)
		if err != nil {
			return err
		}
		boardID = col.BoardID
	}

	b, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	if _, err := s.projectRepo.GetByID(ctx, b.ProjectID, workspaceID); err != nil {
		return err
	}

	if _, err := s.columnRepo.GetByID(ctx, columnID); err != nil {
		return err
	}

	count, err := s.columnRepo.CountTasksInColumn(ctx, columnID)
	if err != nil {
		return err
	}
	if count > 0 {
		return domainProject.ErrColumnNotEmpty
	}

	return s.columnRepo.Delete(ctx, columnID)
}

func (s *Service) MoveTask(ctx context.Context, input MoveTaskInput) (*domainTask.Task, error) {
	col, err := s.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		return nil, err
	}

	b, err := s.boardRepo.GetByID(ctx, col.BoardID)
	if err != nil {
		return nil, err
	}

	if _, err := s.projectRepo.GetByID(ctx, b.ProjectID, input.WorkspaceID); err != nil {
		return nil, err
	}

	var prevColID *uuid.UUID
	var fromColumnName string
	currentTask, err := s.taskRepo.GetTaskByID(ctx, input.TaskID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if currentTask != nil && currentTask.ColumnID != nil {
		prevColID = currentTask.ColumnID
		if prevCol, err := s.columnRepo.GetByID(ctx, *prevColID); err == nil && prevCol != nil {
			fromColumnName = prevCol.Name
		}
	}

	projID := b.ProjectID
	moved, err := s.taskRepo.MoveTaskToColumnAndPosition(
		ctx,
		input.TaskID,
		input.WorkspaceID,
		&col.ID,
		&projID,
		input.Position,
	)
	if err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.TaskMovedEvent{
			TaskID:           moved.ID,
			WorkspaceIDVal:   moved.WorkspaceID,
			ProjectID:        moved.ProjectID,
			ColumnID:         moved.ColumnID,
			PreviousColumnID: prevColID,
			FromColumnName:   fromColumnName,
			ToColumnName:     col.Name,
			Position:         float64(moved.Position),
			OccurredAtVal:    time.Now().UTC(),
			FromAutomation:   events.IsAutomationTriggered(ctx),
		})
	}

	return moved, nil
}

func (s *Service) GetProjectTasks(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, workspaceID); err != nil {
		return nil, err
	}
	return s.taskRepo.ListTasksByProject(ctx, projectID, workspaceID)
}
