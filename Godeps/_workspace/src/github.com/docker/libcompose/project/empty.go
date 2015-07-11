package project

type EmptyService struct {
	ServiceName   string
	ServiceConfig *ServiceConfig
}

func (e *EmptyService) Name() string {
	return e.ServiceName
}

func (e *EmptyService) Config() *ServiceConfig {
	return e.ServiceConfig
}

func (e *EmptyService) Create() error {
	return nil
}

func (e *EmptyService) Up() error {
	return nil
}

func (e *EmptyService) Down() error {
	return nil
}

func (e *EmptyService) Delete() error {
	return nil
}

func (e *EmptyService) Restart() error {
	return nil
}

func (e *EmptyService) Log() error {
	return nil

}
func (e *EmptyService) Scale(count int) error {
	return nil
}

func (e *EmptyService) DependentServices() []string {
	config := e.Config()
	if config != nil {
		return append(config.Links.Slice(), config.VolumesFrom...)
	}
	return []string{}
}
