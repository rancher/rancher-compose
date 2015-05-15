package docker

import (
	"reflect"
	"testing"
	shlex "github.com/flynn/go-shlex"
)

func TestParseCommand(t *testing.T) {
	exp := []string{"sh", "-c", "exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}"}
	cmd, err := shlex.Split("sh -c 'exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}'")
	switch {
	case err != nil:
		t.Errorf("Error: %v\n", err)
	case !reflect.DeepEqual(cmd, exp):
		t.Errorf("got: %v\n", cmd)
		t.Errorf("expected: %v\n", exp)
	}
}
