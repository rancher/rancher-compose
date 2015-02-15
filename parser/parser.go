package parser

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func YamlUnmarshal(filename string) (map[interface{}]interface{}, error) {
	yamlbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Could not open docker-compose file: %s", err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(yamlbytes), &m)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	return m, nil
}
