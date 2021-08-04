package todo

import (
	"github.com/nadilas/todo/todopb"
	"github.com/nadilas/todo/workflows/signalproxy"
	"go.temporal.io/sdk/workflow"
)

func (t *Tasks) handleDeleteTaskSignal(ctx workflow.Context, sel workflow.Selector) func(c workflow.ReceiveChannel, more bool) {
	return func(c workflow.ReceiveChannel, more bool) {
		var r *signalproxy.InputData

		c.Receive(ctx, &r)
		workflow.GetLogger(ctx).Debug("Received delete todo signal", "completionId", r.CompletionTargetId)

		if r.CompletionTargetId == "" {
			workflow.GetLogger(ctx).Warn("Silently ignoring delete signal with no completionId")
			return
		}

		if r.Data == nil {
			reportSignalError(ctx, r.CompletionTargetId, "todo is not defined")
			return
		}

		request := &todopb.DeleteTodoRequest{}
		if err := r.Data.UnmarshalTo(request); err != nil {
			reportSignalError(ctx, r.CompletionTargetId, err.Error())
			return
		}

		idx, err := t.indexOfTask(request.Uuid)
		if err != nil {
			reportSignalError(ctx, r.CompletionTargetId, err.Error())
			return
		}
		// remove task entirely
		t.Items = append(t.Items[0:idx], t.Items[idx+1:]...)
		workflow.GetLogger(ctx).Debug("Deleted task", "atIndex", idx, "taskId", request.Uuid)
		reportSignalSuccess(ctx, r.CompletionTargetId, nil)

		// cleanup reminder
		ridx, err := t.indexOfReminder(request.Uuid)
		if err != nil {
			return
		}
		// cancel previous reminder before setting up new one
		workflow.GetLogger(ctx).Info("Cancelling reminder context", "taskId", request.Uuid)
		rem := t.reminders[idx]
		rem.cancelFn()
		t.reminders = append(t.reminders[0:ridx], t.reminders[ridx+1:]...)
	}
}
