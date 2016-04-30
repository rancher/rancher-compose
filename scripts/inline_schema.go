package main

import (
	"io/ioutil"
	"os"
	"text/template"
)

func main() {
	t, err := template.New("schema_template.go").Parse(`
package config

var schemaString = ` + "`{{.schemaString}}`")

	if err != nil {
		panic(err)
	}

	schema, err := ioutil.ReadFile("./scripts/config_schema_v1.json")

	inlinedFile, err := os.Create("vendor/github.com/docker/libcompose/config/schema.go")

	if err != nil {
		panic(err)
	}

	err = t.Execute(inlinedFile, map[string]string{
		"schemaString": string(schema),
	})

	if err != nil {
		panic(err)
	}
}
