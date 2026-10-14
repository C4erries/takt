package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	domainAutomation "takt/internal/domain/automation"
	"takt/internal/infrastructure/http/generated"
)

func mapDomainAutomationToGenerated(a *domainAutomation.Automation) generated.Automation {
	var boardID *openapi_types.UUID
	if a.BoardID != nil {
		b := *a.BoardID
		boardID = &b
	}

	var conditions map[string]interface{}
	if len(a.ConditionsJSON) > 0 {
		_ = json.Unmarshal(a.ConditionsJSON, &conditions)
	}
	if conditions == nil {
		conditions = map[string]interface{}{}
	}

	var actions map[string]interface{}
	if len(a.ActionsJSON) > 0 {
		_ = json.Unmarshal(a.ActionsJSON, &actions)
	}
	if actions == nil {
		actions = map[string]interface{}{}
	}

	return generated.Automation{
		Id:             a.ID,
		WorkspaceId:    a.WorkspaceID,
		BoardId:        boardID,
		Name:           a.Name,
		TriggerType:    generated.AutomationTriggerType(a.TriggerType),
		ConditionsJson: conditions,
		ActionsJson:    actions,
		IsEnabled:      a.IsEnabled,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func (h *Handler) ListWorkspaceAutomations(c echo.Context, workspaceId openapi_types.UUID) error {
	callerWsID, err := GetWorkspaceIDFromContext(c)
	if err == nil && callerWsID != workspaceId {
		return c.JSON(http.StatusForbidden, generated.Error{
			Code:    "FORBIDDEN",
			Message: "Access to workspace denied",
		})
	}

	if h.automationService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Automation service unavailable",
		})
	}

	rules, err := h.automationService.ListAutomations(c.Request().Context(), workspaceId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list automations",
		})
	}

	resp := make([]generated.Automation, len(rules))
	for i, r := range rules {
		resp[i] = mapDomainAutomationToGenerated(r)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ToggleAutomation(c echo.Context, workspaceId openapi_types.UUID, automationId openapi_types.UUID) error {
	callerWsID, err := GetWorkspaceIDFromContext(c)
	if err == nil && callerWsID != workspaceId {
		return c.JSON(http.StatusForbidden, generated.Error{
			Code:    "FORBIDDEN",
			Message: "Access to workspace denied",
		})
	}

	if h.automationService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Automation service unavailable",
		})
	}

	var req generated.ToggleAutomationJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	updated, err := h.automationService.ToggleAutomation(c.Request().Context(), workspaceId, automationId, req.IsEnabled)
	if err != nil {
		if errors.Is(err, domainAutomation.ErrAutomationNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Automation not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to toggle automation",
		})
	}

	return c.JSON(http.StatusOK, mapDomainAutomationToGenerated(updated))
}

