package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	appAutomation "takt/internal/application/automation"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/infrastructure/http/generated"
)

func (h *Handler) checkAccessAndService(c echo.Context, wsID openapi_types.UUID) bool {
	callerWsID, err := GetWorkspaceIDFromContext(c)
	if err == nil && callerWsID != wsID {
		_ = c.JSON(http.StatusForbidden, generated.Error{Code: "FORBIDDEN", Message: "Access to workspace denied"})
		return false
	}
	if h.automationService == nil {
		_ = c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Automation service unavailable"})
		return false
	}
	return true
}

func mapDomainAutomationToGenerated(a *domainAutomation.Automation) generated.Automation {
	var boardID *openapi_types.UUID
	if a.BoardID != nil {
		b := *a.BoardID
		boardID = &b
	}
	var columnID *openapi_types.UUID
	if a.ColumnID != nil {
		col := *a.ColumnID
		columnID = &col
	}
	var taskID *openapi_types.UUID
	if a.TaskID != nil {
		t := *a.TaskID
		taskID = &t
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

	var ruleDef *generated.RuleDefinition
	if a.RuleDefinition != nil {
		ruleDef = mapDomainRuleDefinitionToGenerated(a.RuleDefinition)
	}

	return generated.Automation{
		Id:             a.ID,
		WorkspaceId:    a.WorkspaceID,
		BoardId:        boardID,
		ColumnId:       columnID,
		TaskId:         taskID,
		Name:           a.Name,
		TriggerType:    generated.AutomationTriggerType(a.TriggerType),
		ConditionsJson: conditions,
		ActionsJson:    actions,
		RuleDefinition: ruleDef,
		IsEnabled:      a.IsEnabled,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func mapDomainRuleDefinitionToGenerated(d *domainAutomation.RuleDefinition) *generated.RuleDefinition {
	if d == nil {
		return nil
	}
	var conditions []generated.Condition
	for _, c := range d.Conditions {
		conditions = append(conditions, generated.Condition{
			Field:    generated.ConditionField(c.Field),
			Operator: generated.ConditionOperator(c.Operator),
			Value:    c.Value,
		})
	}
	var actions []generated.Action
	for _, a := range d.Actions {
		p := a.Params
		actions = append(actions, generated.Action{
			Type:   generated.ActionType(a.Type),
			Params: &p,
		})
	}
	var logicalOp *generated.RuleDefinitionLogicalOperator
	if d.LogicalOperator != "" {
		op := generated.RuleDefinitionLogicalOperator(d.LogicalOperator)
		logicalOp = &op
	}
	triggerParams := d.Trigger.Params
	return &generated.RuleDefinition{
		Trigger: generated.Trigger{
			Type:   generated.TriggerType(d.Trigger.Type),
			Params: &triggerParams,
		},
		LogicalOperator: logicalOp,
		Conditions:      &conditions,
		Actions:         actions,
	}
}

func mapGeneratedRuleDefinitionToDomain(g *generated.RuleDefinition) *domainAutomation.RuleDefinition {
	if g == nil {
		return nil
	}
	var params map[string]interface{}
	if g.Trigger.Params != nil {
		params = *g.Trigger.Params
	}
	var logicalOp string
	if g.LogicalOperator != nil {
		logicalOp = string(*g.LogicalOperator)
	}
	var conditions []domainAutomation.Condition
	if g.Conditions != nil {
		for _, c := range *g.Conditions {
			conditions = append(conditions, domainAutomation.Condition{
				Field:    string(c.Field),
				Operator: string(c.Operator),
				Value:    c.Value,
			})
		}
	}
	var actions []domainAutomation.Action
	for _, a := range g.Actions {
		var actParams map[string]interface{}
		if a.Params != nil {
			actParams = *a.Params
		}
		actions = append(actions, domainAutomation.Action{
			Type:   string(a.Type),
			Params: actParams,
		})
	}
	return &domainAutomation.RuleDefinition{
		Trigger:         domainAutomation.Trigger{Type: string(g.Trigger.Type), Params: params},
		LogicalOperator: logicalOp,
		Conditions:      conditions,
		Actions:         actions,
	}
}

func (h *Handler) ListWorkspaceAutomations(c echo.Context, wsID openapi_types.UUID) error {
	if !h.checkAccessAndService(c, wsID) {
		return nil
	}
	ctx := c.Request().Context()
	var rules []*domainAutomation.Automation
	var err error

	if taskIDParam := c.QueryParam("taskId"); taskIDParam != "" {
		taskUUID, parseErr := uuid.Parse(taskIDParam)
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid taskId parameter"})
		}
		rules, err = h.automationService.ListTaskAutomations(ctx, wsID, taskUUID)
	} else {
		rules, err = h.automationService.ListAutomations(ctx, wsID)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to list automations"})
	}
	resp := make([]generated.Automation, len(rules))
	for i, r := range rules {
		resp[i] = mapDomainAutomationToGenerated(r)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateAutomation(c echo.Context, wsID openapi_types.UUID) error {
	if !h.checkAccessAndService(c, wsID) {
		return nil
	}
	var req generated.CreateAutomationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}

	domRuleDef := mapGeneratedRuleDefinitionToDomain(req.RuleDefinition)
	if domRuleDef != nil {
		if err := domainAutomation.ValidateRuleDefinition(domRuleDef); err != nil {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: err.Error()})
		}
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	condBytes, _ := json.Marshal(req.ConditionsJson)
	actBytes, _ := json.Marshal(req.ActionsJson)

	input := appAutomation.CreateAutomationInput{
		WorkspaceID:    wsID,
		BoardID:        req.BoardId,
		ColumnID:       req.ColumnId,
		TaskID:         req.TaskId,
		Name:           req.Name,
		TriggerType:    string(req.TriggerType),
		ConditionsJSON: condBytes,
		ActionsJSON:    actBytes,
		RuleDefinition: domRuleDef,
		IsEnabled:      isEnabled,
	}

	created, err := h.automationService.CreateAutomation(c.Request().Context(), input)
	if err != nil {
		if isClientError(err) {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to create automation"})
	}
	return c.JSON(http.StatusCreated, mapDomainAutomationToGenerated(created))
}

func (h *Handler) ToggleAutomation(c echo.Context, wsID, autoID openapi_types.UUID) error {
	if !h.checkAccessAndService(c, wsID) {
		return nil
	}
	var req generated.ToggleAutomationJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}
	updated, err := h.automationService.ToggleAutomation(c.Request().Context(), wsID, autoID, req.IsEnabled)
	if err != nil {
		if errors.Is(err, domainAutomation.ErrAutomationNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Automation not found"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to toggle automation"})
	}
	return c.JSON(http.StatusOK, mapDomainAutomationToGenerated(updated))
}

func (h *Handler) UpdateAutomation(c echo.Context, wsID, autoID openapi_types.UUID) error {
	if !h.checkAccessAndService(c, wsID) {
		return nil
	}
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}
	var req generated.UpdateAutomationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}
	var raw map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &raw)

	var clearCol bool
	if _, ok := raw["columnId"]; ok && req.ColumnId == nil {
		clearCol = true
	}
	var clearTask bool
	if _, ok := raw["taskId"]; ok && req.TaskId == nil {
		clearTask = true
	}

	domRuleDef := mapGeneratedRuleDefinitionToDomain(req.RuleDefinition)
	if domRuleDef != nil {
		if err := domainAutomation.ValidateRuleDefinition(domRuleDef); err != nil {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: err.Error()})
		}
	}

	var triggerType *string
	if req.TriggerType != nil {
		s := string(*req.TriggerType)
		triggerType = &s
	}

	var condBytes, actBytes []byte
	if req.ConditionsJson != nil {
		condBytes, _ = json.Marshal(req.ConditionsJson)
	}
	if req.ActionsJson != nil {
		actBytes, _ = json.Marshal(req.ActionsJson)
	}

	input := appAutomation.UpdateAutomationInput{
		AutomationID:   autoID,
		WorkspaceID:    wsID,
		BoardID:        req.BoardId,
		ColumnID:       req.ColumnId,
		ClearColumnID:  clearCol,
		TaskID:         req.TaskId,
		ClearTaskID:    clearTask,
		Name:           req.Name,
		TriggerType:    triggerType,
		ConditionsJSON: condBytes,
		ActionsJSON:    actBytes,
		RuleDefinition: domRuleDef,
		IsEnabled:      req.IsEnabled,
	}

	updated, err := h.automationService.UpdateAutomation(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domainAutomation.ErrAutomationNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Automation not found"})
		}
		if isClientError(err) {
			return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to update automation"})
	}
	return c.JSON(http.StatusOK, mapDomainAutomationToGenerated(updated))
}

