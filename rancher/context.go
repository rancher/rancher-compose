package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
	rancherClient "github.com/rancher/go-rancher/client"
)

var projectRegexp = regexp.MustCompile("[^a-zA-Z0-9-]")

type Context struct {
	project.Context

	RancherConfig       map[string]RancherConfig
	RancherComposeFile  string
	RancherComposeBytes []byte
	Url                 string
	AccessKey           string
	SecretKey           string
	Client              *rancherClient.RancherClient
	Environment         *rancherClient.Environment
	isOpen              bool
	SidekickInfo        *SidekickInfo
	Uploader            Uploader
	PullCached          bool
}

type RancherConfig struct {
	Scale              int                                `yaml:"scale,omitempty"`
	LoadBalancerConfig *rancherClient.LoadBalancerConfig  `yaml:"load_balancer_config,omitempty"`
	ExternalIps        []string                           `yaml:"external_ips,omitempty"`
	Hostname           string                             `yaml:"hostname,omitempty"`
	HealthCheck        *rancherClient.InstanceHealthCheck `yaml:"health_check,omitempty"`
	DefaultCert        string                             `yaml:"default_cert,omitempty"`
	Certs              []string                           `yaml:"certs,omitempty"`
	Metadata           map[string]interface{}             `yaml:"metadata,omitempty"`
}

func (c *Context) readRancherConfig() error {
	if c.RancherComposeBytes == nil && c.RancherComposeFile == "" && c.ComposeFile != "" {
		f, err := filepath.Abs(c.ComposeFile)
		if err != nil {
			return err
		}

		c.RancherComposeFile = path.Join(path.Dir(f), "rancher-compose.yml")
	}

	if c.RancherComposeBytes == nil {
		logrus.Debugf("Opening rancher-compose file: %s", c.RancherComposeFile)
		if composeBytes, err := ioutil.ReadFile(c.RancherComposeFile); os.IsNotExist(err) {
			logrus.Debugf("Not found: %s", c.RancherComposeFile)
			return nil
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.RancherComposeFile)
			return err
		} else {
			c.RancherComposeBytes = composeBytes
		}
	}

	return c.unmarshalBytes(c.RancherComposeBytes)
}

func (c *Context) unmarshalBytes(bytes []byte) error {
	rawServiceMap := project.RawServiceMap{}
	if err := yaml.Unmarshal(bytes, &rawServiceMap); err != nil {
		return err
	}
	if err := project.Interpolate(c.EnvironmentLookup, &rawServiceMap); err != nil {
		return err
	}
	return utils.Convert(rawServiceMap, &c.RancherConfig)
}

func (c *Context) fixUpProjectName() {
	c.ProjectName = projectRegexp.ReplaceAllString(strings.ToLower(c.ProjectName), "-")

	// length can not be zero because libcompose would have failed before this
	if strings.ContainsAny(c.ProjectName[0:1], "_.-") {
		c.ProjectName = "x" + c.ProjectName
	}
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	c.fixUpProjectName()

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

	if envSchema, ok := c.Client.Types["environment"]; !ok || !Contains(envSchema.CollectionMethods, "POST") {
		return fmt.Errorf("Can not create a stack, check API key [%s] for [%s]", c.AccessKey, c.Url)
	}

	if err := c.loadEnv(); err != nil {
		return err
	}

	c.isOpen = true
	return nil
}

func (c *Context) loadEnv() error {
	if c.Environment != nil {
		return nil
	}

	logrus.Debugf("Looking for stack %s", c.ProjectName)
	// First try by name
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

	// Now try not by name for case sensitive databases
	envs, err = c.Client.Environment.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
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
