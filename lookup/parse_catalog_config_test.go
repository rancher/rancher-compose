package lookup

import (
	"reflect"
	"testing"

	"github.com/rancher/rancher-catalog-service/model"
)

func testParseCatalog(t *testing.T, contents string, expectedCatalogConfig *model.RancherCompose) {
	catalogConfig, err := ParseCatalogConfig([]byte(contents))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedCatalogConfig, catalogConfig) {
		t.Fail()
	}
}

func TestParseCatalog(t *testing.T) {
	testParseCatalog(t, `
.catalog:
  name: test`, &model.RancherCompose{
		Name: "test",
	})

	testParseCatalog(t, `
version: '2'
catalog:
  name: test`, &model.RancherCompose{
		Name: "test",
	})

	testParseCatalog(t, `
version: '2'
.catalog:
  name: test`, &model.RancherCompose{
		Name: "test",
	})

	testParseCatalog(t, `
version: '2'
services:
  .catalog:
    name: test`, &model.RancherCompose{
		Name: "test",
	})
}
