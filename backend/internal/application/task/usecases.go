package task

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

type UseCases interface {
	GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error)
	CreateInboxTask(ctx context.Context, input CreateInboxTaskInput) (*domainTask.Task, error)
	ToggleTask(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error)
	DeleteTask(ctx context.Context, id, workspaceID uuid.UUID) error
	CreateInitialInboxTasks(ctx context.Context, workspaceID uuid.UUID) error
	GetTodayTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error)
	UpdateTask(ctx context.Context, input UpdateTaskInput) (*domainTask.Task, error)
	SetTaskFocus(ctx context.Context, input SetFocusInput) (*domainTask.Task, error)
	ListSubtasks(ctx context.Context, workspaceID, parentID uuid.UUID) ([]*domainTask.Task, error)
	CreateTag(ctx context.Context, input CreateTagInput) (*domainTask.Tag, error)
	ListWorkspaceTags(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Tag, error)
	ListWorkspaceTagsWithMetrics(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.TagWithMetrics, error)
	UpdateTag(ctx context.Context, input UpdateTagInput) (*domainTask.Tag, error)
	DeleteTag(ctx context.Context, tagID, workspaceID uuid.UUID) error
	DeleteUnusedTags(ctx context.Context, workspaceID uuid.UUID) error
	AddTagToTask(ctx context.Context, taskID, tagID, workspaceID uuid.UUID) error
	RemoveTagFromTask(ctx context.Context, taskID, tagID, workspaceID uuid.UUID) error
}

type Service struct {
	taskRepo      TaskRepository
	tagRepo       TagRepository
	accessChecker WorkspaceAccessChecker
	eventBus      EventPublisher
}

func NewService(taskRepo TaskRepository, tagRepo TagRepository, accessChecker WorkspaceAccessChecker, publisher ...EventPublisher) *Service {
	var pub EventPublisher
	if len(publisher) > 0 {
		pub = publisher[0]
	}
	return &Service{
		taskRepo:      taskRepo,
		tagRepo:       tagRepo,
		accessChecker: accessChecker,
		eventBus:      pub,
	}
}

func (s *Service) SetEventPublisher(pub EventPublisher) {
	s.eventBus = pub
}

func (s *Service) GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	tasks, err := s.taskRepo.GetInboxTasks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	s.loadTaskTags(ctx, tasks)
	return tasks, nil
}

func (s *Service) CreateInboxTask(ctx context.Context, input CreateInboxTaskInput) (*domainTask.Task, error) {
	var parent *domainTask.Task
	if input.ParentID != nil {
		p, err := s.taskRepo.GetByID(ctx, *input.ParentID, input.WorkspaceID)
		if err != nil {
			return nil, err
		}
		parent = p
	}

	maxPos, err := s.taskRepo.GetMaxInboxPosition(ctx, input.WorkspaceID)
	if err != nil {
		maxPos = -1
	}

	t, err := domainTask.NewInboxTask(
		input.WorkspaceID,
		input.Title,
		input.DescriptionMarkdown,
		input.Priority,
		input.DueDate,
		maxPos+1,
	)
	if err != nil {
		return nil, err
	}

	if parent != nil {
		if err := t.SetParent(input.ParentID); err != nil {
			return nil, err
		}
		t.ProjectID = parent.ProjectID
		t.ColumnID = parent.ColumnID
	}

	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	t.Tags = []*domainTask.Tag{}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.TaskCreatedEvent{
			TaskID:         t.ID,
			WorkspaceIDVal: t.WorkspaceID,
			ParentID:       t.ParentID,
			ProjectID:      t.ProjectID,
			ColumnID:       t.ColumnID,
			Title:          t.Title,
			DueDate:        t.DueDate,
			OccurredAtVal:  time.Now().UTC(),
			FromAutomation: events.IsAutomationTriggered(ctx),
		})
	}

	return t, nil
}

func (s *Service) ToggleTask(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error) {
	t, err := s.taskRepo.Toggle(ctx, id, workspaceID)
	if err != nil {
		return nil, err
	}
	s.loadSingleTaskTags(ctx, t)

	if s.eventBus != nil {
		if t.IsCompleted {
			completedAt := time.Now().UTC()
			if t.CompletedAt != nil {
				completedAt = *t.CompletedAt
			}
			_ = s.eventBus.Publish(ctx, events.TaskCompletedEvent{
				TaskID:         t.ID,
				WorkspaceIDVal: t.WorkspaceID,
				ParentID:       t.ParentID,
				ProjectID:      t.ProjectID,
				ColumnID:       t.ColumnID,
				CompletedAt:    completedAt,
				OccurredAtVal:  time.Now().UTC(),
				FromAutomation: events.IsAutomationTriggered(ctx),
			})
		} else {
			_ = s.eventBus.Publish(ctx, events.TaskUpdatedEvent{
				TaskID:         t.ID,
				WorkspaceIDVal: t.WorkspaceID,
				ParentID:       t.ParentID,
				ProjectID:      t.ProjectID,
				ColumnID:       t.ColumnID,
				Title:          t.Title,
				DueDate:        t.DueDate,
				IsCompleted:    false,
				OccurredAtVal:  time.Now().UTC(),
				FromAutomation: events.IsAutomationTriggered(ctx),
			})
		}
	}

	return t, nil
}

