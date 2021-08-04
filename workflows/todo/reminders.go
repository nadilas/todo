package todo

import (
	"context"
	"errors"
	"github.com/nadilas/todo/todo"
	"github.com/nadilas/todo/todopb"
	"go.temporal.io/sdk/workflow"
	"time"
)

func initReminder(ctx workflow.Context, sel workflow.Selector, item *todopb.TodoItem) {
	if item.Reminder.Every != "" && item.Reminder.At == nil {
		remindEvery(ctx, sel, item)
	}
	if item.Reminder.At != nil && item.Reminder.Every == "" {
		sel.AddFuture(remindAt(ctx, item))
	}
	if item.Reminder.At != nil && item.Reminder.Every != "" {
		remindEveryAfter(ctx, sel, item)
	}
}

// remindEveryAfter sets a timer for
func remindEveryAfter(ctx workflow.Context, sel workflow.Selector, item *todopb.TodoItem) {
	// schedule first reminder
	f, fn := remindAt(ctx, item)

	sel.AddFuture(f, func(f workflow.Future) {
		// execute original function
		fn(f)

		// start loop
		remindEvery(ctx, sel, item)
	})
}

// remindAt sets a timer for a one-time reminder in the future.
func remindAt(ctx workflow.Context, item *todopb.TodoItem) (workflow.Future, func(future workflow.Future)) {
	now := workflow.Now(ctx)
	timeUntilReminder := item.Reminder.At.AsTime().Sub(now)
	workflow.GetLogger(ctx).Info("reminding at", "at", item.Reminder.At.AsTime(), "now", now, "in", timeUntilReminder)

	timer := workflow.NewTimer(ctx, timeUntilReminder)

	return timer, func(future workflow.Future) {
		sendReminder(ctx, item)
	}
}

// remindEvery start a continuously renewing reminder based on item.Reminder.Every time.Duration
func remindEvery(ctx workflow.Context, sel workflow.Selector, item *todopb.TodoItem) {
	duration, err := time.ParseDuration(item.Reminder.Every)
	if err != nil {
		return // ignore reminder if incorrect setup
	}

	t := workflow.NewTimer(ctx, duration)
	sel.AddFuture(t, reminderLoop(ctx, sel, item))
}

func reminderLoop(ctx workflow.Context, sel workflow.Selector, item *todopb.TodoItem) func(f workflow.Future) {
	return func(f workflow.Future) {
		duration, err := time.ParseDuration(item.Reminder.Every)
		if err != nil {
			return // ignore reminder if incorrect setup
		}

		workflow.GetLogger(ctx).Info("sending reminder")
		sendReminder(ctx, item)

		t := workflow.NewTimer(ctx, duration)
		sel.AddFuture(t, reminderLoop(ctx, sel, item))
	}
}

func sendReminder(ctx workflow.Context, item *todopb.TodoItem) {
	if errors.Is(ctx.Err(), context.Canceled) {
		return
	}
	act := todo.Activities{}
	aCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Hour * 2,
		StartToCloseTimeout:    time.Minute * 1,
	})
	// get user
	var user *todopb.ADUser
	workflow.ExecuteActivity(
		aCtx,
		act.FetchUser,
		item.CreatedBy,
	).Get(ctx, &user)
	// get workflow data
	var wfdata *todo.WorkflowData
	workflow.ExecuteActivity(
		aCtx,
		act.CollectWorkflowData,
		workflow.GetInfo(ctx).WorkflowExecution.ID,
	).Get(ctx, &wfdata)
	// get mail data model
	var mailmodel *todopb.TaskReminderModel
	workflow.ExecuteActivity(
		aCtx,
		act.PrepareReminderModel,
		wfdata,
		wfdata.AssignedSince,
		"Reminder from ToDo: "+item.Description,
	).Get(ctx, &mailmodel)
	// send mail
	workflow.ExecuteActivity(
		aCtx,
		act.SendTaskReminder,
		mailmodel,
		[]string{user.EmailAddress},
	).Get(ctx, &user)
}
