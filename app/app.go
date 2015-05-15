package app

import (
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancherio/rancher-compose/librcompose/project"
	"github.com/rancherio/rancher-compose/rancher"
)

type options struct {
	log bool
}

func newContext(c *cli.Context, opts options) *rancher.Context {
	return &rancher.Context{
		Log:                opts.log,
		RancherComposeFile: c.GlobalString("rancher-file"),
		ComposeFile:        c.GlobalString("file"),
		ProjectName:        c.GlobalString("project-name"),
		Url:                c.GlobalString("url"),
		AccessKey:          c.GlobalString("access-key"),
		SecretKey:          c.GlobalString("secret-key"),
	}
}

func ProjectDown(c *cli.Context) {
	err := requireProject(c, options{}).Down(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectCreate(c *cli.Context) {
	err := requireProject(c, options{}).Create(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectUp(c *cli.Context) {
	err := requireProject(c, options{
		log: !c.Bool("d"),
	}).Up(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}

	if !c.Bool("d") {
		wait()
	}
}

func ProjectRestart(c *cli.Context) {
	err := requireProject(c, options{}).Restart(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func ProjectLog(c *cli.Context) {
	err := requireProject(c, options{log: true}).Log(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
	wait()
}

func ProjectDelete(c *cli.Context) {
	if !c.Bool("force") && len(c.Args()) == 0 {
		logrus.Fatal("Will not remove all services with out --force")
	}
	err := requireProject(c, options{}).Delete(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

func requireProject(c *cli.Context, opts options) *project.Project {
	context := newContext(c, opts)
	project, err := rancher.NewProject(context)
	if err != nil {
		logrus.Fatal(err)
	}

	return project
}

func Scale(c *cli.Context) {
	p := requireProject(c, options{})

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
