package workflows

import (
	"github.com/nadilas/todo/config"
	interfaces "github.com/nadilas/todo/if"
	todo2 "github.com/nadilas/todo/todo"
	"github.com/nadilas/todo/workflows/signalproxy"
	"github.com/nadilas/todo/workflows/todo"
	"go.temporal.io/sdk/worker"
)

func Register(worker worker.Worker, configProvider config.Provider, serviceContainer interfaces.Kernel, services ...string) {
	// region Tasklist
	worker.RegisterWorkflow(todo.Tasklist)
	todoActivities := todo2.NewActivities(
		configProvider,
		serviceContainer.InjectEmailService(),
		serviceContainer.InjectActiveDirectoryService(),
	)
	worker.RegisterActivity(todoActivities)
	// endregion

	// region SignalProxy
	worker.RegisterWorkflow(signalproxy.SignalProxy)
	// endregion
}
