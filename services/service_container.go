package services

import (
	"github.com/nadilas/todo/config"
	interfaces "github.com/nadilas/todo/if"
	active_directory "github.com/nadilas/todo/services/active-directory"
	no_op_mail "github.com/nadilas/todo/services/no-op-mail"
	"sync"
)

type kernel struct {
	configProvider config.Provider
}

var (
	k             *kernel
	containerOnce sync.Once
)

func (k *kernel) InjectActiveDirectoryService() interfaces.ActiveDirectoryService {
	return &active_directory.Service{}
}

func (k *kernel) InjectEmailService() interfaces.EmailService {
	return &no_op_mail.Service{}
}

func ServiceContainer(configProvider config.Provider) *kernel {
	if k == nil {
		containerOnce.Do(func() {
			k = &kernel{configProvider: configProvider}
		})
	}
	return k
}
