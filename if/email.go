package interfaces

import (
	"context"
	"github.com/nadilas/todo/todopb"
)

type EmailService interface {
	SendTaskReminder(ctx context.Context, reminderModel *todopb.TaskReminderModel, sender string, addressee ...string) error
}
