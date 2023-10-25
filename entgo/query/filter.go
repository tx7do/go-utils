package entgo

import (
	"encoding/json"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/tx7do/kratos-utils/stringcase"
)

const (
	FilterNot                   = "not"         // 不是
	FilterIn                    = "in"          // 检查字段值是否在传递列表中
	FilterNotIn                 = "not_in"      // 不在列表中
	FilterGTE                   = "gte"         // 大于或等于传递的值
	FilterGT                    = "gt"          // 大于传递值
	FilterLTE                   = "lte"         // 低于或等于传递值
	FilterLT                    = "lt"          // 低于传递值
	FilterRange                 = "range"       // 介于和给定的两个值之间
	FilterIsNull                = "isnull"      // 字段为空
	FilterNotIsNull             = "not_isnull"  // 字段不为空
	FilterContains              = "contains"    // 字段包含指定的子字符串
	FilterInsensitiveContains   = "icontains"   // 不区分大小写,字段包含指定的子字符串
	FilterStartsWith            = "startswith"  // 如果字段以值开头
	FilterInsensitiveStartsWith = "istartswith" // 不区分大小写,如果字段以值开头
	FilterEndsWith              = "endswith"    // 如果字段以值结尾
	FilterInsensitiveEndsWith   = "iendswith"   // 不区分大小写,如果字段以值结尾
	FilterExact                 = "exact"       // 等于
	FilterInsensitiveExact      = "iexact"      // 不区分大小写等于
	FilterSearch                = "search"      // 全文搜索
)

const (
	FilterDatePartYear        = "year"        // 年
	FilterDatePartQuarter     = "quarter"     // 季度
	FilterDatePartMonth       = "month"       // 月
	FilterDatePartWeek        = "week"        // 星期
	FilterDatePartDay         = "day"         // 日
	FilterDatePartHour        = "hour"        // 小时
	FilterDatePartMinute      = "minute"      // 分钟
	FilterDatePartSecond      = "second"      // 秒
	FilterDatePartMicrosecond = "microsecond" // 微秒
)

// QueryCommandToWhereConditions 查询命令转换为选择条件
func QueryCommandToWhereConditions(strJson string, isOr bool) (error, func(s *sql.Selector)) {
	if len(strJson) == 0 {
		return nil, nil
	}

	codec := encoding.GetCodec("json")

	queryMap := make(map[string]string)
	var queryMapArray []map[string]string
	if err1 := codec.Unmarshal([]byte(strJson), &queryMap); err1 != nil {
		if err2 := codec.Unmarshal([]byte(strJson), &queryMapArray); err2 != nil {
			return err2, nil
		}
	}

	return nil, func(s *sql.Selector) {
		var ps []*sql.Predicate
		ps = append(ps, processQueryMap(s, queryMap)...)
		for _, v := range queryMapArray {
			ps = append(ps, processQueryMap(s, v)...)
		}

		if isOr {
			s.Where(sql.Or(ps...))
		} else {
			s.Where(sql.And(ps...))
		}
	}
}

func processQueryMap(s *sql.Selector, queryMap map[string]string) []*sql.Predicate {
	var ps []*sql.Predicate
	for k, v := range queryMap {
		key := stringcase.ToSnakeCase(k)

		keys := strings.Split(key, "__")

		if cond := oneFieldFilter(s, keys, v); cond != nil {
			ps = append(ps, cond)
		}
	}

	return ps
}

func BuildFilterSelector(andFilterJsonString, orFilterJsonString string) (error, []func(s *sql.Selector)) {
	var err error
	var queryConditions []func(s *sql.Selector)

	var andSelector func(s *sql.Selector)
	err, andSelector = QueryCommandToWhereConditions(andFilterJsonString, false)
	if err != nil {
		return err, nil
	}
	if andSelector != nil {
		queryConditions = append(queryConditions, andSelector)
	}

	var orSelector func(s *sql.Selector)
	err, orSelector = QueryCommandToWhereConditions(orFilterJsonString, true)
	if err != nil {
		return err, nil
	}
	if orSelector != nil {
		queryConditions = append(queryConditions, orSelector)
	}

	return nil, queryConditions
}

