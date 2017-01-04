package rancher

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGenerateHAProxyConf(t *testing.T) {
	conf := generateHAProxyConf("daemon\nmaxconn 256", "mode http")
	expectedConf := `global
    daemon
    maxconn 256
defaults
    mode http`
	if conf != expectedConf {
		t.Fail()
	}

	conf = generateHAProxyConf("daemon\n", "")
	expectedConf = "global\n    daemon\n    \n"
	if conf != expectedConf {
		t.Fail()
	}

	conf = generateHAProxyConf("", "mode http")
	expectedConf = "defaults\n    mode http"
	if conf != expectedConf {
		t.Fail()
	}
}

func testRewritePorts(t *testing.T, in, out string) {
	updatedPorts, err := rewritePorts([]string{in})
	if err != nil {
		t.Fatal(err)
	}

	if len(updatedPorts) != 1 {
		t.Fail()
	}

	if updatedPorts[0] != out {
		t.Fail()
	}
}

func TestRewritePorts(t *testing.T) {
	testRewritePorts(t, "80", "80")
	testRewritePorts(t, "80/tcp", "80/tcp")
	testRewritePorts(t, "80:80", "80")
	testRewritePorts(t, "80:80/tcp", "80/tcp")
}

func testConvertLb(t *testing.T, ports, links, externalLinks []string, selector string, expectedPortRules []PortRule) {
	portRules, err := convertLb(ports, links, externalLinks, selector)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(portRules, expectedPortRules) {
		fmt.Println(portRules, expectedPortRules)
		t.Fail()
	}
}

func TestConvertLb(t *testing.T) {
	testConvertLb(t, []string{
		"8080:80",
	}, []string{"web1", "web2"}, []string{"external/web3"}, "", []PortRule{
		PortRule{
			SourcePort: 8080,
			TargetPort: 80,
			Service:    "web1",
			Protocol:   "http",
		},
		PortRule{
			SourcePort: 8080,
			TargetPort: 80,
			Service:    "web2",
			Protocol:   "http",
		},
		PortRule{
			SourcePort: 8080,
			TargetPort: 80,
			Service:    "external/web3",
			Protocol:   "http",
		},
	})

	testConvertLb(t, []string{
		"80",
	}, []string{"web1", "web2"}, []string{}, "", []PortRule{
		PortRule{
			SourcePort: 80,
			TargetPort: 80,
			Service:    "web1",
			Protocol:   "http",
		},
		PortRule{
			SourcePort: 80,
			TargetPort: 80,
			Service:    "web2",
			Protocol:   "http",
		},
	})

	testConvertLb(t, []string{
		"80/tcp",
	}, []string{"web1", "web2"}, []string{}, "", []PortRule{
		PortRule{
			SourcePort: 80,
			TargetPort: 80,
			Service:    "web1",
			Protocol:   "tcp",
		},
		PortRule{
			SourcePort: 80,
			TargetPort: 80,
			Service:    "web2",
			Protocol:   "tcp",
		},
	})

	testConvertLb(t, []string{
		"80/tcp",
	}, nil, nil, "foo=bar", []PortRule{
		PortRule{
			SourcePort: 80,
			TargetPort: 80,
			Selector:   "foo=bar",
			Protocol:   "tcp",
		},
	})
}

func testConvertLabel(t *testing.T, label string, expectedPortRules []PortRule) {
	portRules, err := convertLbLabel(label)
	if err != nil {
		t.Fail()
	}
	if !reflect.DeepEqual(portRules, expectedPortRules) {
		fmt.Println(portRules, expectedPortRules)
		t.Fail()
	}
}

