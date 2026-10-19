package automation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	"takt/internal/application/automation/mocks"
	domainAutomation "takt/internal/domain/automation"
)

var errDB = errors.New("db error")

func TestService_ListAutomations(t *testing.T) {
	wsID := uuid.New()
	rule1, _ := domainAutomation.NewAutomation(wsID, nil, "Rule 1", domainAutomation.TriggerTaskCompleted, nil, nil, true)
	rule2, _ := domainAutomation.NewAutomation(wsID, nil, "Rule 2", domainAutomation.TriggerTaskUpdated, nil, nil, false)

	testCases := []struct {
		name      string
		setupMock func(m *mocks.MockAutomationRepository)
		checkOut  func(t *testing.T, rules []*domainAutomation.Automation)
		wantErr   error
	}{
		{
			name: "successful list",
			setupMock: func(m *mocks.MockAutomationRepository) {
				m.EXPECT().ListWorkspaceAutomations(mock.Anything, wsID).Return([]*domainAutomation.Automation{rule1, rule2}, nil).Once()
			},
			checkOut: func(t *testing.T, rules []*domainAutomation.Automation) {
				require.Len(t, rules, 2)
				assert.Equal(t, "Rule 1", rules[0].Name)
				assert.Equal(t, "Rule 2", rules[1].Name)
			},
		},
		{
			name: "db error",
			setupMock: func(m *mocks.MockAutomationRepository) {
				m.EXPECT().ListWorkspaceAutomations(mock.Anything, wsID).Return(nil, errDB).Once()
			},
			wantErr: errDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockAutomationRepository(t)
			tc.setupMock(repo)

			svc := appAutomation.NewService(repo)
			rules, err := svc.ListAutomations(context.Background(), wsID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, rules)
			} else {
				require.NoError(t, err)
				if tc.checkOut != nil {
					tc.checkOut(t, rules)
				}
			}
		})
	}
}

func TestService_ToggleAutomation(t *testing.T) {
	wsID := uuid.New()
	ruleID := uuid.New()
	rule, _ := domainAutomation.NewAutomation(wsID, nil, "Rule", domainAutomation.TriggerTaskCompleted, nil, nil, true)
	rule.ID = ruleID

	testCases := []struct {
		name      string
		isEnabled bool
		setupMock func(m *mocks.MockAutomationRepository)
		wantErr   error
	}{
		{
			name:      "successful toggle",
			isEnabled: false,
			setupMock: func(m *mocks.MockAutomationRepository) {
				updated := *rule
				updated.IsEnabled = false
				m.EXPECT().ToggleAutomation(mock.Anything, ruleID, wsID, false).Return(&updated, nil).Once()
			},
		},
		{
			name:      "rule not found",
			isEnabled: true,
			setupMock: func(m *mocks.MockAutomationRepository) {
				m.EXPECT().ToggleAutomation(mock.Anything, ruleID, wsID, true).Return(nil, domainAutomation.ErrAutomationNotFound).Once()
			},
			wantErr: domainAutomation.ErrAutomationNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockAutomationRepository(t)
			tc.setupMock(repo)

			svc := appAutomation.NewService(repo)
			res, err := svc.ToggleAutomation(context.Background(), wsID, ruleID, tc.isEnabled)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tc.isEnabled, res.IsEnabled)
			}
		})
	}
}