func (s *Service) DeleteTask(ctx context.Context, id, workspaceID uuid.UUID) error {
	return s.taskRepo.Delete(ctx, id, workspaceID)
}

func (s *Service) CreateInitialInboxTasks(ctx context.Context, workspaceID uuid.UUID) error {
	starterTitles := []string{
		"Нажми J и K для клавиатурного перемещения по строкам",
		"Нажми Пробел или кликни круглый чекбокс, чтобы завершить задачу",
		"Нажми C в любом месте экрана для мгновенного сброса мысли (Quick Capture)",
	}

	for i, title := range starterTitles {
		t, err := domainTask.NewInboxTask(workspaceID, title, "", domainTask.PriorityNone, nil, i)
		if err != nil {
			return err
		}
		if err := s.taskRepo.Create(ctx, t); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetTodayTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	tasks, err := s.taskRepo.GetTodayTasks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	s.loadTaskTags(ctx, tasks)
	return tasks, nil
}

func (s *Service) ListSubtasks(ctx context.Context, workspaceID, parentID uuid.UUID) ([]*domainTask.Task, error) {
	if _, err := s.taskRepo.GetByID(ctx, parentID, workspaceID); err != nil {
		return nil, err
	}
	tasks, err := s.taskRepo.ListSubtasks(ctx, parentID)
	if err != nil {
		return nil, err
	}
	s.loadTaskTags(ctx, tasks)
	return tasks, nil
}

func (s *Service) UpdateTask(ctx context.Context, input UpdateTaskInput) (*domainTask.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, input.TaskID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	prevProj := t.ProjectID
	prevCol := t.ColumnID

	if input.Title != nil {
		trimmed := strings.TrimSpace(*input.Title)
		if trimmed == "" {
			return nil, domainTask.ErrInvalidTaskTitle
		}
		t.Title = trimmed
	}
	if input.DescriptionMarkdown != nil {
		t.DescriptionMarkdown = *input.DescriptionMarkdown
	}
	if input.ClearDueDate {
		t.DueDate = nil
	} else if input.DueDate != nil {
		t.DueDate = input.DueDate
	}
	if input.Priority != nil {
		if !input.Priority.IsValid() {
			return nil, domainTask.ErrInvalidPriority
		}
		t.Priority = *input.Priority
	}
	if input.ClearFocusOrder {
		_ = t.SetFocusOrder(nil)
	} else if input.FocusOrder != nil {
		if err := t.SetFocusOrder(input.FocusOrder); err != nil {
			return nil, err
		}
	}
	if input.ClearProjectID {
		t.ProjectID = nil
	} else if input.ProjectID != nil {
		t.ProjectID = input.ProjectID
	}
	if input.ClearColumnID {
		t.ColumnID = nil
	} else if input.ColumnID != nil {
		t.ColumnID = input.ColumnID
	}
	if input.ClearParentID {
		_ = t.SetParent(nil)
	} else if input.ParentID != nil {
		parent, err := s.taskRepo.GetByID(ctx, *input.ParentID, input.WorkspaceID)
		if err != nil {
			return nil, err
		}
		if err := t.SetParent(input.ParentID); err != nil {
			return nil, err
		}
		if t.ProjectID == nil && parent.ProjectID != nil {
			t.ProjectID = parent.ProjectID
		}
		if t.ColumnID == nil && parent.ColumnID != nil {
			t.ColumnID = parent.ColumnID
		}
	}

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	s.loadSingleTaskTags(ctx, t)

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.TaskUpdatedEvent{
			TaskID:            t.ID,
			WorkspaceIDVal:    t.WorkspaceID,
			ParentID:          t.ParentID,
			ProjectID:         t.ProjectID,
			ColumnID:          t.ColumnID,
			Title:             t.Title,
			DueDate:           t.DueDate,
			IsCompleted:       t.IsCompleted,
			PreviousProjectID: prevProj,
			PreviousColumnID:  prevCol,
			OccurredAtVal:     time.Now().UTC(),
			FromAutomation:    events.IsAutomationTriggered(ctx),
		})
	}

	return t, nil
}

func (s *Service) SetTaskFocus(ctx context.Context, input SetFocusInput) (*domainTask.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, input.TaskID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	if err := t.SetFocusOrder(input.FocusOrder); err != nil {
		return nil, err
	}

	updated, err := s.taskRepo.SetFocusOrder(ctx, input.TaskID, input.WorkspaceID, input.FocusOrder)
	if err != nil {
		return nil, err
	}

	s.loadSingleTaskTags(ctx, updated)

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.TaskUpdatedEvent{
			TaskID:         updated.ID,
			WorkspaceIDVal: updated.WorkspaceID,
			ParentID:       updated.ParentID,
			ProjectID:      updated.ProjectID,
			ColumnID:       updated.ColumnID,
			Title:          updated.Title,
			DueDate:        updated.DueDate,
			IsCompleted:    updated.IsCompleted,
			OccurredAtVal:  time.Now().UTC(),
			FromAutomation: events.IsAutomationTriggered(ctx),
		})
	}

	return updated, nil
}
