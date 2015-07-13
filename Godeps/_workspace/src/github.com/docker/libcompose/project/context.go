package project

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/logger"
)

type Context struct {
	Timeout             int
	Log                 bool
	ComposeFile         string
	ComposeBytes        []byte
	AnswersFile         string
	Answers             map[string]string
	ProjectName         string
	isOpen              bool
	ServiceFactory      ServiceFactory
	EnvironmentLookup   EnvironmentLookup
	ConfigLookup        ConfigLookup
	LoggerFactory       logger.Factory
	IgnoreMissingConfig bool
	Project             *Project
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
	} else if c.ComposeFile != "" {
		if composeBytes, err := ioutil.ReadFile(c.ComposeFile); os.IsNotExist(err) {
			if c.IgnoreMissingConfig {
				return nil
			}
			logrus.Errorf("Failed to find %s", c.ComposeFile)
			return err
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.ComposeFile)
			return err
		} else {
			c.ComposeBytes = composeBytes
		}
	}

	return nil
}

func (c *Context) readAnswersFile() error {
	if c.AnswersFile != "" {
		logrus.Debugf("Opening answers file: %s", c.AnswersFile)
		if answersBytes, err := ioutil.ReadFile(c.AnswersFile); os.IsNotExist(err) {
			logrus.Errorf("Failed to find %s", c.AnswersFile)
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.AnswersFile)
			return err
		} else {
			return yaml.Unmarshal(answersBytes, &c.Answers)
		}
	}

	return nil
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

	if err := c.determineProject(); err != nil {
		return err
	}

	if err := c.readAnswersFile(); err != nil {
		return err
	}

	c.isOpen = true
	return nil
}
