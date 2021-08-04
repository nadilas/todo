package testutils

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"runtime/debug"
)

type BDTemporalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	Env *testsuite.TestWorkflowEnvironment
}

func (s *BDTemporalTestSuite) AfterTest(suiteName, testName string) {
	// after Test methods on Suite -- could include multiple Scenarios
}

func (s *BDTemporalTestSuite) SetupTest() {
	// setup per Test method on Suite -- could include multiple Scenarios
}

func (s *BDTemporalTestSuite) Scenario(name string, mockSetupFn func(mockCtrl *gomock.Controller), steps ...func(suite **BDTemporalTestSuite)) {
	s.Run(name, func() {
		mockCtrl := gomock.NewController(s.T())
		s.Env = s.NewTestWorkflowEnvironment()

		if mockSetupFn != nil {
			mockSetupFn(mockCtrl)
		}

		for _, step := range steps {
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.T().Error("step paniced", r, string(debug.Stack()))
					}
				}()
				step(&s)
			}()
		}
		s.Env.AssertExpectations(s.T())
		mockCtrl.Finish()
	})
}

func (s *BDTemporalTestSuite) Given(fn func(suite **BDTemporalTestSuite)) func(suite **BDTemporalTestSuite) {
	return fn
}

func (s *BDTemporalTestSuite) When(fn func(suite **BDTemporalTestSuite)) func(suite **BDTemporalTestSuite) {
	return fn
}

func (s *BDTemporalTestSuite) Then(fn func(suite **BDTemporalTestSuite)) func(suite **BDTemporalTestSuite) {
	return fn
}

func (s *BDTemporalTestSuite) And(fn func(suite **BDTemporalTestSuite)) func(suite **BDTemporalTestSuite) {
	return fn
}