func oneFieldFilter(s *sql.Selector, keys []string, value string) *sql.Predicate {
	var cond *sql.Predicate

	if len(keys) == 1 {
		field := keys[0]
		cond = filterEqual(s, field, value)
	} else if len(keys) == 2 {
		if len(keys[0]) == 0 {
			return nil
		}
		field := keys[0]
		op := strings.ToLower(keys[1])
		switch op {
		case FilterNot:
			cond = filterNot(s, field, value)
		case FilterIn:
			cond = filterIn(s, field, value)
		case FilterNotIn:
			cond = filterNotIn(s, field, value)
		case FilterGTE:
			cond = filterGTE(s, field, value)
		case FilterGT:
			cond = filterGT(s, field, value)
		case FilterLTE:
			cond = filterLTE(s, field, value)
		case FilterLT:
			cond = filterLT(s, field, value)
		case FilterRange:
			cond = filterRange(s, field, value)
		case FilterIsNull:
			cond = filterIsNull(s, field, value)
		case FilterNotIsNull:
			cond = filterNotIsNull(s, field, value)
		case FilterContains:
			cond = filterContains(s, field, value)
		case FilterInsensitiveContains:
			cond = filterInsensitiveContains(s, field, value)
		case FilterStartsWith:
			cond = filterStartsWith(s, field, value)
		case FilterInsensitiveStartsWith:
			cond = filterInsensitiveStartsWith(s, field, value)
		case FilterEndsWith:
			cond = filterEndsWith(s, field, value)
		case FilterInsensitiveEndsWith:
			cond = filterInsensitiveEndsWith(s, field, value)
		case FilterInsensitiveExact:
			cond = filterInsensitiveExact(s, field, value)
		case FilterExact:
			cond = filterExact(s, field, value)
		case FilterSearch:
			cond = filterSearch(s, field, value)
		default:
			cond = filterDatePart(s, op, field, value)
		}
	}
	return cond
}

// filterEqual 相等 WHERE "name" = $1
func filterEqual(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EQ(s.C(field), value)
}

// filterNot NOT操作 WHERE NOT ("name" = $1 AND "age" = $2)
func filterNot(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.Not(sql.EQ(s.C(field), value))
}

// filterIn IN操作
func filterIn(s *sql.Selector, field, value string) *sql.Predicate {
	var strs []string
	if err := json.Unmarshal([]byte(value), &strs); err == nil {
		return sql.In(s.C(field), strs)
	}

	var float64s []float64
	if err := json.Unmarshal([]byte(value), &float64s); err == nil {
		return sql.In(s.C(field), strs)
	}

	return nil
}

// filterNotIn 操作
func filterNotIn(s *sql.Selector, field, value string) *sql.Predicate {
	var strs []string
	if err := json.Unmarshal([]byte(value), &strs); err == nil {
		return sql.NotIn(s.C(field), strs)
	}

	var float64s []float64
	if err := json.Unmarshal([]byte(value), &float64s); err == nil {
		return sql.NotIn(s.C(field), strs)
	}

	return nil
}

// filterGTE 操作
func filterGTE(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.GTE(s.C(field), value)
}

// filterGT 操作
func filterGT(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.GT(s.C(field), value)
}

// filterLTE 操作
func filterLTE(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.LTE(s.C(field), value)
}

// filterLT 操作
func filterLT(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.LT(s.C(field), value)
}

// filterRange 操作
func filterRange(s *sql.Selector, field, value string) *sql.Predicate {
	var strs []string
	if err := json.Unmarshal([]byte(value), &strs); err == nil {
		if len(strs) != 2 {
			return nil
		}

		return sql.And(
			sql.GTE(s.C(field), strs[0]),
			sql.LTE(s.C(field), strs[1]),
		)
	}

	var float64s []float64
	if err := json.Unmarshal([]byte(value), &float64s); err == nil {
		if len(float64s) != 2 {
			return nil
		}

		return sql.And(
			sql.GTE(s.C(field), float64s[0]),
			sql.LTE(s.C(field), float64s[1]),
		)
	}

	return nil
}

// filterIsNull 操作
func filterIsNull(s *sql.Selector, field, _ string) *sql.Predicate {
	return sql.IsNull(s.C(field))
}

// filterNotIsNull 操作
func filterNotIsNull(s *sql.Selector, field, _ string) *sql.Predicate {
	return sql.Not(sql.IsNull(s.C(field)))
}

// filterContains 前后模糊查询 WHERE city LIKE '%L%';
func filterContains(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.Contains(s.C(field), value)
}

// filterInsensitiveContains 前后模糊查询 WHERE city ILIKE '%L%';
func filterInsensitiveContains(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.ContainsFold(s.C(field), value)
}

// filterStartsWith 前缀+模糊查询 WHERE CustomerName LIKE 'La%';
func filterStartsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.HasPrefix(s.C(field), value)
}

// filterInsensitiveStartsWith 前缀+模糊查询 WHERE CustomerName ILIKE 'La%';
func filterInsensitiveStartsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EqualFold(s.C(field), value+"%")
}

// filterEndsWith 后缀+模糊查询 WHERE CustomerName LIKE '%a';
func filterEndsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.HasSuffix(s.C(field), value)
}

// filterInsensitiveEndsWith 后缀+模糊查询 WHERE CustomerName ILIKE '%a';
func filterInsensitiveEndsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EqualFold(s.C(field), "%"+value)
}

// filterInsensitiveExact 操作 WHERE CustomerName ILIKE 'a';
func filterInsensitiveExact(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EqualFold(s.C(field), value)
}

// filterInsensitiveExact 操作 WHERE CustomerName LIKE 'a';
func filterExact(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.Like(s.C(field), value)
}

// filterSearch 全文搜索
func filterSearch(s *sql.Selector, _, _ string) *sql.Predicate {
	return nil
}

// filterDatePart 时间戳提取日期 select extract(quarter from timestamp '2018-08-15 12:10:10');
func filterDatePart(s *sql.Selector, datePart, field, value string) *sql.Predicate {
	return nil
}