func TestService_SeedDefaultAutomations(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name      string
		setupMock func(m *mocks.MockAutomationRepository)
		wantErr   error
	}{
		{
			name: "successful seed of 3 rules",
			setupMock: func(m *mocks.MockAutomationRepository) {
				m.EXPECT().CreateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
					return a.Name == "Done & Archive" && a.IsEnabled == true
				})).Return(nil).Once()
				m.EXPECT().CreateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
					return a.Name == "Auto-Focus" && a.IsEnabled == true
				})).Return(nil).Once()
				m.EXPECT().CreateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
					return a.Name == "Inbox Zero" && a.IsEnabled == false
				})).Return(nil).Once()
			},
		},
		{
			name: "db error on creation",
			setupMock: func(m *mocks.MockAutomationRepository) {
				m.EXPECT().CreateAutomation(mock.Anything, mock.Anything).Return(errDB).Once()
			},
			wantErr: errDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockAutomationRepository(t)
			tc.setupMock(repo)

			svc := appAutomation.NewService(repo)
			err := svc.SeedDefaultAutomations(context.Background(), wsID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_UpdateAutomation(t *testing.T) {
	wsID := uuid.New()
	ruleID := uuid.New()
	existingColID := uuid.New()
	newColID := uuid.New()

	rule, _ := domainAutomation.NewAutomation(wsID, nil, "Original Name", domainAutomation.TriggerTaskCompleted, nil, nil, true)
	rule.ID = ruleID
	rule.ColumnID = &existingColID

	t.Run("binds rule to new column", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		copied := *rule
		repo.EXPECT().GetAutomationByID(mock.Anything, ruleID, wsID).Return(&copied, nil).Once()
		repo.EXPECT().UpdateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
			return a.ColumnID != nil && *a.ColumnID == newColID
		})).Return(nil).Once()

		svc := appAutomation.NewService(repo)
		updated, err := svc.UpdateAutomation(context.Background(), appAutomation.UpdateAutomationInput{
			AutomationID: ruleID,
			WorkspaceID:  wsID,
			ColumnID:     &newColID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, &newColID, updated.ColumnID)
	})

	t.Run("unbinds rule from column when ClearColumnID is true", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		copied := *rule
		repo.EXPECT().GetAutomationByID(mock.Anything, ruleID, wsID).Return(&copied, nil).Once()
		repo.EXPECT().UpdateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
			return a.ColumnID == nil
		})).Return(nil).Once()

		svc := appAutomation.NewService(repo)
		updated, err := svc.UpdateAutomation(context.Background(), appAutomation.UpdateAutomationInput{
			AutomationID:  ruleID,
			WorkspaceID:   wsID,
			ClearColumnID: true,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Nil(t, updated.ColumnID)
	})

	t.Run("updates rule definition with logical operator", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		copied := *rule
		repo.EXPECT().GetAutomationByID(mock.Anything, ruleID, wsID).Return(&copied, nil).Once()
		repo.EXPECT().UpdateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
			return a.RuleDefinition != nil && a.RuleDefinition.LogicalOperator == domainAutomation.LogicOr
		})).Return(nil).Once()

		ruleDef := &domainAutomation.RuleDefinition{
			Trigger:         domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			LogicalOperator: domainAutomation.LogicOr,
			Conditions:      []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "urgent"}},
			Actions:         []domainAutomation.Action{{Type: domainAutomation.ActionArchiveTask}},
		}

		svc := appAutomation.NewService(repo)
		updated, err := svc.UpdateAutomation(context.Background(), appAutomation.UpdateAutomationInput{
			AutomationID:   ruleID,
			WorkspaceID:    wsID,
			RuleDefinition: ruleDef,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, domainAutomation.LogicOr, updated.RuleDefinition.LogicalOperator)
	})

	t.Run("binds and unbinds rule for task", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		copied := *rule
		taskID := uuid.New()
		repo.EXPECT().GetAutomationByID(mock.Anything, ruleID, wsID).Return(&copied, nil).Twice()
		repo.EXPECT().UpdateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
			return a.TaskID != nil && *a.TaskID == taskID
		})).Return(nil).Once()
		repo.EXPECT().UpdateAutomation(mock.Anything, mock.MatchedBy(func(a *domainAutomation.Automation) bool {
			return a.TaskID == nil
		})).Return(nil).Once()

		svc := appAutomation.NewService(repo)
		updated, err := svc.UpdateAutomation(context.Background(), appAutomation.UpdateAutomationInput{
			AutomationID: ruleID,
			WorkspaceID:  wsID,
			TaskID:       &taskID,
		})
		require.NoError(t, err)
		assert.Equal(t, &taskID, updated.TaskID)

		cleared, err := svc.UpdateAutomation(context.Background(), appAutomation.UpdateAutomationInput{
			AutomationID: ruleID,
			WorkspaceID:  wsID,
			ClearTaskID:  true,
		})
		require.NoError(t, err)
		assert.Nil(t, cleared.TaskID)
	})
}

func TestService_ListTaskAutomations(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	rule, _ := domainAutomation.NewAutomation(wsID, nil, "Task Rule", domainAutomation.TriggerTaskTimer, nil, nil, true)
	rule.TaskID = &taskID

	t.Run("successful list", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		repo.EXPECT().ListAutomationsByTask(mock.Anything, taskID, wsID).Return([]*domainAutomation.Automation{rule}, nil).Once()

		svc := appAutomation.NewService(repo)
		rules, err := svc.ListTaskAutomations(context.Background(), wsID, taskID)
		require.NoError(t, err)
		require.Len(t, rules, 1)
		assert.Equal(t, &taskID, rules[0].TaskID)
	})
}

func TestService_ExecuteTaskAutomation(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	autoID := uuid.New()
	targetColID := uuid.New()

	ruleDef := &domainAutomation.RuleDefinition{
		Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskTimer},
		Actions: []domainAutomation.Action{
			{Type: domainAutomation.ActionMoveToColumn, Params: map[string]interface{}{"target_column_id": targetColID.String()}},
		},
	}
	rule, _ := domainAutomation.NewRuleAutomation(wsID, nil, nil, "Move Task", ruleDef, true)
	rule.ID = autoID
	rule.TaskID = &taskID

	task := &appAutomation.TaskInfo{ID: taskID, WorkspaceID: wsID, Priority: "high"}

	t.Run("successfully executes task automation", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		reader := mocks.NewMockTaskReader(t)
		mod := mocks.NewMockTaskModifier(t)

		repo.EXPECT().GetAutomationByID(mock.Anything, autoID, wsID).Return(rule, nil).Once()
		reader.EXPECT().GetTaskByID(mock.Anything, taskID, wsID).Return(task, nil).Once()
		mod.EXPECT().MoveTaskToColumn(mock.Anything, taskID, wsID, &targetColID, mock.Anything, 1000.0).Return(nil).Once()

		evaluator := appAutomation.NewEvaluator(mod, nil)
		svc := appAutomation.NewService(repo, reader, evaluator)

		err := svc.ExecuteTaskAutomation(context.Background(), wsID, taskID, autoID)
		require.NoError(t, err)
	})

	t.Run("returns ErrAutomationNotFound if task does not exist", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		reader := mocks.NewMockTaskReader(t)

		repo.EXPECT().GetAutomationByID(mock.Anything, autoID, wsID).Return(rule, nil).Once()
		reader.EXPECT().GetTaskByID(mock.Anything, taskID, wsID).Return(nil, nil).Once()

		evaluator := appAutomation.NewEvaluator(nil, nil)
		svc := appAutomation.NewService(repo, reader, evaluator)

		err := svc.ExecuteTaskAutomation(context.Background(), wsID, taskID, autoID)
		require.ErrorIs(t, err, domainAutomation.ErrAutomationNotFound)
	})
}

