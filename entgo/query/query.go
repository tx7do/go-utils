package entgo

import (
	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/encoding"
	_ "github.com/go-kratos/kratos/v2/encoding/json"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
)

// parseJsonMap 解析JSON字符串里面的MAP，包含Array形式的Map
func parseJsonMap(strJson []byte, retMap *map[string]string) error {
	codec := encoding.GetCodec("json")

	var err error
	if err = codec.Unmarshal(strJson, retMap); err != nil {
		var retArray []map[string]string
		if err1 := codec.Unmarshal(strJson, &retArray); err1 == nil {
			for _, itemA := range retArray {
				for k, v := range itemA {
					(*retMap)[k] = v
				}
			}
		} else {
			return err
		}
	}

	return nil
}

// BuildQuerySelector 构建分页查询选择器
func BuildQuerySelector(dbDriverName string,
	andFilterJsonString, orFilterJsonString string,
	page, pageSize int32, noPaging bool,
	orderBys []string, defaultOrderField string) (err error, whereSelectors []func(s *sql.Selector), querySelectors []func(s *sql.Selector)) {
	err, whereSelectors = BuildFilterSelector(andFilterJsonString, orFilterJsonString)
	if err != nil {
		return err, nil, nil
	}

	var orderSelector func(s *sql.Selector)
	err, orderSelector = BuildOrderSelector(orderBys, defaultOrderField)
	if err != nil {
		return err, nil, nil
	}

	pageSelector := BuildPaginationSelector(page, pageSize, noPaging)

	if len(whereSelectors) > 0 {
		querySelectors = append(querySelectors, whereSelectors...)
	}

	if orderSelector != nil {
		querySelectors = append(querySelectors, orderSelector)
	}
	if pageSelector != nil {
		querySelectors = append(querySelectors, pageSelector)
	}

	return
}
