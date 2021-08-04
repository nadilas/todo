package todo

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

const maxEventsPerRun = 90000
const MaxHoursPerRun = 168
const MaxExecutionDurationPerRun = time.Hour * MaxHoursPerRun

const (
	AddTaskSignal     = "add_task"
	DeleteTaskSignal  = "delete_task"
	UpdateTaskSignal  = "update_task"
	PendingTasksQuery = "pending_tasks"
	AllTasksQuery     = "all_tasks"
)

func Tasklist(ctx workflow.Context, tasks *Tasks) (*Tasks, error) {
	logger := workflow.GetLogger(ctx)
	eventLoop := 0
	sel := workflow.NewSelector(ctx)
	codeRefreshTriggered := false
	tasks.refreshReminders(ctx, sel)

	// code refresh for all workflows
	sel.AddFuture(workflow.NewTimer(ctx, MaxExecutionDurationPerRun), func(f workflow.Future) {
		logger.Debug("Max execution per run exceeded, refreshing workflow code")
		codeRefreshTriggered = true
	})

	tasks.setupSignals(ctx, sel)

	if err := workflow.SetQueryHandler(ctx, PendingTasksQuery, tasks.queryPendingTasks); err != nil {
		return tasks, err
	}

	if err := workflow.SetQueryHandler(ctx, AllTasksQuery, tasks.queryAllTasks); err != nil {
		return tasks, err
	}

	for {
		sel.Select(ctx)
		eventLoop++
		if tasks.pendingTasksCount() < 1 || eventLoop >= maxEventsPerRun || codeRefreshTriggered {
			break
		}
	}

	notFinishedButLongHistory := eventLoop >= maxEventsPerRun
	if tasks.pendingTasksCount() > 0 && (notFinishedButLongHistory || codeRefreshTriggered) {
		logger.Debug("Clearing workflow history and carrying state to continue as new instance")
		return nil, workflow.NewContinueAsNewError(ctx, Tasklist, tasks)
	}

	logger.Debug("No more tasks remaining, closing workflow")
	return tasks, nil
}
