package task

import (
	"context"
	"time"

	"github.com/google/uuid"
	domainTask "takt/internal/domain/task"
)

type CreateInboxTaskInput struct {
	WorkspaceID         uuid.UUID
	Title               string
	DescriptionMarkdown string
	Priority            domainTask.Priority
	DueDate             *time.Time
}

type UseCases interface {
	GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error)
	CreateInboxTask(ctx context.Context, input CreateInboxTaskInput) (*domainTask.Task, error)
	ToggleTask(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error)
	DeleteTask(ctx context.Context, id, workspaceID uuid.UUID) error
	CreateInitialInboxTasks(ctx context.Context, workspaceID uuid.UUID) error
}

type Service struct {
	taskRepo      TaskRepository
	accessChecker WorkspaceAccessChecker
}

func NewService(taskRepo TaskRepository, accessChecker WorkspaceAccessChecker) *Service {
	return &Service{
		taskRepo:      taskRepo,
		accessChecker: accessChecker,
	}
}

func (s *Service) GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	return s.taskRepo.GetInboxTasks(ctx, workspaceID)
}

func (s *Service) CreateInboxTask(ctx context.Context, input CreateInboxTaskInput) (*domainTask.Task, error) {
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

	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) ToggleTask(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error) {
	return s.taskRepo.Toggle(ctx, id, workspaceID)
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
