package entgo

import (
	"strings"

	"entgo.io/ent/dialect/sql"
)

// QueryCommandToOrderConditions 查询命令转换为排序条件
func QueryCommandToOrderConditions(orderBys []string) (error, func(s *sql.Selector)) {
	if len(orderBys) == 0 {
		return nil, nil
	}

	return nil, func(s *sql.Selector) {
		for _, v := range orderBys {
			if strings.HasPrefix(v, "-") {
				// 降序
				key := v[1:]
				if len(key) == 0 {
					continue
				}

				s.OrderBy(sql.Desc(s.C(key)))
			} else {
				// 升序
				if len(v) == 0 {
					continue
				}

				s.OrderBy(sql.Asc(s.C(v)))
			}
		}
	}
}

func BuildOrderSelector(orderBys []string, defaultOrderField string) (error, func(s *sql.Selector)) {
	if len(orderBys) == 0 {
		return nil, func(s *sql.Selector) {
			s.OrderBy(sql.Desc(s.C(defaultOrderField)))
		}
	} else {
		return QueryCommandToOrderConditions(orderBys)
	}
}
