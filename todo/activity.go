package todo

import (
	"context"
	"github.com/nadilas/todo/config"
	interfaces "github.com/nadilas/todo/if"
	"github.com/nadilas/todo/todopb"
	"time"
)

type WorkflowData struct {
	AssignedSince time.Duration
}

type Activities struct {
	configProvider config.Provider
	emailService   interfaces.EmailService
	adService      interfaces.ActiveDirectoryService
}

func NewActivities(
	configProvider config.Provider,
	emailService interfaces.EmailService,
	directoryService interfaces.ActiveDirectoryService,
) *Activities {
	return &Activities{
		configProvider: configProvider,
		emailService:   emailService,
		adService:      directoryService,
	}
}

func (a *Activities) FetchUser(
	ctx context.Context,
	username string,
) (*todopb.ADUser, error) {
	return &todopb.ADUser{
		DisplayName:    username,
		Domain:         "domain",
		EmailAddress:   username + "@domain.com",
		SamAccountName: username,
	}, nil
}

func (a *Activities) CollectWorkflowData(
	ctx context.Context,
	stepId string,
) (*WorkflowData, error) {
	return &WorkflowData{
		AssignedSince: time.Minute * 60,
	}, nil
}

func (a *Activities) PrepareReminderModel(
	ctx context.Context,
	data *WorkflowData,
	assignedSince time.Duration,
	remark string,
) (*todopb.TaskReminderModel, error) {
	return &todopb.TaskReminderModel{}, nil
}

func (a *Activities) SendTaskReminder(
	ctx context.Context,
	model *todopb.TaskReminderModel,
	addressee []string,
) error {
	return a.emailService.SendTaskReminder(ctx, model, "tool@domain.com", addressee...)
}
