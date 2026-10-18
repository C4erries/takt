package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/infrastructure/http/generated"
)

func TestCreateAutomation(t *testing.T) {
	wsID := uuid.New()
	colID := uuid.New()

	t.Run("successful rule creation with low-code DSL", func(t *testing.T) {
		e := echo.New()
		reqBody := generated.CreateAutomationRequest{
			Name:        "Move on Enter",
			TriggerType: generated.CreateAutomationRequestTriggerTypeTaskEnteredColumn,
			ColumnId:    &colID,
			RuleDefinition: &generated.RuleDefinition{
				Trigger: generated.Trigger{
					Type: generated.TriggerTypeTaskEnteredColumn,
					Params: &map[string]interface{}{
						"column_id": colID.String(),
					},
				},
				Conditions: &[]generated.Condition{
					{
						Field:    generated.Priority,
						Operator: generated.Equals,
						Value:    "urgent",
					},
				},
				Actions: []generated.Action{
					{
						Type: generated.SetPriority,
						Params: &map[string]interface{}{
							"priority": "high",
						},
					},
				},
			},
		}

		payload, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/workspaces/"+wsID.String()+"/automations", bytes.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		ruleDef := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{{Type: domainAutomation.ActionSetPriority}},
		}
		expectedRule, err := domainAutomation.NewRuleAutomation(wsID, nil, &colID, "Move on Enter", ruleDef, true)
		require.NoError(t, err)

		mockSvc := &mockAutomationService{
			createFn: func(ctx context.Context, input appAutomation.CreateAutomationInput) (*domainAutomation.Automation, error) {
				assert.Equal(t, "Move on Enter", input.Name)
				assert.Equal(t, domainAutomation.TriggerTaskEnteredColumn, input.TriggerType)
				require.NotNil(t, input.RuleDefinition)
				return expectedRule, nil
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err = h.CreateAutomation(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp generated.Automation
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Move on Enter", resp.Name)
	})

	t.Run("fails on invalid rule definition (> 5 actions)", func(t *testing.T) {
		e := echo.New()
		reqBody := generated.CreateAutomationRequest{
			Name:        "Too Many Actions",
			TriggerType: generated.CreateAutomationRequestTriggerTypeTaskEnteredColumn,
			RuleDefinition: &generated.RuleDefinition{
				Trigger: generated.Trigger{Type: generated.TriggerTypeTaskEnteredColumn},
				Actions: []generated.Action{
					{Type: generated.ArchiveTask},
					{Type: generated.ArchiveTask},
					{Type: generated.ArchiveTask},
					{Type: generated.ArchiveTask},
					{Type: generated.ArchiveTask},
					{Type: generated.ArchiveTask}, // 6 actions
				},
			},
		}

		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/workspaces/"+wsID.String()+"/automations", bytes.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, &mockAutomationService{})
		err := h.CreateAutomation(c, wsID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUpdateAutomation(t *testing.T) {
	wsID := uuid.New()
	autoID := uuid.New()

	t.Run("successful update", func(t *testing.T) {
		e := echo.New()
		newName := "Updated Rule Name"
		reqBody := generated.UpdateAutomationRequest{
			Name: &newName,
		}

		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), bytes.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		now := time.Now().UTC()
		updatedRule := &domainAutomation.Automation{
			ID:          autoID,
			WorkspaceID: wsID,
			Name:        newName,
			TriggerType: domainAutomation.TriggerTaskCompleted,
			IsEnabled:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockSvc := &mockAutomationService{
			updateFn: func(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error) {
				assert.Equal(t, autoID, input.AutomationID)
				assert.Equal(t, newName, *input.Name)
				return updatedRule, nil
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.UpdateAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("binds rule to column", func(t *testing.T) {
		e := echo.New()
		targetCol := uuid.New()
		body := `{"columnId":"` + targetCol.String() + `"}`
		req := httptest.NewRequest(http.MethodPut, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), bytes.NewReader([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		now := time.Now().UTC()
		updatedRule := &domainAutomation.Automation{
			ID:          autoID,
			WorkspaceID: wsID,
			ColumnID:    &targetCol,
			Name:        "Bound Rule",
			TriggerType: domainAutomation.TriggerTaskEnteredColumn,
			IsEnabled:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockSvc := &mockAutomationService{
			updateFn: func(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error) {
				assert.Equal(t, &targetCol, input.ColumnID)
				assert.False(t, input.ClearColumnID)
				return updatedRule, nil
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.UpdateAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unbinds rule from column with null columnId", func(t *testing.T) {
		e := echo.New()
		body := `{"columnId":null}`
		req := httptest.NewRequest(http.MethodPut, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), bytes.NewReader([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		now := time.Now().UTC()
		updatedRule := &domainAutomation.Automation{
			ID:          autoID,
			WorkspaceID: wsID,
			ColumnID:    nil,
			Name:        "Unbound Rule",
			TriggerType: domainAutomation.TriggerTaskEnteredColumn,
			IsEnabled:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockSvc := &mockAutomationService{
			updateFn: func(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error) {
				assert.Nil(t, input.ColumnID)
				assert.True(t, input.ClearColumnID)
				return updatedRule, nil
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.UpdateAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("updates rule definition with logical operator OR", func(t *testing.T) {
		e := echo.New()
		body := `{"ruleDefinition":{"trigger":{"type":"task_entered_column"},"logicalOperator":"or","conditions":[{"field":"priority","operator":"equals","value":"urgent"}],"actions":[{"type":"archive_task"}]}}`
		req := httptest.NewRequest(http.MethodPut, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), bytes.NewReader([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		now := time.Now().UTC()
		ruleDef := &domainAutomation.RuleDefinition{
			Trigger:         domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			LogicalOperator: domainAutomation.LogicOr,
			Conditions:      []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "urgent"}},
			Actions:         []domainAutomation.Action{{Type: domainAutomation.ActionArchiveTask}},
		}
		updatedRule := &domainAutomation.Automation{
			ID:             autoID,
			WorkspaceID:    wsID,
			Name:           "OR Rule",
			TriggerType:    domainAutomation.TriggerTaskEnteredColumn,
			RuleDefinition: ruleDef,
			IsEnabled:      true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		mockSvc := &mockAutomationService{
			updateFn: func(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error) {
				require.NotNil(t, input.RuleDefinition)
				assert.Equal(t, domainAutomation.LogicOr, input.RuleDefinition.LogicalOperator)
				return updatedRule, nil
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.UpdateAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp generated.Automation
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.NotNil(t, resp.RuleDefinition)
		require.NotNil(t, resp.RuleDefinition.LogicalOperator)
		assert.Equal(t, generated.Or, *resp.RuleDefinition.LogicalOperator)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), bytes.NewReader([]byte("{}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		mockSvc := &mockAutomationService{
			updateFn: func(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error) {
				return nil, domainAutomation.ErrAutomationNotFound
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.UpdateAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDeleteAutomation(t *testing.T) {
	wsID := uuid.New()
	autoID := uuid.New()

	t.Run("successful deletion returns 204", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		mockSvc := &mockAutomationService{
			deleteFn: func(ctx context.Context, wid, aid uuid.UUID) error {
				assert.Equal(t, wsID, wid)
				assert.Equal(t, autoID, aid)
				return nil
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.DeleteAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/workspaces/"+wsID.String()+"/automations/"+autoID.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("workspace_id", wsID.String())

		mockSvc := &mockAutomationService{
			deleteFn: func(ctx context.Context, wid, aid uuid.UUID) error {
				return domainAutomation.ErrAutomationNotFound
			},
		}

		h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
		err := h.DeleteAutomation(c, wsID, autoID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
