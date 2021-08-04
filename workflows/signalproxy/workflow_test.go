package signalproxy_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/nadilas/todo/workflows/signalproxy"
	. "github.com/nadilas/todo/workflows/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"log"
	"testing"
	"time"
)

type TestSignalProxySuite struct {
	BDTemporalTestSuite
}

func TestSignalProxyTestSuite(t *testing.T) {
	suite.Run(t, &TestSignalProxySuite{})
}

func (s *TestSignalProxySuite) setupMocks(mockCtrl *gomock.Controller) {

}

func (s *TestSignalProxySuite) Test_Basics() {
	var payload *signalproxy.Payload
	var result *signalproxy.Result

	s.Scenario("proxy succeeds", s.setupMocks,
		s.Given(signalWithSuccessfulProxy(
			&payload,
			&result,
			"targetWorkflow",
			"approve",
			durationpb.New(time.Minute*90),
			nil,
		)),
		s.When(sendSignalViaProxy(&payload)),
		s.Then(completeWithResult(&result)),
	)

	s.Scenario("proxy cannot signal target workflow", s.setupMocks,
		s.Given(signalWithFailedProxy(
			&payload,
			"targetWorkflow",
			"approve",
			durationpb.New(time.Minute*90),
			"can't find workflow to signal",
		)),
		s.When(sendSignalViaProxy(&payload)),
		s.Then(completesWithError("can't find workflow to signal")),
	)
}

func completeWithResult(result **signalproxy.Result) func(s **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite
		st.True(st.Env.IsWorkflowCompleted())
		st.NoError(st.Env.GetWorkflowError())

		var res *signalproxy.Result
		err := st.Env.GetWorkflowResult(&res)
		st.NoError(err, "failed to fetch workflow result")
		st.NotNil(res, "workflow result empty")
		st.EqualValues(*result, res, "workflow result doesn't match expectation")
	}
}

func completesWithError(expectedErrorMsg string) func(s **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite
		st.True(st.Env.IsWorkflowCompleted())
		err := st.Env.GetWorkflowError()
		st.Error(err, "workflow was supposed to end with error")
		var weErr *temporal.WorkflowExecutionError
		st.True(errors.As(err, &weErr), "error is not execution error")
		st.EqualError(weErr.Unwrap(), expectedErrorMsg, "expected error message was not received")
	}
}

func sendSignalViaProxy(payload **signalproxy.Payload) func(s **BDTemporalTestSuite) {
	return func(st **BDTemporalTestSuite) {
		p := *payload
		(*st).Env.ExecuteWorkflow(signalproxy.SignalProxy, p)
	}
}

func signalWithSuccessfulProxy(payload **signalproxy.Payload, result **signalproxy.Result, targetWorkflowId, targetSignalName string, inputData proto.Message, outputData proto.Message) func(st **BDTemporalTestSuite) {
	return func(st **BDTemporalTestSuite) {
		var any *anypb.Any
		var outAny *anypb.Any
		var err error

		if inputData != nil {
			any, err = anypb.New(inputData)
			if err != nil {
				panic(err)
			}

		}
		*payload = &signalproxy.Payload{
			TargetId:   targetWorkflowId,
			SignalName: targetSignalName,
			Data:       any,
		}

		if outputData != nil {
			outAny, err = anypb.New(outputData)
			if err != nil {
				panic(err)
			}
		}
		*result = &signalproxy.Result{
			Success: true,
			Data:    outAny,
		}

		// setup target workflow listener for completion of request
		(*st).Env.OnSignalExternalWorkflow(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			func(namespace, workflowID, runID, signalName string, arg interface{}) error {
				if workflowID == targetWorkflowId {
					data := arg.(signalproxy.InputData)
					log.Println("signaled with completion target workflowid", data.CompletionTargetId)
					(*st).Env.SignalWorkflowByID(data.CompletionTargetId, signalproxy.CompletedSignal, *result)
				}
				return nil // succeed
			})
	}
}

func signalWithFailedProxy(payload **signalproxy.Payload, targetWorkflowId, targetSignalName string, inputData proto.Message, expectedErrMsg string) func(st **BDTemporalTestSuite) {
	return func(st **BDTemporalTestSuite) {
		any, err := anypb.New(inputData)
		if err != nil {
			panic(err)
		}
		*payload = &signalproxy.Payload{
			TargetId:   targetWorkflowId,
			SignalName: targetSignalName,
			Data:       any,
		}

		// setup target workflow listener for completion of request
		(*st).Env.OnSignalExternalWorkflow(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			func(namespace, workflowID, runID, signalName string, arg interface{}) error {
				if workflowID == targetWorkflowId {
					return errors.New(expectedErrMsg)
				}
				return nil // succeed
			})
	}
}
