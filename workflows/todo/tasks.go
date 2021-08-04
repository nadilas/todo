package todo

import (
	"fmt"
	"github.com/nadilas/todo/todopb"
	"go.temporal.io/sdk/workflow"
)

type Tasks struct {
	Items     []*todopb.TodoItem
	reminders []reminder
}

type reminder struct {
	taskId   string
	cancelFn workflow.CancelFunc
}

type signalSetup struct {
	channel workflow.ReceiveChannel
	fn      func(c workflow.ReceiveChannel, more bool)
}

func (t *Tasks) setupSignals(ctx workflow.Context, sel workflow.Selector) {
	sel.AddReceive(workflow.GetSignalChannel(ctx, AddTaskSignal), t.handleAddTaskSignal(ctx, sel))
	sel.AddReceive(workflow.GetSignalChannel(ctx, UpdateTaskSignal), t.handleUpdateTaskSignal(ctx, sel))
	sel.AddReceive(workflow.GetSignalChannel(ctx, DeleteTaskSignal), t.handleDeleteTaskSignal(ctx, sel))
}

func (t *Tasks) queryPendingTasks() ([]*todopb.TodoItem, error) {
	i := make([]*todopb.TodoItem, 0, len(t.Items))
	for _, item := range t.Items {
		if item.CompletedAt == nil {
			i = append(i, item)
		}
	}
	return i, nil
}

func (t *Tasks) queryAllTasks() ([]*todopb.TodoItem, error) {
	return t.Items, nil
}

func (t *Tasks) pendingTasksCount() int {
	pending, _ := t.queryPendingTasks()
	return len(pending)
}

func (t *Tasks) indexOfTask(uuid string) (int, error) {
	for i, task := range t.Items {
		if task.Uuid == uuid {
			return i, nil
		}
	}
	return -1, fmt.Errorf("task not found: %s", uuid)
}

func (t *Tasks) indexOfReminder(uuid string) (int, error) {
	for i, reminder := range t.reminders {
		if reminder.taskId == uuid {
			return i, nil
		}
	}
	return -1, fmt.Errorf("reminder not found: %s", uuid)
}

func (t *Tasks) refreshReminders(ctx workflow.Context, sel workflow.Selector) {
	for _, item := range t.Items {
		// if we need a reminder start timer
		if item.Reminder != nil {
			idx, _ := t.indexOfReminder(item.Uuid)
			if idx >= 0 {
				// cancel previous reminder before setting up new one
				workflow.GetLogger(ctx).Info("Cancelling reminder context", "taskId", item.Uuid)
				rem := t.reminders[idx]
				rem.cancelFn()
				t.reminders = append(t.reminders[0:idx], t.reminders[idx+1:]...)
			}

			// setup new reminder
			timerCtx, cancel := workflow.WithCancel(ctx)
			t.reminders = append(t.reminders, reminder{
				taskId:   item.Uuid,
				cancelFn: cancel,
			})
			workflow.GetLogger(ctx).Info("Setup new reminder context", "taskId", item.Uuid)
			initReminder(timerCtx, sel, item)
		}
	}
}
