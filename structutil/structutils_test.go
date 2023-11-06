package structutil_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tx7do/go-utils/structutil"
)

type TestStruct struct {
	First  string
	Second int
	Third  bool `struct:"third"`
}

func TestForEach(t *testing.T) {
	instance := TestStruct{
		"moishe",
		22,
		true,
	}
	structutil.ForEach(instance, func(key string, value any, tag reflect.StructTag) {
		switch key {
		case "First":
			assert.Equal(t, "moishe", value.(string))
			assert.Zero(t, tag)
		case "Second":
			assert.Equal(t, 22, value.(int))
			assert.Zero(t, tag)
		case "Third":
			assert.Equal(t, true, value.(bool))
			assert.Equal(t, "third", tag.Get("struct"))
		}
	})
}

func TestToMap(t *testing.T) {
	instance := TestStruct{
		"moishe",
		22,
		true,
	}
	assert.Equal(t, map[string]any{
		"First":  "moishe",
		"Second": 22,
		"Third":  true,
	}, structutil.ToMap(instance))

	assert.Equal(t, map[string]any{
		"First":  "moishe",
		"Second": 22,
		"third":  true,
	}, structutil.ToMap(instance, "struct"))
}
