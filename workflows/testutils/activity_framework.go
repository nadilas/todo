package testutils

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"runtime/debug"
)

type BDTemporalActivityTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	Env *testsuite.TestActivityEnvironment
}

func (s *BDTemporalActivityTestSuite) AfterTest(suiteName, testName string) {
	// after Test methods on Suite -- could include multiple Scenarios
}

func (s *BDTemporalActivityTestSuite) SetupTest() {
	// setup per Test method on Suite -- could include multiple Scenarios
}

func (s *BDTemporalActivityTestSuite) Scenario(name string, mockSetupFn func(mockCtrl *gomock.Controller), steps ...func(suite **BDTemporalActivityTestSuite)) {
	s.Run(name, func() {
		mockCtrl := gomock.NewController(s.T())
		s.Env = s.NewTestActivityEnvironment()

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
		mockCtrl.Finish()
	})
}

func (s *BDTemporalActivityTestSuite) Given(fn func(suite **BDTemporalActivityTestSuite)) func(suite **BDTemporalActivityTestSuite) {
	return fn
}

func (s *BDTemporalActivityTestSuite) When(fn func(suite **BDTemporalActivityTestSuite)) func(suite **BDTemporalActivityTestSuite) {
	return fn
}

func (s *BDTemporalActivityTestSuite) Then(fn func(suite **BDTemporalActivityTestSuite)) func(suite **BDTemporalActivityTestSuite) {
	return fn
}

func (s *BDTemporalActivityTestSuite) And(fn func(suite **BDTemporalActivityTestSuite)) func(suite **BDTemporalActivityTestSuite) {
	return fn
}
