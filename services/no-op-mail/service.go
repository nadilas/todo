package no_op_mail

import (
	"context"
	"github.com/nadilas/todo/todopb"
)

type Service struct {
}

func (s *Service) SendTaskReminder(ctx context.Context, reminderModel *todopb.TaskReminderModel, sender string, addressee ...string) error {
	panic("implement me")
}
