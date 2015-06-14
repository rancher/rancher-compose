package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancherio/go-rancher/client"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

type Context struct {
	RancherConfig      map[string]RancherConfig
	Log                bool
	RancherComposeFile string
	ComposeFile        string
	ComposeBytes       []byte
	ProjectName        string
	Url                string
	AccessKey          string
	SecretKey          string
	Client             *rancherClient.RancherClient
	Environment        *rancherClient.Environment
	isOpen             bool
	Project            *project.Project
	SidekickInfo       *SidekickInfo
}

type RancherConfig struct {
	Scale              int                                `yaml:"scale,omitempty"`
	LoadBalancerConfig *rancherClient.LoadBalancerConfig  `yaml:"load_balancer_config,omitempty"`
	ExternalIps        []string                           `yaml:"external_ips,omitempty"`
	HealthCheck        *rancherClient.InstanceHealthCheck `yaml:"health_check,omitempty"`
}

func (c *Context) readComposeFile() error {
	if c.ComposeBytes != nil {
		return nil
	}

	logrus.Debugf("Opening compose file: %s", c.ComposeFile)

	if c.ComposeFile == "-" {
		if composeBytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			logrus.Errorf("Failed to read compose file from stdin: %v", err)
			return err
		} else {
			c.ComposeBytes = composeBytes
		}
	} else if composeBytes, err := ioutil.ReadFile(c.ComposeFile); os.IsNotExist(err) {
		logrus.Errorf("Failed to find %s", c.ComposeFile)
		return err
	} else if err != nil {
		logrus.Errorf("Failed to open %s", c.ComposeFile)
		return err
	} else {
		c.ComposeBytes = composeBytes
	}

	return nil
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

func (c *Context) determineProject() error {
	if c.ProjectName == "" {
		f, err := filepath.Abs(c.ComposeFile)
		if err != nil {
			logrus.Errorf("Failed to get absolute directory for: %s", c.ComposeFile)
			return err
		}

		parent := path.Base(path.Dir(f))
		if parent != "" {
			c.ProjectName = parent
		} else if wd, err := os.Getwd(); err != nil {
			return err
		} else {
			c.ProjectName = path.Base(wd)
		}
	}

	return nil
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	if err := c.readComposeFile(); err != nil {
		return err
	}

	if err := c.readRancherConfig(); err != nil {
		return err
	}

	if err := c.determineProject(); err != nil {
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

func (c *Context) Lookup(key, serviceName string, config *project.ServiceConfig) []string {
	ret := os.Getenv(key)
	if ret == "" {
		return []string{}
	} else {
		return []string{fmt.Sprintf("%s=%s", key, ret)}
	}
}
