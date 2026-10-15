package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	workspaceApp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

func mapUserWorkspaceToGenerated(w *domainWorkspace.UserWorkspace) generated.Workspace {
	return generated.Workspace{
		Id:        w.ID,
		Name:      w.Name,
		Role:      generated.WorkspaceRole(w.Role),
		OwnerId:   w.OwnerID,
		CreatedAt: w.CreatedAt,
	}
}

func mapMemberInfoToGenerated(m *workspaceApp.WorkspaceMemberInfo) generated.WorkspaceMember {
	return generated.WorkspaceMember{
		Id:          m.ID,
		UserId:      m.UserID,
		Email:       openapi_types.Email(m.Email),
		DisplayName: m.DisplayName,
		Role:        generated.WorkspaceMemberRole(m.Role),
		CreatedAt:   m.CreatedAt,
	}
}

func mapInviteToGenerated(inv *domainWorkspace.WorkspaceInvite) generated.WorkspaceInvite {
	return generated.WorkspaceInvite{
		Id:          inv.ID,
		WorkspaceId: inv.WorkspaceID,
		Token:       inv.Token,
		InviteUrl:   "/invite/" + inv.Token,
		ExpiresAt:   inv.ExpiresAt,
		CreatedAt:   inv.CreatedAt,
	}
}

// ListWorkspaces lists all workspaces for current user.
func (h *Handler) ListWorkspaces(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	workspaces, err := h.workspaceService.ListWorkspaces(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to list workspaces"})
	}

	resp := make([]generated.Workspace, len(workspaces))
	for i, w := range workspaces {
		resp[i] = mapUserWorkspaceToGenerated(w)
	}

	return c.JSON(http.StatusOK, resp)
}

// CreateWorkspace creates a new workspace and sets the caller as owner.
func (h *Handler) CreateWorkspace(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	var req generated.CreateWorkspaceJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}

	ws, err := h.workspaceService.CreateWorkspace(c.Request().Context(), userID, req.Name)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInvalidWorkspaceName) {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "VALIDATION_ERROR", Message: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to create workspace"})
	}

	return c.JSON(http.StatusCreated, mapUserWorkspaceToGenerated(ws))
}

// UpdateWorkspace renames a workspace (owner only).
func (h *Handler) UpdateWorkspace(c echo.Context, workspaceId openapi_types.UUID) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	var req generated.UpdateWorkspaceJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}

	ws, err := h.workspaceService.UpdateWorkspace(c.Request().Context(), userID, workspaceId, req.Name)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInvalidWorkspaceName) {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "VALIDATION_ERROR", Message: err.Error()})
		}
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{Code: "FORBIDDEN", Message: "Only workspace owner can rename workspace"})
		}
		if errors.Is(err, domainWorkspace.ErrWorkspaceNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Workspace not found"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to update workspace"})
	}

	return c.JSON(http.StatusOK, mapUserWorkspaceToGenerated(ws))
}

// SwitchWorkspace switches active workspace and issues a new JWT.
func (h *Handler) SwitchWorkspace(c echo.Context, workspaceId openapi_types.UUID) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	res, err := h.workspaceService.SwitchWorkspace(c.Request().Context(), userID, workspaceId)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{Code: "FORBIDDEN", Message: "Access denied to workspace"})
		}
		if errors.Is(err, domainWorkspace.ErrWorkspaceNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Workspace not found"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to switch workspace"})
	}

	SetAuthCookie(c, res.Token, h.isProduction)

	return c.JSON(http.StatusOK, generated.WorkspaceSwitchResponse{
		Token:     res.Token,
		Workspace: mapUserWorkspaceToGenerated(res.Workspace),
	})
}

// ListWorkspaceMembers lists all members of the workspace.
func (h *Handler) ListWorkspaceMembers(c echo.Context, workspaceId openapi_types.UUID) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	members, err := h.workspaceService.ListMembers(c.Request().Context(), userID, workspaceId)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{Code: "FORBIDDEN", Message: "Access denied to workspace"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to list members"})
	}

	resp := make([]generated.WorkspaceMember, len(members))
	for i, m := range members {
		resp[i] = mapMemberInfoToGenerated(m)
	}

	return c.JSON(http.StatusOK, resp)
}

// RemoveWorkspaceMember removes a member or leaves the workspace.
func (h *Handler) RemoveWorkspaceMember(c echo.Context, workspaceId openapi_types.UUID, userId openapi_types.UUID) error {
	actorID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	err = h.workspaceService.RemoveMember(c.Request().Context(), actorID, workspaceId, userId)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrCannotRemoveOwner) {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "CANNOT_REMOVE_OWNER", Message: "Cannot remove workspace owner"})
		}
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{Code: "FORBIDDEN", Message: "Access denied"})
		}
		if errors.Is(err, domainWorkspace.ErrWorkspaceNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Member or workspace not found"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to remove member"})
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateWorkspaceInvite generates or returns an active invite link for workspace.
func (h *Handler) CreateWorkspaceInvite(c echo.Context, workspaceId openapi_types.UUID) error {
	actorID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	invite, err := h.workspaceService.CreateInvite(c.Request().Context(), actorID, workspaceId)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{Code: "FORBIDDEN", Message: "Only workspace owner can generate invites"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to create invite"})
	}

	return c.JSON(http.StatusCreated, mapInviteToGenerated(invite))
}

// GetInvitePreview retrieves public preview data for an invite token.
func (h *Handler) GetInvitePreview(c echo.Context, token string) error {
	preview, err := h.workspaceService.GetInvitePreview(c.Request().Context(), token)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInviteNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Invite not found"})
		}
		if errors.Is(err, domainWorkspace.ErrInviteExpired) {
			return c.JSON(http.StatusGone, generated.Error{Code: "INVITE_EXPIRED", Message: "Invite has expired"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to get invite preview"})
	}

	return c.JSON(http.StatusOK, generated.InvitePreview{
		Token:         preview.Token,
		WorkspaceName: preview.WorkspaceName,
		InvitedByName: preview.InvitedByName,
		ExpiresAt:     preview.ExpiresAt,
	})
}

// AcceptInvite accepts an invite by token and joins the workspace.
func (h *Handler) AcceptInvite(c echo.Context, token string) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: "Unauthorized"})
	}

	res, err := h.workspaceService.AcceptInvite(c.Request().Context(), userID, token)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInviteNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Invite not found"})
		}
		if errors.Is(err, domainWorkspace.ErrInviteExpired) {
			return c.JSON(http.StatusGone, generated.Error{Code: "INVITE_EXPIRED", Message: "Invite has expired"})
		}
		if errors.Is(err, domainWorkspace.ErrAlreadyMember) {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "ALREADY_MEMBER", Message: "You are already a member of this workspace"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to accept invite"})
	}

	SetAuthCookie(c, res.Token, h.isProduction)

	return c.JSON(http.StatusOK, generated.WorkspaceSwitchResponse{
		Token:     res.Token,
		Workspace: mapUserWorkspaceToGenerated(res.Workspace),
	})
}