func TestConvertLabel(t *testing.T) {
	testConvertLabel(t, "example2.com:80/path=81", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			SourcePort: 80,
			Path:       "/path",
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "example2.com:80/path/a", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			SourcePort: 80,
			Path:       "/path/a",
		},
	})
	testConvertLabel(t, "example2.com:80=81", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			SourcePort: 80,
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "example2.com:80", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			SourcePort: 80,
		},
	})
	testConvertLabel(t, "example2.com/path/b/c=81", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			Path:       "/path/b/c",
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "example2.com/path", []PortRule{
		PortRule{
			Hostname: "example2.com",
			Path:     "/path",
		},
	})
	testConvertLabel(t, "example2.com=81", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "example2.com", []PortRule{
		PortRule{
			Hostname: "example2.com",
		},
	})

	testConvertLabel(t, "80/path=81", []PortRule{
		PortRule{
			SourcePort: 80,
			Path:       "/path",
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "80/path", []PortRule{
		PortRule{
			SourcePort: 80,
			Path:       "/path",
		},
	})
	testConvertLabel(t, "80=81", []PortRule{
		PortRule{
			SourcePort: 80,
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "/path=81", []PortRule{
		PortRule{
			Path:       "/path",
			TargetPort: 81,
		},
	})
	testConvertLabel(t, "www.abc.com", []PortRule{
		PortRule{
			Hostname: "www.abc.com",
		},
	})
	testConvertLabel(t, "www.abc2.com", []PortRule{
		PortRule{
			Hostname: "www.abc2.com",
		},
	})
	testConvertLabel(t, "/path", []PortRule{
		PortRule{
			Path: "/path",
		},
	})
	testConvertLabel(t, "www.abc2.com/service.html", []PortRule{
		PortRule{
			Hostname: "www.abc2.com",
			Path:     "/service.html",
		},
	})
	testConvertLabel(t, "81", []PortRule{
		PortRule{
			TargetPort: 81,
		},
	})

	testConvertLabel(t, "81,82", []PortRule{
		PortRule{
			TargetPort: 81,
		},
		PortRule{
			TargetPort: 82,
		},
	})
	testConvertLabel(t, "example2.com:80/path=81,example2.com:82/path2=83", []PortRule{
		PortRule{
			Hostname:   "example2.com",
			SourcePort: 80,
			Path:       "/path",
			TargetPort: 81,
		},
		PortRule{
			Hostname:   "example2.com",
			SourcePort: 82,
			Path:       "/path2",
			TargetPort: 83,
		},
	})
}

func testMergePortRules(t *testing.T, baseRules, overrideRules, expectedPortRules []PortRule) {
	portRules := mergePortRules(baseRules, overrideRules)
	if !reflect.DeepEqual(portRules, expectedPortRules) {
		fmt.Println(portRules, expectedPortRules)
		t.Fail()
	}
}

func TestMergePortRules(t *testing.T) {
	testMergePortRules(t, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 80,
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 81,
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 81,
		},
	})

	testMergePortRules(t, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 80,
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			Path:       "/path",
			SourcePort: 80,
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			Path:       "/path",
			SourcePort: 80,
			TargetPort: 80,
		},
	})

	testMergePortRules(t, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web",
			SourcePort: 81,
			TargetPort: 81,
		},
	}, []PortRule{
		PortRule{
			Service: "web",
			Path:    "/path",
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			Path:       "/path",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web",
			Path:       "/path",
			SourcePort: 81,
			TargetPort: 81,
		},
	})

	testMergePortRules(t, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web",
			SourcePort: 81,
			TargetPort: 81,
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			TargetPort: 90,
			Hostname:   "www.example2.com",
			Path:       "/path",
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			Hostname:   "www.example2.com",
			Path:       "/path",
			SourcePort: 80,
			TargetPort: 90,
		},
		PortRule{
			Service:    "web",
			Hostname:   "www.example2.com",
			Path:       "/path",
			SourcePort: 81,
			TargetPort: 90,
		},
	})

	testMergePortRules(t, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 80,
		},
	}, []PortRule{
		PortRule{
			Service:  "web",
			Hostname: "www.example1.com",
			Path:     "/path1",
		},
		PortRule{
			Service:  "web",
			Hostname: "www.example2.com",
			Path:     "/path2",
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			Hostname:   "www.example1.com",
			Path:       "/path1",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web",
			Hostname:   "www.example2.com",
			Path:       "/path2",
			SourcePort: 80,
			TargetPort: 80,
		},
	})

	testMergePortRules(t, []PortRule{
		PortRule{
			Service:    "web",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web2",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web3",
			SourcePort: 80,
			TargetPort: 80,
		},
	}, []PortRule{
		PortRule{
			Service:  "web",
			Hostname: "www.example1.com",
			Path:     "/path1",
		},
		PortRule{
			Service:  "web",
			Hostname: "www.example2.com",
			Path:     "/path2",
		},
		PortRule{
			Service:    "web3",
			TargetPort: 90,
		},
	}, []PortRule{
		PortRule{
			Service:    "web",
			Hostname:   "www.example1.com",
			Path:       "/path1",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web",
			Hostname:   "www.example2.com",
			Path:       "/path2",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web2",
			SourcePort: 80,
			TargetPort: 80,
		},
		PortRule{
			Service:    "web3",
			SourcePort: 80,
			TargetPort: 90,
		},
	})
}
