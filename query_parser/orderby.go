package query_parser

import (
	"strings"
)

type OrderByHandler func(field string, desc bool)

// ParseOrderByString 解析排序字符串，调用处理函数。
func ParseOrderByString(orderBy string, handler OrderByHandler) error {
	if orderBy == "" {
		return nil
	}

	parts := strings.Split(orderBy, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		ParseOrderByField(part, handler)
	}

	return nil
}

// ParseOrderByStrings 解析多个排序字符串，调用处理函数。
func ParseOrderByStrings(orderBys []string, handler OrderByHandler) error {
	for _, v := range orderBys {
		if v == "" {
			continue
		}

		ParseOrderByField(v, handler)
	}
	return nil
}

// ParseOrderByField 解析单个排序字段，调用处理函数。
func ParseOrderByField(orderBy string, handler OrderByHandler) {
	orderBy = strings.TrimSpace(orderBy)
	if orderBy == "" {
		return // 没有排序条件
	}

	if strings.HasPrefix(orderBy, "-") {
		handler(orderBy[1:], true) // 降序
	} else if strings.HasPrefix(orderBy, "+") {
		handler(orderBy[1:], false) // 升序
	} else {
		handler(orderBy, false) // 升序
	}
}
