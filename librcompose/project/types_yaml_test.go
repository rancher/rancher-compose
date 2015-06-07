package project

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

type StructStringorslice struct {
	Foo Stringorslice
}

func TestStringorsliceYaml(t *testing.T) {
	str := `{foo: [bar, baz]}`

	s := StructStringorslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, []string{"bar", "baz"}, s.Foo.parts)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructStringorslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, []string{"bar", "baz"}, s2.Foo.parts)
}

type StructSliceorMap struct {
	Foo SliceorMap
}

func TestSliceOrMapYaml(t *testing.T) {
	str := `{foo: [bar=baz, far=faz]}`

	s := StructSliceorMap{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, map[string]string{"bar": "baz", "far": "faz"}, s.Foo.parts)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructSliceorMap{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, map[string]string{"bar": "baz", "far": "faz"}, s2.Foo.parts)
}

type StructMaporslice struct {
	Foo Maporslice
}

func contains(list []string, item string) bool {
	for _, test := range list {
		if test == item {
			return true
		}
	}
	return false
}

func TestMaporsliceYaml(t *testing.T) {
	str := `{foo: {bar: baz, far: faz}}`

	s := StructMaporslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, 2, len(s.Foo.parts))
	assert.True(t, contains(s.Foo.parts, "bar=baz"))
	assert.True(t, contains(s.Foo.parts, "far=faz"))

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructMaporslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, 2, len(s2.Foo.parts))
	assert.True(t, contains(s2.Foo.parts, "bar=baz"))
	assert.True(t, contains(s2.Foo.parts, "far=faz"))
}
