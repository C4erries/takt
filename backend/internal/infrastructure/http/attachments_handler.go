package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	appAttachment "takt/internal/application/attachment"
	domainAttachment "takt/internal/domain/attachment"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

func mapAttachmentToGenerated(att *domainAttachment.Attachment) generated.Attachment {
	contentURL := fmt.Sprintf("/api/v1/workspaces/%s/attachments/%s/content", att.WorkspaceID.String(), att.ID.String())
	return generated.Attachment{
		Id:          att.ID,
		WorkspaceId: att.WorkspaceID,
		TaskId:      att.TaskID,
		FileName:    att.FileName,
		FileSize:    att.FileSize,
		ContentType: att.ContentType,
		ContentUrl:  contentURL,
		CreatedAt:   att.CreatedAt,
	}
}

// UploadTaskAttachment handles POST /workspaces/{workspace_id}/tasks/{task_id}/attachments.
func (h *Handler) UploadTaskAttachment(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	if h.attachmentService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Attachment service is unavailable",
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "A multipart 'file' field is required",
		})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Failed to open uploaded file",
		})
	}
	defer src.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	att, err := h.attachmentService.UploadAttachment(c.Request().Context(), appAttachment.UploadAttachmentInput{
		WorkspaceID:  workspaceId,
		TaskID:       taskId,
		UserID:       userID,
		FileName:     fileHeader.Filename,
		FileSize:     fileHeader.Size,
		ContentType:  contentType,
		Body:         src,
		MaxSizeBytes: h.maxAttachmentSize,
	})
	if err != nil {
		if errors.Is(err, appAttachment.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Task not found",
			})
		}
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{
				Code:    "FORBIDDEN",
				Message: "Access to workspace denied",
			})
		}
		if errors.Is(err, domainAttachment.ErrFileTooLarge) ||
			errors.Is(err, domainAttachment.ErrInvalidFileName) ||
			errors.Is(err, domainAttachment.ErrInvalidFileSize) {
			return c.JSON(http.StatusBadRequest, generated.Error{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to upload attachment",
		})
	}

	return c.JSON(http.StatusCreated, mapAttachmentToGenerated(att))
}

// ListTaskAttachments handles GET /workspaces/{workspace_id}/tasks/{task_id}/attachments.
func (h *Handler) ListTaskAttachments(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	if h.attachmentService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Attachment service is unavailable",
		})
	}

	items, err := h.attachmentService.ListTaskAttachments(c.Request().Context(), workspaceId, taskId, userID)
	if err != nil {
		if errors.Is(err, appAttachment.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Task not found",
			})
		}
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{
				Code:    "FORBIDDEN",
				Message: "Access to workspace denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list task attachments",
		})
	}

	res := make([]generated.Attachment, 0, len(items))
	for _, it := range items {
		res = append(res, mapAttachmentToGenerated(it))
	}

	return c.JSON(http.StatusOK, res)
}

// GetAttachmentContent handles GET /workspaces/{workspace_id}/attachments/{attachment_id}/content.
func (h *Handler) GetAttachmentContent(c echo.Context, workspaceId openapi_types.UUID, attachmentId openapi_types.UUID, params generated.GetAttachmentContentParams) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	if h.attachmentService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Attachment service is unavailable",
		})
	}

	out, err := h.attachmentService.GetAttachmentContent(c.Request().Context(), workspaceId, attachmentId, userID)
	if err != nil {
		if errors.Is(err, domainAttachment.ErrAttachmentNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Attachment not found",
			})
		}
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{
				Code:    "FORBIDDEN",
				Message: "Access to workspace denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve attachment content",
		})
	}
	defer out.Body.Close()

	c.Response().Header().Set("Cache-Control", "public, max-age=86400")

	disposition := "inline"
	if params.Download != nil && *params.Download {
		disposition = "attachment"
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", disposition, out.FileName))

	return c.Stream(http.StatusOK, out.ContentType, out.Body)
}

// DeleteAttachment handles DELETE /workspaces/{workspace_id}/attachments/{attachment_id}.
func (h *Handler) DeleteAttachment(c echo.Context, workspaceId openapi_types.UUID, attachmentId openapi_types.UUID) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	if h.attachmentService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Attachment service is unavailable",
		})
	}

	err = h.attachmentService.DeleteAttachment(c.Request().Context(), workspaceId, attachmentId, userID)
	if err != nil {
		if errors.Is(err, domainAttachment.ErrAttachmentNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Attachment not found",
			})
		}
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{
				Code:    "FORBIDDEN",
				Message: "Access to workspace denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete attachment",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

