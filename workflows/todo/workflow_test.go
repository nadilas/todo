package todo_test

import (
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	todo2 "github.com/nadilas/todo/todo"
	"github.com/nadilas/todo/todopb"
	. "github.com/nadilas/todo/workflows/testutils"
	"github.com/nadilas/todo/workflows/todo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

var dummyUser = &todopb.ADUser{
	DisplayName:    "Last First (dept)",
	Domain:         "DOMAIN",
	EmailAddress:   "last.first@domain.com",
	SamAccountName: "user",
}

type TasklistTestSuite struct {
	BDTemporalTestSuite
}

func TestTodoTestSuite(t *testing.T) {
	suite.Run(t, &TasklistTestSuite{})
}

func (s *TasklistTestSuite) setupMocks(mockCtrl *gomock.Controller) {
	// no-op
}

func (s *TasklistTestSuite) Test_Basics() {
	var tasks *todo.Tasks
	dummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
	}
	dummyCompletedTask := &todopb.TodoItem{
		Uuid:        "t2",
		Description: "some task 2",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
		CompletedAt: timestamppb.Now(),
		CompletedBy: "user1",
	}
	s.Scenario("perpetual workflow renews",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask)),
		s.And(queryTasksIn(time.Minute*5, todo.PendingTasksQuery, dummyTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("perpetual workflows stops with no tasks",
		s.setupMocks,
		s.Given(aTasklist(&tasks)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
	s.Scenario("perpetual workflows stops with no pending tasks",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyCompletedTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
}

func (s *TasklistTestSuite) Test_Tasks_Queries() {
	var tasks *todo.Tasks
	// var reminder *todopb.Reminder
	dummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
	}
	dummyCompletedTask := &todopb.TodoItem{
		Uuid:        "t2",
		Description: "some task 2",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
		CompletedAt: timestamppb.Now(),
		CompletedBy: "user1",
	}
	dummyCompletedTask2 := &todopb.TodoItem{
		Uuid:        "t3",
		Description: "some task 3",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
		CompletedAt: timestamppb.Now(),
		CompletedBy: "user1",
	}

	s.Scenario("1/2 completed tasks gives 1 pending in query",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(queryTasksIn(time.Minute*5, todo.PendingTasksQuery, dummyTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("2/2 => 0 pending and 2 tasks",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyCompletedTask, dummyCompletedTask2)),
		s.And(queryTasksIn(time.Minute*5, todo.PendingTasksQuery)),
		s.And(queryTasksIn(time.Minute*5, todo.AllTasksQuery, dummyCompletedTask, dummyCompletedTask2)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
}

func (s *TasklistTestSuite) Test_AddSignal() {
	var tasks *todo.Tasks
	// var reminder *todopb.Reminder
	dummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
	}
	s.Scenario("adding one todo is persisted",
		s.setupMocks,
		s.Given(aTasklist(&tasks)),
		s.And(ProxySignalSucceeded(time.Minute*1, todo.AddTaskSignal, MustMarshalAny(&todopb.AddTodoRequest{
			Item: dummyTask,
		}), nil)),
		s.And(queryTasksIn(time.Minute*2, todo.PendingTasksQuery, dummyTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("adding a non-todo fails",
		s.setupMocks,
		s.Given(aTasklist(&tasks)),
		s.And(ProxySignalInvalidPayload(time.Minute*1, todo.AddTaskSignal, todopb.AddTodoRequest{})),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
	s.Scenario("adding todo with proxy issue is ignored",
		s.setupMocks,
		s.Given(aTasklist(&tasks)),
		s.And(ProxySignalIgnored(time.Minute*1, todo.AddTaskSignal, MustMarshalAny(&todopb.AddTodoRequest{
			Item: dummyTask,
		}), "")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
	s.Scenario("adding a todo with missing payload fails",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask)),
		s.And(ProxySignalErrored(time.Minute*1, todo.AddTaskSignal, nil, "todo is missing")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
}

func (s *TasklistTestSuite) Test_DeleteTaskSignal() {
	var tasks *todo.Tasks
	// var reminder *todopb.Reminder
	dummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
	}
	dummyCompletedTask := &todopb.TodoItem{
		Uuid:        "t2",
		Description: "some task 2",
		CreatedBy:   "user1",
		CreatedAt:   timestamppb.Now(),
		CompletedAt: timestamppb.Now(),
		CompletedBy: "user1",
	}
	s.Scenario("deleting 1 task reduces all tasks",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalSucceeded(time.Minute*3, todo.DeleteTaskSignal, MustMarshalAny(&todopb.DeleteTodoRequest{
			Uuid: "t2",
		}), nil)),
		s.And(queryTasksIn(time.Minute*5, todo.AllTasksQuery, dummyTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("deleting all tasks finishes workflow",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalSucceeded(time.Minute*3, todo.DeleteTaskSignal, MustMarshalAny(&todopb.DeleteTodoRequest{
			Uuid: "t2",
		}), nil)),
		s.And(queryTasksIn(time.Minute*5, todo.AllTasksQuery, dummyTask)),
		s.And(ProxySignalSucceeded(time.Minute*7, todo.DeleteTaskSignal, MustMarshalAny(&todopb.DeleteTodoRequest{
			Uuid: "t1",
		}), nil)),
		s.And(queryTasksIn(time.Minute*10, todo.AllTasksQuery)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
	s.Scenario("deleting a non-todo fails",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalInvalidPayload(time.Minute*1, todo.DeleteTaskSignal, todopb.DeleteTodoRequest{})),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("adding todo with proxy issue is ignored",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalIgnored(time.Minute*1, todo.DeleteTaskSignal, MustMarshalAny(&todopb.DeleteTodoRequest{
			Uuid: "t2",
		}), "")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("deleting a todo with missing payload fails",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalErrored(time.Minute*1, todo.DeleteTaskSignal, nil, "todo is not defined")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("trying to delete a todo not in the workflow",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask)),
		s.And(ProxySignalErrored(time.Minute*1, todo.DeleteTaskSignal, MustMarshalAny(&todopb.DeleteTodoRequest{
			Uuid: "t2",
		}), "task not found: t2")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
}

func (s *TasklistTestSuite) Test_UpdateTaskSignal() {
	var tasks *todo.Tasks
	// var reminder *todopb.Reminder
	dummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   timestamppb.Now(),
	}
	modifiedDummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task changed slightly",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   dummyTask.CreatedAt,
		CompletedAt: timestamppb.Now(),
		CompletedBy: dummyUser.SamAccountName,
	}
	dummyCompletedTask := &todopb.TodoItem{
		Uuid:        "t2",
		Description: "some task 2",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   timestamppb.Now(),
		CompletedAt: timestamppb.Now(),
		CompletedBy: dummyUser.SamAccountName,
	}
	modifiedCompletedDummyTask := &todopb.TodoItem{
		Uuid:        "t2",
		Description: "some task 2 changed",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   dummyCompletedTask.CreatedAt,
	}
	s.Scenario("updating a non-todo fails",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalInvalidPayload(time.Minute*1, todo.UpdateTaskSignal, todopb.UpdateTodoRequest{})),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("updating a todo with proxy issue is ignored",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalIgnored(time.Minute*1, todo.UpdateTaskSignal, MustMarshalAny(&todopb.UpdateTodoRequest{
			Item: modifiedDummyTask,
		}), "")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("updating a todo with missing payload fails",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalErrored(time.Minute*1, todo.UpdateTaskSignal, nil, "todo is not defined")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("trying to update a todo not in the workflow",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask)),
		s.And(ProxySignalErrored(time.Minute*1, todo.UpdateTaskSignal, MustMarshalAny(&todopb.UpdateTodoRequest{
			Item: modifiedCompletedDummyTask,
		}), "task not found: t2")),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("undo complete of a task",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalSucceeded(time.Minute*1, todo.UpdateTaskSignal, MustMarshalAny(&todopb.UpdateTodoRequest{
			Item: modifiedCompletedDummyTask,
		}), nil)),
		s.And(queryTasksIn(time.Minute*2, todo.PendingTasksQuery, dummyTask, modifiedCompletedDummyTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("completing the last open task, finishes workflow",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask, dummyCompletedTask)),
		s.And(ProxySignalSucceeded(time.Minute*1, todo.UpdateTaskSignal, MustMarshalAny(&todopb.UpdateTodoRequest{
			Item: modifiedDummyTask,
		}), nil)),
		s.And(queryTasksIn(time.Minute*2, todo.PendingTasksQuery)),
		s.And(queryTasksIn(time.Minute*3, todo.AllTasksQuery, modifiedDummyTask, dummyCompletedTask)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowFinished()),
	)
}

func (s *TasklistTestSuite) Test_Reminders() {
	var tasks *todo.Tasks
	oneDayHours := 24
	dummyTask := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   timestamppb.Now(),
		Reminder: &todopb.Reminder{
			Every: (time.Hour * time.Duration(oneDayHours)).String(),
		},
	}
	dummyTaskWith1Reminder := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   timestamppb.Now(),
		Reminder: &todopb.Reminder{
			At: timestamppb.New(time.Now().Add(time.Hour * 48)),
		},
	}
	afterDays := 4
	dummyTaskWithDailyReminderAfter := &todopb.TodoItem{
		Uuid:        "t1",
		Description: "some task",
		CreatedBy:   dummyUser.SamAccountName,
		CreatedAt:   timestamppb.Now(),
		Reminder: &todopb.Reminder{
			At:    timestamppb.New(time.Now().Add(time.Hour * time.Duration(oneDayHours) * time.Duration(afterDays))),
			Every: (time.Hour * time.Duration(oneDayHours)).String(),
		},
	}
	everyDayReminderCount := todo.MaxHoursPerRun/oneDayHours - 1
	s.Scenario("reminder every day",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask)),
		s.And(queryTasksIn(time.Minute*1, todo.PendingTasksQuery, dummyTask)),
		s.And(func(suite **BDTemporalTestSuite) {
			st := *suite
			act := &todo2.Activities{}
			st.Env.OnActivity(act.FetchUser, mock.Anything, dummyUser.SamAccountName).Return(dummyUser, nil).Times(everyDayReminderCount)
			wfData := &todo2.WorkflowData{
				AssignedSince: time.Minute * 60,
			}
			st.Env.OnActivity(act.CollectWorkflowData, mock.Anything, "default-test-workflow-id").Return(wfData, nil).Times(everyDayReminderCount)
			st.Env.OnActivity(act.SendTaskReminder, mock.Anything, mock.Anything, []string{dummyUser.EmailAddress}).Return(nil).Times(everyDayReminderCount)
		}),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("reminder every day update to less frequent",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTask)),
		s.And(func(suite **BDTemporalTestSuite) {
			st := *suite
			act := &todo2.Activities{}
			st.Env.OnActivity(act.FetchUser, mock.Anything, dummyUser.SamAccountName).Return(dummyUser, nil).Times(2)
			wfData := &todo2.WorkflowData{
				AssignedSince: time.Minute * 60,
			}
			st.Env.OnActivity(act.CollectWorkflowData, mock.Anything, "default-test-workflow-id").Return(wfData, nil).Times(2)
			st.Env.OnActivity(act.SendTaskReminder, mock.Anything, mock.Anything, []string{dummyUser.EmailAddress}).Return(nil).Times(2)
		}),
		s.And(ProxySignalSucceeded(time.Hour*49, todo.UpdateTaskSignal, MustMarshalAny(&todopb.UpdateTodoRequest{
			Item: &todopb.TodoItem{
				Uuid:        "t1",
				Description: "some task",
				CreatedBy:   dummyUser.SamAccountName,
				CreatedAt:   timestamppb.Now(),
				Reminder: &todopb.Reminder{
					Every: (time.Hour * 168).String(),
				},
			},
		}), nil)),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("reminder in 2 days",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTaskWith1Reminder)),
		s.And(func(suite **BDTemporalTestSuite) {
			st := *suite
			act := &todo2.Activities{}
			st.Env.OnActivity(act.FetchUser, mock.Anything, dummyUser.SamAccountName).Return(dummyUser, nil).Times(1)
			wfData := &todo2.WorkflowData{
				AssignedSince: time.Minute * 60,
			}
			st.Env.OnActivity(act.CollectWorkflowData, mock.Anything, "default-test-workflow-id").Return(wfData, nil).Times(1)
			st.Env.OnActivity(act.SendTaskReminder, mock.Anything, mock.Anything, []string{dummyUser.EmailAddress}).Return(nil).Times(1)
		}),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
	s.Scenario("reminder every day after 4 days",
		s.setupMocks,
		s.Given(aTasklist(&tasks, dummyTaskWithDailyReminderAfter)),
		s.And(func(suite **BDTemporalTestSuite) {
			st := *suite
			act := &todo2.Activities{}
			callCount := todo.MaxHoursPerRun/oneDayHours - afterDays
			st.Env.OnActivity(act.FetchUser, mock.Anything, dummyUser.SamAccountName).Return(dummyUser, nil).Times(callCount)
			wfData := &todo2.WorkflowData{
				AssignedSince: time.Minute * 60,
			}
			st.Env.OnActivity(act.CollectWorkflowData, mock.Anything, "default-test-workflow-id").Return(wfData, nil).Times(callCount)
			st.Env.OnActivity(act.SendTaskReminder, mock.Anything, mock.Anything, []string{dummyUser.EmailAddress}).Return(nil).Times(callCount)
		}),
		s.When(startAWorkflow(&tasks)),
		s.Then(WorkflowContinuedAsNew()),
	)
}

func startAWorkflow(t **todo.Tasks) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		st := *suite
		st.Env.ExecuteWorkflow(todo.Tasklist, *t)
	}
}

func aTasklist(t **todo.Tasks, items ...*todopb.TodoItem) func(suite **BDTemporalTestSuite) {
	return func(suite **BDTemporalTestSuite) {
		*t = &todo.Tasks{
			Items: items,
		}
	}
}

func queryTasksIn(delay time.Duration, queryType string, expectedTasks ...*todopb.TodoItem) func(s **BDTemporalTestSuite) {
	return func(s **BDTemporalTestSuite) {
		st := *s
		st.Env.RegisterDelayedCallback(func() {
			workflow, err := st.Env.QueryWorkflow(queryType)
			st.NoError(err, "executing query: '"+queryType+"' shouldn't error out")
			var resultContainer []*todopb.TodoItem
			err = workflow.Get(&resultContainer)
			st.NoError(err, "query result '"+queryType+"' should be available")

			if expectedTasks != nil {
				diff := cmp.Diff(expectedTasks, resultContainer, cmp.Comparer(func(t1, t2 *todopb.TodoItem) bool {
					return t1.Uuid == t2.Uuid && t1.Description == t2.Description
				}))
				st.True(diff == "", "query result '"+queryType+"' doesn't match expectations: "+diff)
			} else {
				st.True(len(resultContainer) == 0, "query result '"+queryType+"' doesn't match expectations: "+pretty.Sprint(resultContainer))
			}
		}, delay)
	}
}
