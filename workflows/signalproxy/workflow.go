package signalproxy

import (
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	CompletedSignal = "completed"
)

// Payload is the input for a signal execution, which will be transformed and passed down to the target as InputData
type Payload struct {
	TargetId   string
	SignalName string
	Data       *anypb.Any
}

// InputData is the input data for the signal. Any signal handler using this proxied method has to unmarshal to this received type
type InputData struct {
	CompletionTargetId string
	Data               *anypb.Any
}

// Result is the output of a signal execution, which will be returned on workflow completion
type Result struct {
	Success bool
	Error   string
	Data    *anypb.Any
}

// SignalProxy is a workflow to coordinate a signal delivery and output check in a Request-Response fashion
func SignalProxy(ctx workflow.Context, payload *Payload) (*Result, error) {
	logger := workflow.GetLogger(ctx)
	completeChannel := workflow.GetSignalChannel(ctx, CompletedSignal)
	logger.Debug("Starting to proxy signal data", "targetId", payload.TargetId, "signalName", payload.SignalName)
	// wrap the target data
	wi := workflow.GetInfo(ctx)
	inputData := InputData{
		CompletionTargetId: wi.WorkflowExecution.ID,
		Data:               payload.Data,
	}
	// target the latest runID
	err := workflow.SignalExternalWorkflow(ctx, payload.TargetId, "", payload.SignalName, inputData).Get(ctx, nil)
	if err != nil {
		logger.Error("Proxy signal failed", "error", err)
		return nil, err
	}
	// wait for completion
	var result *Result
	completeChannel.Receive(ctx, &result)
	logger.Debug("Finished proxying signal data", "targetId", payload.TargetId, "signalName", payload.SignalName)
	return result, nil
}
