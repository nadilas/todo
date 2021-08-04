# Run tests:
```shell
go test ./...
```

## Failing unit test:
https://github.com/nadilas/todo/blob/37ea4e9004ac9c746f0e7db71e482d2b79efd242/workflows/todo/workflow_test.go#L372

```shell
--- FAIL: TestTodoTestSuite (0.02s)
--- FAIL: TestTodoTestSuite/Test_Reminders (0.01s)
--- FAIL: TestTodoTestSuite/Test_Reminders/reminder_every_day_update_to_less_frequent (0.00s)

2021/08/04 11:31:49 DEBUG RequestCancelTimer TimerID 31
2021/08/04 11:31:49 INFO  Setup new reminder context taskId t1
2021/08/04 11:31:49 INFO  sending reminder
2021/08/04 11:31:49 DEBUG RequestCancelActivity ActivityID 33
2021/08/04 11:31:49 DEBUG RequestCancelActivity ActivityID 34
2021/08/04 11:31:49 ERROR Activity panic. WorkflowID default-test-workflow-id RunID default-test-run-id ActivityType FetchUser Attempt 1 PanicError
assert: mock: The method has been called over 2 times.
Either do one more Mock.On("FetchUser").Return(...), or remove extra call.
This call was unexpected:
FetchUser(*context.timerCtx,string)
```
