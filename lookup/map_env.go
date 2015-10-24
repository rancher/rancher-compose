package lookup

import (
	"fmt"

	"github.com/docker/libcompose/project"
)

type MapEnvLookup struct {
	Env map[string]interface{}
}

func (m *MapEnvLookup) Lookup(key, serviceName string, config *project.ServiceConfig) []string {
	if v, ok := m.Env[key]; ok {
		return []string{fmt.Sprintf("%s=%v", key, v)}
	}
	return []string{}
}
