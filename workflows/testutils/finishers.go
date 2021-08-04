package testutils

import (
	"errors"
	"github.com/kr/pretty"
	"go.temporal.io/sdk/temporal"
	"log"
	"strings"
	"time"
)

// WorkflowFinished checks if the WorkflowExecution completed without an error
func WorkflowFinished() func(s **BDTemporalTestSuite) {
	return func(s **BDTemporalTestSuite) {
		st := *s
		st.True(st.Env.IsWorkflowCompleted())
		st.Nil(st.Env.GetWorkflowError())
	}
}

// WorkflowContinuedAsNew checks if the WorkflowExecution completed with a temporal.WorkflowExecutionError
func WorkflowContinuedAsNew() func(s **BDTemporalTestSuite) {
	return func(s **BDTemporalTestSuite) {
		st := *s
		st.True(st.Env.IsWorkflowCompleted())
		err := st.Env.GetWorkflowError()
		st.NotNil(err)
		var casn *temporal.WorkflowExecutionError
		st.True(errors.As(err, &casn))
		st.NotNil(casn, "workflow execution error was not present")
		if casn != nil {
			st.Equal("continue as new", casn.Unwrap().Error())
		}
	}
}

// WorkflowErrored checks for a generic error message as a substring within the wrapped error of a temporal.WorkflowExecutionError
func WorkflowErrored(expectedErrMsg string) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite
		st.True(st.Env.IsWorkflowCompleted())
		err := st.Env.GetWorkflowError()
		st.NotNil(err, "workflow didn't error out")
		if err != nil {
			var weErr *temporal.WorkflowExecutionError
			st.True(errors.As(err, &weErr), "err is not workflowExecutionError but "+GetErrorType(err))
			st.NotNil(weErr, "workflow execution error is not present")
			if weErr != nil {
				st.True(strings.Contains(weErr.Unwrap().Error(), expectedErrMsg), "workflow did not fail with expected error")
			}
		}
	}
}

func WorkflowTimedOut(start time.Time, expectedTimeout time.Duration) func(s **BDTemporalTestSuite) {
	return func(s **BDTemporalTestSuite) {
		st := *s
		st.True(st.Env.IsWorkflowCompleted(), "workflow didn't complete")
		err := st.Env.GetWorkflowError()
		st.NotNil(err, "workflow didn't return error")
		if err != nil {
			var tout *temporal.TimeoutError
			st.True(errors.As(err, &tout), "workflow didn't return timeout error:\n"+pretty.Sprint(err))
			st.Equal("ScheduleToClose", tout.TimeoutType().String(), "workflow didnt't return ScheduleToClose timeout type")
		}
		now := st.Env.Now()
		after := now.After(start.Add(expectedTimeout))
		log.Println("now", now, "start", start, "expectedTimeout", expectedTimeout, "after", after)
		st.True(after, "workflow timed out earlier than expected")
	}
}
