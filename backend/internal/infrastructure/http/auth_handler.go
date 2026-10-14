package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	workspaceApp "takt/internal/application/workspace"
	domainWorkspace "takt/internal/domain/workspace"
)

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type WorkspaceResponse struct {
	ID        uuid.UUID                      `json:"id"`
	Name      string                         `json:"name"`
	Type      domainWorkspace.WorkspaceType  `json:"type"`
	Role      domainWorkspace.MemberRole     `json:"role"`
	CreatedAt time.Time                      `json:"createdAt"`
}

type AuthResponse struct {
	Token     string            `json:"token"`
	User      UserResponse      `json:"user"`
	Workspace WorkspaceResponse `json:"workspace"`
}

type MeResponse struct {
	User      UserResponse      `json:"user"`
	Workspace WorkspaceResponse `json:"workspace"`
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Invalid request payload",
		})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Email, password and displayName are required",
		})
	}

	out, err := h.workspaceService.Register(c.Request().Context(), workspaceApp.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"code":    "CONFLICT",
				"message": "User with this email already exists",
			})
		}
		if errors.Is(err, domainWorkspace.ErrInvalidPassword) ||
			errors.Is(err, domainWorkspace.ErrInvalidEmail) ||
			errors.Is(err, domainWorkspace.ErrInvalidDisplayName) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"code":    "BAD_REQUEST",
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to register user",
		})
	}

	SetAuthCookie(c, out.Token, h.isProduction)

	return c.JSON(http.StatusCreated, AuthResponse{
		Token: out.Token,
		User: UserResponse{
			ID:          out.User.ID,
			Email:       out.User.Email,
			DisplayName: out.User.DisplayName,
			CreatedAt:   out.User.CreatedAt,
		},
		Workspace: WorkspaceResponse{
			ID:        out.Workspace.ID,
			Name:      out.Workspace.Name,
			Type:      out.Workspace.Type,
			Role:      out.Workspace.Role,
			CreatedAt: out.Workspace.CreatedAt,
		},
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Invalid request payload",
		})
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Email and password are required",
		})
	}

	out, err := h.workspaceService.Login(c.Request().Context(), workspaceApp.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInvalidCredentials) || errors.Is(err, domainWorkspace.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to log in",
		})
	}

	SetAuthCookie(c, out.Token, h.isProduction)

	return c.JSON(http.StatusOK, AuthResponse{
		Token: out.Token,
		User: UserResponse{
			ID:          out.User.ID,
			Email:       out.User.Email,
			DisplayName: out.User.DisplayName,
			CreatedAt:   out.User.CreatedAt,
		},
		Workspace: WorkspaceResponse{
			ID:        out.Workspace.ID,
			Name:      out.Workspace.Name,
			Type:      out.Workspace.Type,
			Role:      out.Workspace.Role,
			CreatedAt: out.Workspace.CreatedAt,
		},
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
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
	}

	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
	}

	out, err := h.workspaceService.GetMe(c.Request().Context(), userID, workspaceID)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserNotFound) || errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch user profile",
		})
	}

	return c.JSON(http.StatusOK, MeResponse{
		User: UserResponse{
			ID:          out.User.ID,
			Email:       out.User.Email,
			DisplayName: out.User.DisplayName,
			CreatedAt:   out.User.CreatedAt,
		},
		Workspace: WorkspaceResponse{
			ID:        out.Workspace.ID,
			Name:      out.Workspace.Name,
			Type:      out.Workspace.Type,
			Role:      out.Workspace.Role,
			CreatedAt: out.Workspace.CreatedAt,
		},
	})
}
