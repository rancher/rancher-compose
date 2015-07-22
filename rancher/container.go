package rancher

type Container struct {
	id, name string
}

func NewContainer(id, name string) *Container {
	return &Container{
		id:   id,
		name: name,
	}
}

func (c *Container) Id() (string, error) {
	return c.id, nil
}

func (c *Container) Name() string {
	return c.name
}

func (c *Container) Port(port string) (string, error) {
	//TODO: implement
	return "", nil
}
