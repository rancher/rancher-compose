package project

import (
	"github.com/rancherio/go-rancher/client"
	"gopkg.in/yaml.v2"
)

type Event string

const (
	CONTAINER_ID = "container_id"

	CONTAINER_CREATED = Event("Created container")
	CONTAINER_STARTED = Event("Started container")

	SERVICE_ADD      = Event("Adding")
	SERVICE_UP_START = Event("Starting")
	SERVICE_UP       = Event("Started")

	PROJECT_UP_START       = Event("Starting project")
	PROJECT_RELOAD         = Event("Reloading project")
	PROJECT_RELOAD_TRIGGER = Event("Triggering project reload")
)

type StringOnSlice struct {
	parts []string
}

func (e *StringOnSlice) MarshalYAML() ([]byte, error) {
	if e == nil {
		return []byte{}, nil
	}
	return yaml.Marshal(e.Slice())
}

// UnmarshalYAML decode the StringOnSlice whether it's a string or an array of strings.
func (e *StringOnSlice) UnmarshalYAML(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	p := make([]string, 0, 1)
	if err := yaml.Unmarshal(b, &p); err != nil {
		p = append(p, string(b))
	}
	e.parts = p
	return nil
}

func (e *StringOnSlice) Len() int {
	if e == nil {
		return 0
	}
	return len(e.parts)
}

func (e *StringOnSlice) Slice() []string {
	if e == nil {
		return nil
	}
	return e.parts
}

func NewEntrypoint(parts ...string) *StringOnSlice {
	return &StringOnSlice{parts}
}

type Dns1 struct {
	StringOnSlice
}

type ServiceConfig struct {
	CapAdd      []string `yaml:"cap_add,omitempty"`
	CapDrop     []string `yaml:"cap_drop,omitempty"`
	CpuShares   int64    `yaml:"cpu_shares,omitempty"`
	Command     string   `yaml:"command,omitempty"`
	Detach      string   `yaml:"detach,omitempty"`
	Dns         *Dns1
	DnsSearch   string   `yaml:"dns_search,omitempty"`
	DomainName  string   `yaml:"domainname,omitempty"`
	Entrypoint  string   `yaml:"entrypoint,omitempty"`
	EnvFile     string   `yaml:"env_file,omitempty"`
	Environment []string `yaml:"environment,omitempty"`
	Hostname    string   `yaml:"hostname,omitempty"`
	Image       string   `yaml:"image,omitempty"`
	Labels      []string `yaml:"labels,omitempty"`
	Links       []string `yaml:"links,omitempty"`
	MemLimit    int64    `yaml:"mem_limit,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Net         string   `yaml:"net,omitempty"`
	Pid         string   `yaml:"pid,omitempty"`
	Ipc         string   `yaml:"ipc,omitempty"`
	Ports       []string `yaml:"ports,omitempty"`
	Privileged  bool     `yaml:"privileged,omitempty"`
	Restart     string   `yaml:"restart,omitempty"`
	ReadOnly    bool     `yaml:"read_only,omitempty"`
	StdinOpen   bool     `yaml:"stdin_open,omitempty"`
	Tty         bool     `yaml:"tty,omitempty"`
	User        string   `yaml:"user,omitempty"`
	Volumes     []string `yaml:"volumes,omitempty"`
	VolumesFrom []string `yaml:"volumes_from,omitempty"`
	WorkingDir  string   `yaml:"working_dir,omitempty"`
	//`yaml:"build,omitempty"`
	Expose        []string `yaml:"expose,omitempty"`
	ExternalLinks []string `yaml:"external_links,omitempty"`
}

type Project struct {
	Name           string
	configs        map[string]*ServiceConfig
	Services       map[string]Service
	file           string
	content        []byte
	client         *client.RancherClient
	factory        ServiceFactory
	ReloadCallback func() error
	upCount        int
}

type Service interface {
	Name() string
	Up() error
	Config() *ServiceConfig
}

type ServiceFactory interface {
	Create(project *Project, name string, serviceConfig *ServiceConfig) (Service, error)
}
