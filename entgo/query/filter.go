package entgo

import (
	"encoding/json"
	"strings"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/go-kratos/kratos/v2/encoding"

	"github.com/tx7do/go-utils/stringcase"
)

type FilterOp int

const (
	FilterNot                   = "not"         // 不等于
	FilterIn                    = "in"          // 检查值是否在列表中
	FilterNotIn                 = "not_in"      // 不在列表中
	FilterGTE                   = "gte"         // 大于或等于传递的值
	FilterGT                    = "gt"          // 大于传递值
	FilterLTE                   = "lte"         // 小于或等于传递值
	FilterLT                    = "lt"          // 小于传递值
	FilterRange                 = "range"       // 是否介于和给定的两个值之间
	FilterIsNull                = "isnull"      // 是否为空
	FilterNotIsNull             = "not_isnull"  // 是否不为空
	FilterContains              = "contains"    // 是否包含指定的子字符串
	FilterInsensitiveContains   = "icontains"   // 不区分大小写，是否包含指定的子字符串
	FilterStartsWith            = "startswith"  // 以值开头
	FilterInsensitiveStartsWith = "istartswith" // 不区分大小写，以值开头
	FilterEndsWith              = "endswith"    // 以值结尾
	FilterInsensitiveEndsWith   = "iendswith"   // 不区分大小写，以值结尾
	FilterExact                 = "exact"       // 精确匹配
	FilterInsensitiveExact      = "iexact"      // 不区分大小写，精确匹配
	FilterRegex                 = "regex"       // 正则表达式
	FilterInsensitiveRegex      = "iregex"      // 不区分大小写，正则表达式
	FilterSearch                = "search"      // 全文搜索
)

type DatePart int

const (
	DatePartDate        DatePart = iota // 日期
	DatePartYear                        // 年
	DatePartISOYear                     // ISO 8601 一年中的周数
	DatePartQuarter                     // 季度
	DatePartMonth                       // 月
	DatePartWeek                        // ISO 8601 周编号 一年中的周数
	DatePartWeekDay                     // 星期几
	DatePartISOWeekDay                  // 星期几
	DatePartDay                         // 日
	DatePartTime                        // 小时：分钟：秒
	DatePartHour                        // 小时
	DatePartMinute                      // 分钟
	DatePartSecond                      // 秒
	DatePartMicrosecond                 // 微秒
)

var dateParts = [...]string{
	DatePartDate:        "date",
	DatePartYear:        "year",
	DatePartISOYear:     "iso_year",
	DatePartQuarter:     "quarter",
	DatePartMonth:       "month",
	DatePartWeek:        "week",
	DatePartWeekDay:     "week_day",
	DatePartISOWeekDay:  "iso_week_day",
	DatePartDay:         "day",
	DatePartTime:        "time",
	DatePartHour:        "hour",
	DatePartMinute:      "minute",
	DatePartSecond:      "second",
	DatePartMicrosecond: "microsecond",
}

func hasDatePart(str string) bool {
	for _, item := range dateParts {
		if str == item {
			return true
		}
	}
	return false
}

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
			cond = filterIsNotNull(s, field, value)
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
		case FilterExact:
			cond = filterExact(s, field, value)
		case FilterInsensitiveExact:
			cond = filterInsensitiveExact(s, field, value)
		case FilterRegex:
			cond = filterRegex(s, field, value)
		case FilterInsensitiveRegex:
			cond = filterInsensitiveRegex(s, field, value)
		case FilterSearch:
			cond = filterSearch(s, field, value)
		default:
			cond = filterDatePart(s, op, field, value)
		}
	}
	return cond
}

// filterEqual = 相等操作
// SQL: WHERE "name" = "tom"
func filterEqual(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EQ(s.C(field), value)
}

// filterNot NOT 不相等操作
// SQL: WHERE NOT ("name" = "tom")
// 或者： WHERE "name" <> "tom"
// 用NOT可以过滤出NULL，而用<>、!=则不能。
func filterNot(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.Not(sql.EQ(s.C(field), value))
}

// filterIn IN操作
// SQL: WHERE name IN ("tom", "jimmy")
func filterIn(s *sql.Selector, field, value string) *sql.Predicate {
	var values []any
	if err := json.Unmarshal([]byte(value), &values); err == nil {
		return sql.In(s.C(field), values...)
	}
	return nil
}

// filterNotIn NOT IN操作
// SQL: WHERE name NOT IN ("tom", "jimmy")`
func filterNotIn(s *sql.Selector, field, value string) *sql.Predicate {
	var values []any
	if err := json.Unmarshal([]byte(value), &values); err == nil {
		return sql.NotIn(s.C(field), values...)
	}
	return nil
}

// filterGTE GTE (Greater Than or Equal) 大于等于 >=操作
// SQL: WHERE "create_time" >= "2023-10-25"
func filterGTE(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.GTE(s.C(field), value)
}

// filterGT GT (Greater than) 大于 >操作
// SQL: WHERE "create_time" > "2023-10-25"
func filterGT(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.GT(s.C(field), value)
}

// filterLTE LTE (Less Than or Equal) 小于等于 <=操作
// SQL: WHERE "create_time" <= "2023-10-25"
func filterLTE(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.LTE(s.C(field), value)
}

// filterLT LT (Less than) 小于 <操作
// SQL: WHERE "create_time" < "2023-10-25"
func filterLT(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.LT(s.C(field), value)
}

