package container

// Services container holding all domain and application use cases.
type Services struct {
	// Domain application services will be wired here in upcoming sprints
}

// initServices initializes application use case services.
func initServices(repos *Repositories) *Services {
	return &Services{}
}
