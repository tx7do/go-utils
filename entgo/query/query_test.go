package entgo

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestBuildQuerySelectorDefault(t *testing.T) {
	t.Run("MySQL_Pagination", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		err, whereSelectors, querySelectors := BuildQuerySelector("", "",
			1, 10, false,
			[]string{}, "created_at",
			[]string{},
		)
		require.Nil(t, err)
		require.Nil(t, whereSelectors)
		require.NotNil(t, querySelectors)

		for _, fnc := range whereSelectors {
			fnc(s)
		}
		for _, fnc := range querySelectors {
			fnc(s)
		}

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` ORDER BY `users`.`created_at` DESC LIMIT 10 OFFSET 0", query)
		require.Empty(t, args)
	})
	t.Run("PostgreSQL_Pagination", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		err, whereSelectors, querySelectors := BuildQuerySelector("", "",
			1, 10, false,
			[]string{}, "created_at",
			[]string{},
		)
		require.Nil(t, err)
		require.Nil(t, whereSelectors)
		require.NotNil(t, querySelectors)

		for _, fnc := range whereSelectors {
			fnc(s)
		}
		for _, fnc := range querySelectors {
			fnc(s)
		}

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" ORDER BY \"users\".\"created_at\" DESC LIMIT 10 OFFSET 0", query)
		require.Empty(t, args)
	})

	t.Run("MySQL_NoPagination", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		err, whereSelectors, querySelectors := BuildQuerySelector("", "",
			1, 10, true,
			[]string{}, "created_at",
			[]string{},
		)
		require.Nil(t, err)
		require.Nil(t, whereSelectors)
		require.NotNil(t, querySelectors)

		for _, fnc := range whereSelectors {
			fnc(s)
		}
		for _, fnc := range querySelectors {
			fnc(s)
		}

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` ORDER BY `users`.`created_at` DESC", query)
		require.Empty(t, args)
	})
	t.Run("PostgreSQL_NoPagination", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		err, whereSelectors, querySelectors := BuildQuerySelector("", "",
			1, 10, true,
			[]string{}, "created_at",
			[]string{},
		)
		require.Nil(t, err)
		require.Nil(t, whereSelectors)
		require.NotNil(t, querySelectors)

		for _, fnc := range whereSelectors {
			fnc(s)
		}
		for _, fnc := range querySelectors {
			fnc(s)
		}

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" ORDER BY \"users\".\"created_at\" DESC", query)
		require.Empty(t, args)
	})
}
