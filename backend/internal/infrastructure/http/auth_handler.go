package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	workspaceApp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

func mapUserToGenerated(u *domainWorkspace.User) generated.User {
	return generated.User{
		Id:          u.ID,
		Email:       openapi_types.Email(u.Email),
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}

func mapWorkspaceToGenerated(w *domainWorkspace.UserWorkspace) generated.Workspace {
	return generated.Workspace{
		Id:        w.ID,
		Name:      w.Name,
		Type:      generated.WorkspaceType(w.Type),
		Role:      generated.WorkspaceRole(w.Role),
		CreatedAt: w.CreatedAt,
	}
}

func (h *Handler) Register(c echo.Context) error {
	var req generated.RegisterJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	email := strings.TrimSpace(string(req.Email))
	displayName := strings.TrimSpace(req.DisplayName)

	if email == "" || req.Password == "" || displayName == "" {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Email, password and displayName are required",
		})
	}

	out, err := h.workspaceService.Register(c.Request().Context(), workspaceApp.RegisterInput{
		Email:       email,
		Password:    req.Password,
		DisplayName: displayName,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, generated.Error{
				Code:    "CONFLICT",
				Message: "User with this email already exists",
			})
		}
		if errors.Is(err, domainWorkspace.ErrInvalidPassword) ||
			errors.Is(err, domainWorkspace.ErrInvalidEmail) ||
			errors.Is(err, domainWorkspace.ErrInvalidDisplayName) {
			return c.JSON(http.StatusBadRequest, generated.Error{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to register user",
		})
	}

	SetAuthCookie(c, out.Token, h.isProduction)

	return c.JSON(http.StatusCreated, generated.AuthResponse{
		Token:     out.Token,
		User:      mapUserToGenerated(out.User),
		Workspace: mapWorkspaceToGenerated(out.Workspace),
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req generated.LoginJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	email := strings.TrimSpace(string(req.Email))
	if email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Email and password are required",
		})
	}

	out, err := h.workspaceService.Login(c.Request().Context(), workspaceApp.LoginInput{
		Email:    email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInvalidCredentials) || errors.Is(err, domainWorkspace.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, generated.Error{
				Code:    "UNAUTHORIZED",
				Message: "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to log in",
		})
	}

	SetAuthCookie(c, out.Token, h.isProduction)

	return c.JSON(http.StatusOK, generated.AuthResponse{
		Token:     out.Token,
		User:      mapUserToGenerated(out.User),
		Workspace: mapWorkspaceToGenerated(out.Workspace),
	})
}

func (h *Handler) Logout(c echo.Context) error {
	ClearAuthCookie(c)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *Handler) GetMe(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	out, err := h.workspaceService.GetMe(c.Request().Context(), userID, workspaceID)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserNotFound) || errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusUnauthorized, generated.Error{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to fetch user profile",
		})
	}

	return c.JSON(http.StatusOK, generated.MeResponse{
		User:      mapUserToGenerated(out.User),
		Workspace: mapWorkspaceToGenerated(out.Workspace),
	})
}
