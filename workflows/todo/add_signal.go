package todo

import (
	"github.com/nadilas/todo/todopb"
	"github.com/nadilas/todo/workflows/signalproxy"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/anypb"
)

func (t *Tasks) handleAddTaskSignal(ctx workflow.Context, sel workflow.Selector) func(c workflow.ReceiveChannel, more bool) {
	return func(c workflow.ReceiveChannel, more bool) {
		var r *signalproxy.InputData

		c.Receive(ctx, &r)
		workflow.GetLogger(ctx).Debug("Received add todo signal", "completionId", r.CompletionTargetId)

		if r.CompletionTargetId == "" {
			workflow.GetLogger(ctx).Warn("Silently ignoring Add signal with no completionId")
			return
		}

		if r.Data == nil {
			reportSignalError(ctx, r.CompletionTargetId, "todo is missing")
			return
		}

		addRequest := &todopb.AddTodoRequest{}
		if err := r.Data.UnmarshalTo(addRequest); err != nil {
			reportSignalError(ctx, r.CompletionTargetId, err.Error())
			return
		}

		t.Items = append(t.Items, addRequest.Item)
		reportSignalSuccess(ctx, r.CompletionTargetId, nil)

		t.refreshReminders(ctx, sel)
	}
}

func reportSignalError(ctx workflow.Context, id string, errMsg string) {
	workflow.GetLogger(ctx).Debug("Sending back error for signal", "error", errMsg)
	workflow.SignalExternalWorkflow(ctx, id, "", signalproxy.CompletedSignal, &signalproxy.Result{
		Error: errMsg,
	}).Get(ctx, nil)
}

func reportSignalSuccess(ctx workflow.Context, id string, resp *anypb.Any) {
	workflow.GetLogger(ctx).Debug("Sending back success for signal")
	workflow.SignalExternalWorkflow(ctx, id, "", signalproxy.CompletedSignal, &signalproxy.Result{
		Success: true,
		Data:    resp,
	}).Get(ctx, nil)
}
