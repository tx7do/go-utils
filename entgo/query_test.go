package entgo

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-kratos/kratos/v2/encoding"
	_ "github.com/go-kratos/kratos/v2/encoding/json"
)

func TestKratosJsonCodec(t *testing.T) {
	var req struct {
		Query map[string]string `json:"query,omitempty"`
	}
	req.Query = make(map[string]string)
	req.Query["key1"] = "val1"
	req.Query["key2"] = "val2"

	codec := encoding.GetCodec("json")

	var err error
	var data []byte
	data, err = codec.Marshal(req)
	assert.Nil(t, err)
	fmt.Println(string(data))

	err = codec.Unmarshal(data, &req)
	assert.Nil(t, err)

	data1 := `{"query":{"key1":"val1","key2":"val2"}}`
	err = codec.Unmarshal([]byte(data1), &req)
	assert.Nil(t, err)

	//data2 := `{"query":[{"key1":"val1"},{"key2":"val2"}]}`
	//err = codec.Unmarshal([]byte(data2), &req)
	//assert.Nil(t, err)
}

func TestJsonCodec(t *testing.T) {
	var req struct {
		Query map[string]string `json:"query,omitempty"`
	}
	req.Query = make(map[string]string)
	req.Query["key1"] = "val1"
	req.Query["key2"] = "val2"

	var err error
	var data []byte
	data, err = json.Marshal(req)
	assert.Nil(t, err)
	fmt.Println(string(data))

	err = json.Unmarshal(data, &req)
	assert.Nil(t, err)

	data1 := `{"query":{"key1":"val1","key2":"val2"}}`
	err = json.Unmarshal([]byte(data1), &req)
	assert.Nil(t, err)

	data2 := `[1.0,2,3]`
	var float64s []float64
	err = json.Unmarshal([]byte(data2), &float64s)
	assert.Nil(t, err)
	fmt.Println(float64s)

	data3 := `["1.0","2","3"]`
	var strs []string
	err = json.Unmarshal([]byte(data3), &strs)
	assert.Nil(t, err)
	fmt.Println(strs)

	data4 := `{"key1":"val1", "key1":"val2", "key2":"val2"}`
	var mapstrs map[string]string
	err = json.Unmarshal([]byte(data4), &mapstrs)
	assert.Nil(t, err)
	fmt.Println(mapstrs)
}

func TestParseJsonMap(t *testing.T) {
	var err error

	req1 := make(map[string]string)
	data1 := `{"key1":"val1", "key1":"val2", "key2":"val2"}`
	err = parseJsonMap([]byte(data1), &req1)
	assert.Nil(t, err)
	fmt.Println(req1)

	req2 := make(map[string]string)
	data2 := `[{"key1":"val1"},{"key2":"val2"}]`
	err = parseJsonMap([]byte(data2), &req2)
	assert.Nil(t, err)
	fmt.Println(req1)
}

func TestSplitQuery(t *testing.T) {
	var keys []string

	keys = strings.Split("id", "__")
	assert.Equal(t, len(keys), 1)
	assert.Equal(t, keys[0], "id")

	keys = strings.Split("id__not", "__")
	assert.Equal(t, len(keys), 2)
	assert.Equal(t, keys[0], "id")
	assert.Equal(t, keys[1], "not")
}
