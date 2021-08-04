package testutils

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type BDTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func (s *BDTestSuite) AfterTest(suiteName, testName string) {
}

func (s *BDTestSuite) SetupTest() {
}

func (s *BDTestSuite) Scenario(name string, mockSetupFn func(mockCtrl *gomock.Controller), steps ...func(suite **BDTestSuite)) {
	s.Run(name, func() {
		mockCtrl := gomock.NewController(s.T())
		if mockSetupFn != nil {
			mockSetupFn(mockCtrl)
		}

		for _, step := range steps {
			s.NotPanics(func() {
				step(&s)
			})
		}
		mockCtrl.Finish()
	})
}

func (s *BDTestSuite) Given(fn func(suite **BDTestSuite)) func(suite **BDTestSuite) {
	return fn
}

func (s *BDTestSuite) When(fn func(suite **BDTestSuite)) func(suite **BDTestSuite) {
	return fn
}

func (s *BDTestSuite) Then(fn func(suite **BDTestSuite)) func(suite **BDTestSuite) {
	return fn
}

func (s *BDTestSuite) And(fn func(suite **BDTestSuite)) func(suite **BDTestSuite) {
	return fn
}
