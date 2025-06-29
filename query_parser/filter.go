package query_parser

import (
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
	_ "github.com/go-kratos/kratos/v2/encoding/json"

	"github.com/tx7do/go-utils/stringcase"
)

type FilterOperator string

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

const (
	DatePartDate        = "date"         // 日期
	DatePartYear        = "year"         // 年
	DatePartISOYear     = "iso_year"     // ISO8601 一年中的周数
	DatePartQuarter     = "quarter"      // 季度
	DatePartMonth       = "month"        // 月
	DatePartWeek        = "week"         // ISO8601 周编号 一年中的周数
	DatePartWeekDay     = "week_day"     // 星期几
	DatePartISOWeekDay  = "iso_week_day" // ISO8601 星期几
	DatePartDay         = "day"          // 日
	DatePartTime        = "time"         // 小时：分钟：秒
	DatePartHour        = "hour"         // 小时
	DatePartMinute      = "minute"       // 分钟
	DatePartSecond      = "second"       // 秒
	DatePartMicrosecond = "microsecond"  // 微秒
)

const (
	JSONFilterFieldOperatorDelimiter = "__" // JSON过滤器 - 字段名和操作符的分隔符

	QueryFilterFieldOperatorDelimiter = ":" // 自定义查询字符串过滤器 - 字段名和操作符的分隔符
	QueryFilterQueriesDelimiter       = "," // 自定义查询字符串过滤器 - 多个键值对的分隔符
	QueryFilterValuesDelimiter        = "|" // 自定义查询字符串过滤器 - 多个值的分隔符

	JsonFieldDelimiter = "." // JSON字段分隔符
)

type FilterHandler func(field, operator, value string)

// ParseFilterJSONString 解析过滤条件的JSON字符串，调用处理函数
func ParseFilterJSONString(query string, handler FilterHandler) error {
	if query == "" {
		return nil
	}

	codec := encoding.GetCodec("json")

	var err error
	queryMap := make(map[string]string)
	if err = codec.Unmarshal([]byte(query), &queryMap); err == nil {
		for k, v := range queryMap {
			ParseFilterField(k, v, handler)
		}
		return nil
	}

	var queryMapArray []map[string]string
	if err = codec.Unmarshal([]byte(query), &queryMapArray); err == nil {
		for _, item := range queryMapArray {
			for k, v := range item {
				ParseFilterField(k, v, handler)
			}
		}
		return nil
	}

	return err
}

// ParseFilterQueryString 解析过滤条件的查询字符串，调用处理函数
func ParseFilterQueryString(query string, handler FilterHandler) error {
	if query == "" {
		return nil // 如果查询字符串为空，直接返回
	}

	// 按逗号分割查询字符串，得到多个键值对
	pairs := SplitQueryQueries(query)
	for _, pair := range pairs {
		// 按冒号分割键值对，提取字段名和值
		parts := SplitQueryFieldAndOperator(pair)
		if len(parts) != 2 {
			continue // 跳过无效的键值对
		}

		// 解码字段名
		key, err := DecodeSpecialCharacters(strings.TrimSpace(parts[0]))
		if err != nil {
			continue // 跳过解码失败的键值对
		}

		// 解码字段值
		value, err := DecodeSpecialCharacters(strings.TrimSpace(parts[1]))
		if err != nil {
			continue // 跳过解码失败的键值对
		}

		// 调用 ParseFilterField 解析字段和操作符
		ParseFilterField(key, value, handler)
	}

	return nil
}

// ParseFilterField 解析过滤条件字符串，调用处理函数
func ParseFilterField(key, value string, handler FilterHandler) {
	if key == "" || value == "" {
		return // 没有过滤条件
	}

	// 处理字段和操作符
	parts := SplitJsonFieldAndOperator(key)
	if len(parts) < 1 {
		return // 无效的字段格式
	}

	field := strings.TrimSpace(parts[0])
	if field == "" {
		return
	}
	field = stringcase.ToSnakeCase(parts[0])

	op := ""
	if len(parts) > 1 {
		op = parts[1]
	}

	handler(field, op, value)
}

// SplitJsonFieldAndOperator JSON过滤器 - 分割“字段名”和“操作符”
func SplitJsonFieldAndOperator(field string) []string {
	return strings.Split(field, JSONFilterFieldOperatorDelimiter)
}

// SplitQueryFieldAndOperator 自定义查询字符串过滤器 - 分割“字段名”和“操作符”
func SplitQueryFieldAndOperator(field string) []string {
	return strings.Split(field, QueryFilterFieldOperatorDelimiter)
}

// SplitQueryQueries 自定义查询字符串过滤器 - 分割多个键值对
func SplitQueryQueries(field string) []string {
	return strings.Split(field, QueryFilterQueriesDelimiter)
}

// SplitQueryValues 自定义查询字符串过滤器 - 分割多个值
func SplitQueryValues(field string) []string {
	return strings.Split(field, QueryFilterValuesDelimiter)
}

// SplitJSONField 将JSONB字段字符串按分隔符分割成多个字段
func SplitJSONField(field string) []string {
	return strings.Split(field, JsonFieldDelimiter)
}
