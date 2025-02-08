package entgo

import (
	"encoding/json"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/go-kratos/kratos/v2/encoding"
	_ "github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/stretchr/testify/assert"
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

	keys = splitQueryKey("id")
	assert.Equal(t, len(keys), 1)
	assert.Equal(t, keys[0], "id")

	keys = splitQueryKey("id__not")
	assert.Equal(t, len(keys), 2)
	assert.Equal(t, keys[0], "id")
	assert.Equal(t, keys[1], "not")
}

func TestBuildQuerySelectorDefault(t *testing.T) {
	testcases := []struct {
		name      string
		dialect   string
		and       string
		or        string
		noPaging  bool
		actualSql string
	}{
		{"MySQL_Pagination", dialect.MySQL, "", "", false, "SELECT * FROM `users` ORDER BY `users`.`created_at` DESC LIMIT 10 OFFSET 0"},
		{"PostgreSQL_Pagination", dialect.Postgres, "", "", false, "SELECT * FROM \"users\" ORDER BY \"users\".\"created_at\" DESC LIMIT 10 OFFSET 0"},

		{"MySQL_NoPagination", dialect.MySQL, "", "", true, "SELECT * FROM `users` ORDER BY `users`.`created_at` DESC"},
		{"PostgreSQL_NoPagination", dialect.Postgres, "", "", true, "SELECT * FROM \"users\" ORDER BY \"users\".\"created_at\" DESC"},

		{"MySQL_JsonbQuery", dialect.MySQL, "{\"preferences__daily_email\" : \"true\"}", "", true, "SELECT * FROM `users` WHERE JSON_EXTRACT(`users`.`preferences`, '$.daily_email') = ? ORDER BY `users`.`created_at` DESC"},
		{"PostgreSQL_JsonbQuery", dialect.Postgres, "{\"preferences__daily_email\" : \"true\"}", "", true, "SELECT * FROM \"users\" WHERE \"users\".\"preferences\" -> daily_email = $1 ORDER BY \"users\".\"created_at\" DESC"},

		{"MySQL_DatePartQuery", dialect.MySQL, "{\"created_at__date\" : \"2023-01-01\"}", "", true, "SELECT * FROM `users` WHERE DATE(`users`.`created_at`) = ? ORDER BY `users`.`created_at` DESC"},
		{"PostgreSQL_DatePartQuery", dialect.Postgres, "{\"created_at__date\" : \"2023-01-01\"}", "", true, "SELECT * FROM \"users\" WHERE EXTRACT('DATE' FROM \"users\".\"created_at\") = $1 ORDER BY \"users\".\"created_at\" DESC"},

		{"MySQL_JsonbCombineQuery", dialect.MySQL, "{\"preferences__pub_date__not\" : \"true\"}", "", true, "SELECT * FROM `users` WHERE NOT JSON_EXTRACT(`users`.`preferences`, '$.pub_date') = ? ORDER BY `users`.`created_at` DESC"},
		{"PostgreSQL_JsonbCombineQuery", dialect.Postgres, "{\"preferences__pub_date__not\" : \"true\"}", "", true, "SELECT * FROM \"users\" WHERE NOT \"users\".\"preferences\" -> pub_date = $1 ORDER BY \"users\".\"created_at\" DESC"},

		{"MySQL_DatePartCombineQuery", dialect.MySQL, "{\"pub_date__date__not\" : \"true\"}", "", true, "SELECT * FROM `users` WHERE NOT DATE(`users`.`pub_date`) = ? ORDER BY `users`.`created_at` DESC"},
		{"PostgreSQL_DatePartCombineQuery", dialect.Postgres, "{\"pub_date__date__not\" : \"true\"}", "", true, "SELECT * FROM \"users\" WHERE NOT EXTRACT('DATE' FROM \"users\".\"pub_date\") = $1 ORDER BY \"users\".\"created_at\" DESC"},

		{"MySQL_DatePartRangeQuery", dialect.MySQL, "{\"pub_date__date__range\" : \"[\\\"2023-10-25\\\", \\\"2024-10-25\\\"]\"}", "", true, "SELECT * FROM `users` WHERE DATE(`users`.`pub_date`) >= ? AND DATE(`users`.`pub_date`) <= ? ORDER BY `users`.`created_at` DESC"},
		{"PostgreSQL_DatePartRangeQuery", dialect.Postgres, "{\"pub_date__date__range\" : \"[\\\"2023-10-25\\\", \\\"2024-10-25\\\"]\"}", "", true, "SELECT * FROM \"users\" WHERE EXTRACT('DATE' FROM \"users\".\"pub_date\") >= $1 AND EXTRACT('DATE' FROM \"users\".\"pub_date\") <= $2 ORDER BY \"users\".\"created_at\" DESC"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			checker := assert.New(t)
			s := sql.Dialect(tc.dialect).Select("*").From(sql.Table("users"))

			err, _, querySelectors := BuildQuerySelector(tc.and, tc.or,
				1, 10, tc.noPaging,
				[]string{}, "created_at",
				[]string{},
			)
			checker.Nil(err)
			//checker.NotNil(whereSelectors)
			checker.NotNil(querySelectors)

			for _, fnc := range querySelectors {
				fnc(s)
			}

			query, _ := s.Query()
			checker.Equal(tc.actualSql, query)
			//checker.Empty(t, args)
		})
	}
}
