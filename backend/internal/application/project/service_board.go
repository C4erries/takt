package project

import (
	"context"

	"github.com/google/uuid"

	domainProject "takt/internal/domain/project"
)

func (s *Service) GetProjectBoards(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainProject.Board, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, workspaceID); err != nil {
		return nil, err
	}
	return s.boardRepo.ListByProject(ctx, projectID)
}

func (s *Service) CreateBoard(ctx context.Context, projectID, workspaceID uuid.UUID, name string) (*domainProject.Board, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, workspaceID); err != nil {
		return nil, err
	}

	b, err := domainProject.NewBoard(projectID, name)
	if err != nil {
		return nil, err
	}

	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.boardRepo.Create(txCtx, b); err != nil {
			return err
		}

		cols := domainProject.NewDefaultColumns(b.ID)
		for _, col := range cols {
			if err := s.columnRepo.Create(txCtx, col); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *Service) UpdateBoard(ctx context.Context, input UpdateBoardInput) (*domainProject.Board, error) {
	if _, err := s.projectRepo.GetByID(ctx, input.ProjectID, input.WorkspaceID); err != nil {
		return nil, err
	}

	b, err := s.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		return nil, err
	}

	if b.ProjectID != input.ProjectID {
		return nil, domainProject.ErrBoardNotFound
	}

	if input.Name != nil {
		if err := b.Update(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Position != nil {
		b.UpdatePosition(*input.Position)
	}

	if err := s.boardRepo.Update(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *Service) DeleteBoard(ctx context.Context, projectID, boardID, workspaceID uuid.UUID) error {
	if _, err := s.projectRepo.GetByID(ctx, projectID, workspaceID); err != nil {
		return err
	}

	b, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	if b.ProjectID != projectID {
		return domainProject.ErrBoardNotFound
	}

	boards, err := s.boardRepo.ListByProject(ctx, projectID)
	if err != nil {
		return err
	}

	if len(boards) <= 1 {
		return domainProject.ErrLastBoard
	}

	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.taskRepo.ResetBoardTasksToInbox(txCtx, boardID, workspaceID); err != nil {
			return err
		}
		return s.boardRepo.Delete(txCtx, boardID)
	})
}

