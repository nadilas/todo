package todo

import (
	"github.com/nadilas/todo/todopb"
	"github.com/nadilas/todo/workflows/signalproxy"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/anypb"
)

func (t *Tasks) handleUpdateTaskSignal(ctx workflow.Context, sel workflow.Selector) func(c workflow.ReceiveChannel, more bool) {
	return func(c workflow.ReceiveChannel, more bool) {
		var r *signalproxy.InputData

		c.Receive(ctx, &r)
		workflow.GetLogger(ctx).Debug("Received update todo signal", "completionId", r.CompletionTargetId)

		if r.CompletionTargetId == "" {
			workflow.GetLogger(ctx).Warn("Silently ignoring update signal with no completionId")
			return
		}

		if r.Data == nil {
			reportSignalError(ctx, r.CompletionTargetId, "todo is not defined")
			return
		}

		request := &todopb.UpdateTodoRequest{}
		if err := r.Data.UnmarshalTo(request); err != nil {
			reportSignalError(ctx, r.CompletionTargetId, err.Error())
			return
		}

		if request.Item == nil {
			reportSignalError(ctx, r.CompletionTargetId, "todo is not defined")
			return
		}

		idx, err := t.indexOfTask(request.Item.Uuid)
		if err != nil {
			reportSignalError(ctx, r.CompletionTargetId, err.Error())
			return
		}

		// update elements task entirely
		task := t.Items[idx]

		task.Description = request.Item.Description
		task.CompletedAt = request.Item.CompletedAt
		task.CompletedBy = request.Item.CompletedBy
		task.Reminder = request.Item.Reminder

		workflow.GetLogger(ctx).Debug("Updated task", "atIndex", idx, "taskId", request.Item.Uuid)
		resp, _ := anypb.New(&todopb.UpdateTodoResponse{
			Item: task,
		})
		reportSignalSuccess(ctx, r.CompletionTargetId, resp)
		t.refreshReminders(ctx, sel)
	}
}