// filterRange 在值域之中 BETWEEN操作
// SQL: WHERE "create_time" BETWEEN "2023-10-25" AND "2024-10-25"
// 或者： WHERE "create_time" >= "2023-10-25" AND "create_time" <= "2024-10-25"
func filterRange(s *sql.Selector, field, value string) *sql.Predicate {
	var values []any
	if err := json.Unmarshal([]byte(value), &values); err == nil {
		if len(values) != 2 {
			return nil
		}

		return sql.And(
			sql.GTE(s.C(field), values[0]),
			sql.LTE(s.C(field), values[1]),
		)
	}

	return nil
}

// filterIsNull 为空 IS NULL操作
// SQL: WHERE name IS NULL
func filterIsNull(s *sql.Selector, field, _ string) *sql.Predicate {
	return sql.IsNull(s.C(field))
}

// filterIsNotNull 不为空 IS NOT NULL操作
// SQL: WHERE name IS NOT NULL
func filterIsNotNull(s *sql.Selector, field, _ string) *sql.Predicate {
	return sql.Not(sql.IsNull(s.C(field)))
}

// filterContains LIKE 前后模糊查询
// SQL: WHERE name LIKE '%L%';
func filterContains(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.Contains(s.C(field), value)
}

// filterInsensitiveContains ILIKE 前后模糊查询
// SQL: WHERE name ILIKE '%L%';
func filterInsensitiveContains(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.ContainsFold(s.C(field), value)
}

// filterStartsWith LIKE 前缀+模糊查询
// SQL: WHERE name LIKE 'La%';
func filterStartsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.HasPrefix(s.C(field), value)
}

// filterInsensitiveStartsWith ILIKE 前缀+模糊查询
// SQL: WHERE name ILIKE 'La%';
func filterInsensitiveStartsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EqualFold(s.C(field), value+"%")
}

// filterEndsWith LIKE 后缀+模糊查询
// SQL: WHERE name LIKE '%a';
func filterEndsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.HasSuffix(s.C(field), value)
}

// filterInsensitiveEndsWith ILIKE 后缀+模糊查询
// SQL: WHERE name ILIKE '%a';
func filterInsensitiveEndsWith(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EqualFold(s.C(field), "%"+value)
}

// filterExact LIKE 操作 精确比对
// SQL: WHERE name LIKE 'a';
func filterExact(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.Like(s.C(field), value)
}

// filterInsensitiveExact ILIKE 操作 不区分大小写，精确比对
// SQL: WHERE name ILIKE 'a';
func filterInsensitiveExact(s *sql.Selector, field, value string) *sql.Predicate {
	return sql.EqualFold(s.C(field), value)
}

// filterRegex 正则查找
// MySQL: WHERE title REGEXP BINARY '^(An?|The) +'
// Oracle: WHERE REGEXP_LIKE(title, '^(An?|The) +', 'c');
// PostgreSQL: WHERE title ~ '^(An?|The) +';
// SQLite: WHERE title REGEXP '^(An?|The) +';
func filterRegex(s *sql.Selector, field, value string) *sql.Predicate {
	p := sql.P()
	p.Append(func(b *sql.Builder) {
		switch s.Builder.Dialect() {
		case dialect.Postgres:
			b.Ident(s.C(field)).WriteString(" ~ ")
			b.Arg(value)
			break
		case dialect.MySQL:
			b.Ident(s.C(field)).WriteString(" REGEXP BINARY ")
			b.Arg(value)
			break
		case dialect.SQLite:
			b.Ident(s.C(field)).WriteString(" REGEXP ")
			b.Arg(value)
			break
		case dialect.Gremlin:
			break
		}
	})
	return p
}

// filterInsensitiveRegex 正则查找 不区分大小写
// MySQL: WHERE title REGEXP '^(an?|the) +'
// Oracle: WHERE REGEXP_LIKE(title, '^(an?|the) +', 'i');
// PostgreSQL: WHERE title ~* '^(an?|the) +';
// SQLite: WHERE title REGEXP '(?i)^(an?|the) +';
func filterInsensitiveRegex(s *sql.Selector, field, value string) *sql.Predicate {
	p := sql.P()
	p.Append(func(b *sql.Builder) {
		switch s.Builder.Dialect() {
		case dialect.Postgres:
			b.Ident(s.C(field)).WriteString(" ~* ")
			b.Arg(strings.ToLower(value))
			break
		case dialect.MySQL:
			b.Ident(s.C(field)).WriteString(" REGEXP ")
			b.Arg(strings.ToLower(value))
			break
		case dialect.SQLite:
			b.Ident(s.C(field)).WriteString(" REGEXP ")
			if !strings.HasPrefix(value, "(?i)") {
				value = "(?i)" + value
			}
			b.Arg(strings.ToLower(value))
			break
		case dialect.Gremlin:
			break
		}
	})
	return p
}

// filterSearch 全文搜索
// SQL:
func filterSearch(s *sql.Selector, _, _ string) *sql.Predicate {
	p := sql.P()
	p.Append(func(b *sql.Builder) {
		switch s.Builder.Dialect() {

		}
	})

	return nil
}

// filterDatePart 时间戳提取日期 select extract(quarter from timestamp '2018-08-15 12:10:10');
// SQL:
func filterDatePart(s *sql.Selector, datePart, field, value string) *sql.Predicate {
	return nil
}
