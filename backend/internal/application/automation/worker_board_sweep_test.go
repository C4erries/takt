package automation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	appMocks "takt/internal/application/automation/mocks"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

func TestWorker_ScheduledSweep_BoardLevel(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	boardID := uuid.New()
	col1 := uuid.New()
	col2 := uuid.New()

	testCases := []struct {
		name      string
		ruleDef   *domainAutomation.RuleDefinition
		setupMocks func(
			repo *appMocks.MockAutomationRepository,
			proj *appMocks.MockProjectReader,
			batch *appMocks.MockBatchTaskExecutor,
			pub *appMocks.MockEventPublisher,
			reader *mockTaskReader,
			modifier *mockTaskModifier,
		)
	}{
		{
			name: "board sweep with fast path batch execution across all columns of board",
			ruleDef: &domainAutomation.RuleDefinition{
				Trigger: domainAutomation.Trigger{
					Type: domainAutomation.TriggerScheduledSweep,
					Params: map[string]interface{}{},
				},
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
			},
			setupMocks: func(
				repo *appMocks.MockAutomationRepository,
				proj *appMocks.MockProjectReader,
				batch *appMocks.MockBatchTaskExecutor,
				pub *appMocks.MockEventPublisher,
				reader *mockTaskReader,
				modifier *mockTaskModifier,
			) {
				proj.EXPECT().ListColumnIDsByBoard(mock.Anything, boardID).Return([]uuid.UUID{col1, col2}, nil).Once()

				batch.EXPECT().ArchiveCompletedTasksBatchSQL(mock.Anything, wsID, col1, 0.0).Return(int64(2), nil).Once()
				pub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
					ev, ok := e.(events.BatchSweepCompletedEvent)
					return ok && ev.WorkspaceID() == wsID && ev.ColumnID() == col1 && ev.AffectedCount == 2
				})).Return(nil).Once()

				batch.EXPECT().ArchiveCompletedTasksBatchSQL(mock.Anything, wsID, col2, 0.0).Return(int64(4), nil).Once()
				pub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
					ev, ok := e.(events.BatchSweepCompletedEvent)
					return ok && ev.WorkspaceID() == wsID && ev.ColumnID() == col2 && ev.AffectedCount == 4
				})).Return(nil).Once()
			},
		},
		{
			name: "board sweep with standard evaluator when action is not archive",
			ruleDef: &domainAutomation.RuleDefinition{
				Trigger: domainAutomation.Trigger{
					Type: domainAutomation.TriggerScheduledSweep,
					Params: map[string]interface{}{},
				},
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "low"}},
				},
			},
			setupMocks: func(
				repo *appMocks.MockAutomationRepository,
				proj *appMocks.MockProjectReader,
				batch *appMocks.MockBatchTaskExecutor,
				pub *appMocks.MockEventPublisher,
				reader *mockTaskReader,
				modifier *mockTaskModifier,
			) {
				task1 := &appAutomation.TaskInfo{ID: uuid.New(), WorkspaceID: wsID, ColumnID: &col1, Title: "Task 1"}
				task2 := &appAutomation.TaskInfo{ID: uuid.New(), WorkspaceID: wsID, ColumnID: &col2, Title: "Task 2"}

				proj.EXPECT().ListColumnIDsByBoard(mock.Anything, boardID).Return([]uuid.UUID{col1, col2}, nil).Once()

				reader.On("ListTasksByColumn", mock.Anything, wsID, col1).Return([]*appAutomation.TaskInfo{task1}, nil).Once()
				modifier.On("SetTaskPriority", mock.Anything, task1.ID, wsID, "low").Return(nil).Once()

				reader.On("ListTasksByColumn", mock.Anything, wsID, col2).Return([]*appAutomation.TaskInfo{task2}, nil).Once()
				modifier.On("SetTaskPriority", mock.Anything, task2.ID, wsID, "low").Return(nil).Once()
			},
		},
		{
			name: "board sweep handles project reader error gracefully",
			ruleDef: &domainAutomation.RuleDefinition{
				Trigger: domainAutomation.Trigger{
					Type: domainAutomation.TriggerScheduledSweep,
					Params: map[string]interface{}{},
				},
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
			},
			setupMocks: func(
				repo *appMocks.MockAutomationRepository,
				proj *appMocks.MockProjectReader,
				batch *appMocks.MockBatchTaskExecutor,
				pub *appMocks.MockEventPublisher,
				reader *mockTaskReader,
				modifier *mockTaskModifier,
			) {
				proj.EXPECT().ListColumnIDsByBoard(mock.Anything, boardID).Return(nil, errors.New("db error")).Once()
			},
		},
		{
			name: "explicit column_id param overrides board scope",
			ruleDef: &domainAutomation.RuleDefinition{
				Trigger: domainAutomation.Trigger{
					Type: domainAutomation.TriggerScheduledSweep,
					Params: map[string]interface{}{
						"column_id": col1.String(),
					},
				},
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
			},
			setupMocks: func(
				repo *appMocks.MockAutomationRepository,
				proj *appMocks.MockProjectReader,
				batch *appMocks.MockBatchTaskExecutor,
				pub *appMocks.MockEventPublisher,
				reader *mockTaskReader,
				modifier *mockTaskModifier,
			) {
				// Only col1 should be processed via fast path, ListColumnIDsByBoard must NOT be called
				batch.EXPECT().ArchiveCompletedTasksBatchSQL(mock.Anything, wsID, col1, 0.0).Return(int64(1), nil).Once()
				pub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
					ev, ok := e.(events.BatchSweepCompletedEvent)
					return ok && ev.WorkspaceID() == wsID && ev.ColumnID() == col1 && ev.AffectedCount == 1
				})).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(appMocks.MockAutomationRepository)
			mockBatch := new(appMocks.MockBatchTaskExecutor)
			mockPub := new(appMocks.MockEventPublisher)
			mockReader := new(mockTaskReader)
			mockMod := new(mockTaskModifier)
			mockProj := new(appMocks.MockProjectReader)

			eval := appAutomation.NewEvaluator(mockMod, mockProj)
			worker := appAutomation.NewWorker(mockRepo, mockReader, eval, 1*time.Minute, mockProj)
			worker.SetBatchExecutor(mockBatch)
			worker.SetEventPublisher(mockPub)

			autoRule, err := domainAutomation.NewRuleAutomation(wsID, &boardID, nil, "Board Sweep", tc.ruleDef, true)
			require.NoError(t, err)

			mockRepo.EXPECT().ListAllEnabledTimerAutomations(mock.Anything).Return([]*domainAutomation.Automation{autoRule}, nil).Once()
			tc.setupMocks(mockRepo, mockProj, mockBatch, mockPub, mockReader, mockMod)

			err = worker.Tick(context.Background())
			require.NoError(t, err)

			mockRepo.AssertExpectations(t)
			mockProj.AssertExpectations(t)
			mockBatch.AssertExpectations(t)
			mockPub.AssertExpectations(t)
			mockReader.AssertExpectations(t)
			mockMod.AssertExpectations(t)
		})
	}
}
