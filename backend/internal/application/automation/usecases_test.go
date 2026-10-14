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

