package task

import (
	"context"

	"github.com/google/uuid"
	domainTask "takt/internal/domain/task"
)

func (s *Service) loadTaskTags(ctx context.Context, tasks []*domainTask.Task) {
	if s.tagRepo == nil {
		return
	}
	for _, t := range tasks {
		tags, err := s.tagRepo.GetTaskTags(ctx, t.ID)
		if err == nil && tags != nil {
			t.Tags = tags
		} else if t.Tags == nil {
			t.Tags = []*domainTask.Tag{}
		}
	}
}

func (s *Service) loadSingleTaskTags(ctx context.Context, t *domainTask.Task) {
	if s.tagRepo == nil || t == nil {
		return
	}
	tags, err := s.tagRepo.GetTaskTags(ctx, t.ID)
	if err == nil && tags != nil {
		t.Tags = tags
	} else if t.Tags == nil {
		t.Tags = []*domainTask.Tag{}
	}
}

func (s *Service) CreateTag(ctx context.Context, input CreateTagInput) (*domainTask.Tag, error) {
	tag, err := domainTask.NewScopedTag(input.WorkspaceID, input.ProjectID, input.Name, input.ColorHex)
	if err != nil {
		return nil, err
	}
	if err := s.tagRepo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *Service) ListWorkspaceTags(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) ([]*domainTask.Tag, error) {
	return s.tagRepo.ListWorkspaceTags(ctx, workspaceID, projectID)
}

func (s *Service) UpdateTag(ctx context.Context, input UpdateTagInput) (*domainTask.Tag, error) {
	tag, err := s.tagRepo.GetTagByID(ctx, input.TagID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		if err := tag.SetName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.ColorHex != nil {
		if err := tag.SetColorHex(*input.ColorHex); err != nil {
			return nil, err
		}
	}
	if err := s.tagRepo.UpdateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *Service) DeleteTag(ctx context.Context, tagID, workspaceID uuid.UUID) error {
	if _, err := s.tagRepo.GetTagByID(ctx, tagID, workspaceID); err != nil {
		return err
	}
	return s.tagRepo.DeleteTag(ctx, tagID, workspaceID)
}

func (s *Service) AddTagToTask(ctx context.Context, taskID, tagID, workspaceID uuid.UUID) error {
	if _, err := s.taskRepo.GetByID(ctx, taskID, workspaceID); err != nil {
		return err
	}
	tag, err := s.tagRepo.GetTagByID(ctx, tagID, workspaceID)
	if err != nil {
		return err
	}
	if err := s.tagRepo.AddTagToTask(ctx, taskID, tagID); err != nil {
		return err
	}
	actorID := getActorIDFromContext(ctx)
	activity, err := domainTask.NewTaskActivity(taskID, actorID, domainTask.ActivityTagAdded, map[string]interface{}{
		"tag_name":  tag.Name,
		"color_hex": tag.ColorHex,
	})
	if err == nil {
		_ = s.taskRepo.CreateActivity(ctx, activity)
	}
	return nil
}

func (s *Service) RemoveTagFromTask(ctx context.Context, taskID, tagID, workspaceID uuid.UUID) error {
	if _, err := s.taskRepo.GetByID(ctx, taskID, workspaceID); err != nil {
		return err
	}
	tag, err := s.tagRepo.GetTagByID(ctx, tagID, workspaceID)
	if err != nil {
		return err
	}
	if err := s.tagRepo.RemoveTagFromTask(ctx, taskID, tagID); err != nil {
		return err
	}
	actorID := getActorIDFromContext(ctx)
	activity, err := domainTask.NewTaskActivity(taskID, actorID, domainTask.ActivityTagRemoved, map[string]interface{}{
		"tag_name": tag.Name,
	})
	if err == nil {
		_ = s.taskRepo.CreateActivity(ctx, activity)
	}
	return nil
}

func (s *Service) checkWorkspaceAccess(ctx context.Context, workspaceID uuid.UUID) error {
	if s.accessChecker == nil {
		return nil
	}
	if val := ctx.Value("userID"); val != nil {
		if uid, ok := val.(uuid.UUID); ok && uid != uuid.Nil {
			return s.accessChecker.CheckAccess(ctx, workspaceID, uid)
		}
	}
	return nil
}

func (s *Service) ListWorkspaceTagsWithMetrics(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.TagWithMetrics, error) {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return nil, err
	}
	return s.tagRepo.ListWorkspaceTagsWithMetrics(ctx, workspaceID)
}

func (s *Service) DeleteUnusedTags(ctx context.Context, workspaceID uuid.UUID) error {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return err
	}
	return s.tagRepo.DeleteUnusedTags(ctx, workspaceID)
}

