package testutils

import (
	"github.com/nadilas/todo/workflows/signalproxy"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"reflect"
	"strings"
	"time"
)

// ProxySignalIgnored region signal helpers
func ProxySignalIgnored(expectedIn time.Duration, signal string, data *anypb.Any, completionTargetId string) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite

		st.Env.RegisterDelayedCallback(func() {
			// gRPC logic:
			//   get taskGroupId from requestflow via TaskGroupIdForStepQuery
			//   exeucte SignalProxy workflow with taskGroupId

			// region fake SignalProxy workflow trigger:
			// 	trigger approve on defined task group
			st.Env.SignalWorkflow(signal, signalproxy.InputData{
				CompletionTargetId: completionTargetId,
				Data:               data,
			})
			// endregion
		}, expectedIn)
	}
}

func ProxySignalErrored(expectedIn time.Duration, signal string, data *anypb.Any, expectedErrMsg string) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite

		completionTargetId := "completionId"
		// fake SignalProxy workflow completion:
		// 	accept completion signal
		st.Env.OnSignalExternalWorkflow(mock.Anything, completionTargetId, mock.Anything, signalproxy.CompletedSignal, mock.Anything).Return(
			func(namespace, workflowID, runID, signalName string, arg interface{}) error {
				st.IsType(&signalproxy.Result{}, arg, "result is not appropriate")
				res := arg.(*signalproxy.Result)
				st.False(res.Success, signal+" did not fail as expected")
				st.True(strings.Contains(res.Error, expectedErrMsg), "returned error was not expected")
				return nil
			},
		).Once()

		st.Env.RegisterDelayedCallback(func() {
			// gRPC logic:
			//   get taskGroupId from requestflow via TaskGroupIdForStepQuery
			//   exeucte SignalProxy workflow with taskGroupId

			// region fake SignalProxy workflow trigger:
			// 	trigger approve on defined task group
			st.Env.SignalWorkflow(signal, signalproxy.InputData{
				CompletionTargetId: completionTargetId,
				Data:               data,
			})
			// endregion
		}, expectedIn)
	}
}

func ProxySignalSucceeded(expectedIn time.Duration, signal string, data *anypb.Any, expectedResult *anypb.Any) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite

		completionTargetId := "completionId"
		// fake SignalProxy workflow completion:
		// 	accept completion signal
		st.Env.OnSignalExternalWorkflow(mock.Anything, completionTargetId, mock.Anything, signalproxy.CompletedSignal, mock.Anything).Return(
			func(namespace, workflowID, runID, signalName string, arg interface{}) error {
				st.IsType(&signalproxy.Result{}, arg, "result is not appropriate")
				res := arg.(*signalproxy.Result)
				st.True(res.Success, signal+" did not succeed as expected")
				if expectedResult != nil {
					st.Equal(expectedResult, res.Data, "returned data does not match expected")
				}
				return nil
			},
		).Once()

		st.Env.RegisterDelayedCallback(func() {
			// gRPC logic:
			//   get taskGroupId from requestflow via TaskGroupIdForStepQuery
			//   exeucte SignalProxy workflow with taskGroupId

			// region fake SignalProxy workflow trigger:
			// 	trigger approve on defined task group
			st.Env.SignalWorkflow(signal, signalproxy.InputData{
				CompletionTargetId: completionTargetId,
				Data:               data,
			})
			// endregion
		}, expectedIn)
	}
}

// ProxySignalInvalidPayload use non-pointers for expected payload type, as reflection Type.String() adds a pointer * at the beginning, which will cause an error
func ProxySignalInvalidPayload(expectedIn time.Duration, signal string, expectedPayloadType interface{}) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite

		// fake SignalProxy workflow completion:
		// 	accept completion signal
		completionTargetId := "cbwfid"
		typeOf := reflect.TypeOf(expectedPayloadType)
		st.Env.OnSignalExternalWorkflow(mock.Anything, completionTargetId, mock.Anything, signalproxy.CompletedSignal, mock.Anything).Return(
			func(namespace, workflowID, runID, signalName string, arg interface{}) error {
				st.IsType(&signalproxy.Result{}, arg, "result is not appropriate")
				res := arg.(*signalproxy.Result)
				st.False(res.Success, "approve did not fail as expected")
				st.True(strings.Contains(res.Error, "mismatched message type:"), "returned error was unexpected")
				st.True(strings.Contains(res.Error, "got \""+typeOf.String()+"\""), "mismatch got is not: "+typeOf.String()+" but: "+res.Error)
				return nil
			},
		).Once()

		st.Env.RegisterDelayedCallback(func() {
			// gRPC logic:
			//   get taskGroupId from requestflow via TaskGroupIdForStepQuery
			//   exeucte SignalProxy workflow with taskGroupId

			// region fake SignalProxy workflow trigger:
			// 	trigger approve on defined task group
			st.Env.SignalWorkflow(signal, signalproxy.InputData{
				CompletionTargetId: completionTargetId,
				Data:               MustMarshalAny(durationpb.New(time.Minute * 60)),
			})
			// endregion
		}, expectedIn)
	}
}

// endregion