func (h *Handler) DeleteAutomation(c echo.Context, wsID, autoID openapi_types.UUID) error {
	if !h.checkAccessAndService(c, wsID) {
		return nil
	}
	err := h.automationService.DeleteAutomation(c.Request().Context(), wsID, autoID)
	if err != nil {
		if errors.Is(err, domainAutomation.ErrAutomationNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{Code: "NOT_FOUND", Message: "Automation not found"})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{Code: "INTERNAL_ERROR", Message: "Failed to delete automation"})
	}
	return c.NoContent(http.StatusNoContent)
}

func isClientError(err error) bool {
	return errors.Is(err, domainAutomation.ErrInvalidAutomationName) ||
		errors.Is(err, domainAutomation.ErrInvalidTriggerType) ||
		errors.Is(err, domainAutomation.ErrRuleDefinitionNil) ||
		errors.Is(err, domainAutomation.ErrTooManyConditions) ||
		errors.Is(err, domainAutomation.ErrTooManyActions) ||
		errors.Is(err, domainAutomation.ErrEmptyActions) ||
		errors.Is(err, domainAutomation.ErrInvalidActionType) ||
		errors.Is(err, domainAutomation.ErrInvalidConditionField) ||
		errors.Is(err, domainAutomation.ErrInvalidConditionOperator)
}
