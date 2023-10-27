package entgo

import (
	"entgo.io/ent/dialect/sql"
	_ "github.com/go-kratos/kratos/v2/encoding/json"
)

// BuildQuerySelector 构建分页查询选择器
func BuildQuerySelector(
	dbDriverName string,
	andFilterJsonString, orFilterJsonString string,
	page, pageSize int32, noPaging bool,
	orderBys []string, defaultOrderField string,
) (err error, whereSelectors []func(s *sql.Selector), querySelectors []func(s *sql.Selector)) {
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
