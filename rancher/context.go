package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancherio/go-rancher/client"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

type Context struct {
	ComposeFile  string
	ComposeBytes []byte
	ProjectName  string
	Url          string
	AccessKey    string
	SecretKey    string
	Client       *rancherClient.RancherClient
	Environment  *rancherClient.Environment
	isOpen       bool
	Project      *project.Project
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	if c.ComposeBytes == nil {
		logrus.Debugf("Opening compose file: %s", c.ComposeFile)

		if composeBytes, err := ioutil.ReadFile(c.ComposeFile); os.IsNotExist(err) {
			logrus.Errorf("Failed to find %s", c.ComposeFile)
			return err
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.ComposeFile)
			return err
		} else {
			c.ComposeBytes = composeBytes
		}
	}

	if c.ProjectName == "" {
		if wd, err := os.Getwd(); err != nil {
			return err
		} else {
			c.ProjectName = path.Base(wd)
		}
	}

	if c.Url == "" {
		return fmt.Errorf("RANCHER_URL is not set")
	}

	client, err := rancherClient.NewRancherClient(&rancherClient.ClientOpts{
		Url:       c.Url,
		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
	})

	if err != nil {
		return err
	} else {
		c.Client = client
	}

	if err := c.loadEnv(); err != nil {
		return err
	}

	c.isOpen = true
	return nil
}

func (c *Context) loadEnv() error {
	envs, err := c.Client.Environment.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"name":         c.ProjectName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return err
	}

	for _, env := range envs.Data {
		if c.ProjectName == env.Name {
			logrus.Debugf("Found environment: %s(%s)", env.Name, env.Id)
			c.Environment = &env
			return nil
		}
	}

	logrus.Infof("Creating environment %s", c.ProjectName)
	env, err := c.Client.Environment.Create(&rancherClient.Environment{
		Name: c.ProjectName,
	})
	if err != nil {
		return err
	}

	c.Environment = env

	return nil
}
