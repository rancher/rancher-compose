package app

import (
	"log"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
)

type ProjectAction func(project *project.Project, c *cli.Context)

func BeforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func WithProject(factory ProjectFactory, action ProjectAction) func(context *cli.Context) {
	return func(context *cli.Context) {
		p, err := factory.Create(context)
		if err != nil {
			log.Fatalf("Failed to read project: %v", err)
		}
		action(p, context)
	}
}

func ProjectDown(p *project.Project, c *cli.Context) {
	err := p.Down(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectCreate(p *project.Project, c *cli.Context) {
	err := p.Create(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectUp(p *project.Project, c *cli.Context) {
	err := p.Up(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}

	if !c.Bool("d") {
		wait()
	}
}

func ProjectRestart(p *project.Project, c *cli.Context) {
	err := p.Restart(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectLog(p *project.Project, c *cli.Context) {
	err := p.Log(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
	wait()
}

func ProjectPull(p *project.Project, c *cli.Context) {
	err := p.Pull(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectDelete(p *project.Project, c *cli.Context) {
	if !c.Bool("force") && len(c.Args()) == 0 {
		logrus.Fatal("Will not remove all services with out --force")
	}
	err := p.Delete(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectScale(p *project.Project, c *cli.Context) {
	// This code is a bit verbose but I wanted to parse everything up front
	order := make([]string, 0, 0)
	serviceScale := make(map[string]int)
	services := make(map[string]project.Service)

	for _, arg := range c.Args() {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			logrus.Fatalf("Invalid scale parameter: %s", arg)
		}

		name := kv[0]

		count, err := strconv.Atoi(kv[1])
		if err != nil {
			logrus.Fatalf("Invalid scale parameter: %v", err)
		}

		if _, ok := p.Configs[name]; !ok {
			logrus.Fatalf("% is not defined in the template", name)
		}

		service, err := p.CreateService(name)
		if err != nil {
			logrus.Fatalf("Failed to lookup service: %s: %v", service, err)
		}

		order = append(order, name)
		serviceScale[name] = count
		services[name] = service
	}

	for _, name := range order {
		scale := serviceScale[name]
		logrus.Infof("Setting scale %s=%d...", name, scale)
		err := services[name].Scale(scale)
		if err != nil {
			logrus.Fatalf("Failed to set the scale %s=%d: %v", name, scale, err)
		}
	}
}

func wait() {
	<-make(chan interface{})
}
