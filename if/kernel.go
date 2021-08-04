package interfaces

type Kernel interface {
	InjectActiveDirectoryService() ActiveDirectoryService
	InjectEmailService() EmailService
}
