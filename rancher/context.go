package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	rancherClient "github.com/rancher/go-rancher/client"
)

type Context struct {
	project.Context

	RancherConfig      map[string]RancherConfig
	RancherComposeFile string
	Url                string
	AccessKey          string
	SecretKey          string
	Client             *rancherClient.RancherClient
	Environment        *rancherClient.Environment
	isOpen             bool
	SidekickInfo       *SidekickInfo
}

type RancherConfig struct {
	Scale              int                                `yaml:"scale,omitempty"`
	LoadBalancerConfig *rancherClient.LoadBalancerConfig  `yaml:"load_balancer_config,omitempty"`
	ExternalIps        []string                           `yaml:"external_ips,omitempty"`
	Hostname           string                             `yaml:"hostname,omitempty"`
	HealthCheck        *rancherClient.InstanceHealthCheck `yaml:"health_check,omitempty"`
}

func (c *Context) readRancherConfig() error {
	if c.RancherComposeFile == "" {
		f, err := filepath.Abs(c.ComposeFile)
		if err != nil {
			return err
		}

		c.RancherComposeFile = path.Join(path.Dir(f), "rancher-compose.yml")
	}

	logrus.Debugf("Opening rancher-compose file: %s", c.RancherComposeFile)

	if composeBytes, err := ioutil.ReadFile(c.RancherComposeFile); os.IsNotExist(err) {
		logrus.Debugf("Not found: %s", c.RancherComposeFile)
		return nil
	} else if err != nil {
		logrus.Errorf("Failed to open %s", c.RancherComposeFile)
		return err
	} else {
		return yaml.Unmarshal(composeBytes, &c.RancherConfig)
	}
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	if err := c.readRancherConfig(); err != nil {
		return err
	}

	if c.Url == "" {
		return fmt.Errorf("RANCHER_URL is not set")
	}

	if client, err := rancherClient.NewRancherClient(&rancherClient.ClientOpts{
		Url:       c.Url,
		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
	}); err != nil {
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
		if strings.EqualFold(c.ProjectName, env.Name) {
			logrus.Debugf("Found stack: %s(%s)", env.Name, env.Id)
			c.Environment = &env
			return nil
		}
	}

	logrus.Infof("Creating stack %s", c.ProjectName)
	env, err := c.Client.Environment.Create(&rancherClient.Environment{
		Name: c.ProjectName,
	})
	if err != nil {
		return err
	}

	c.Environment = env

	return nil
}
