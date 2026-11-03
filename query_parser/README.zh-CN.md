# query_parser

查询解析器：将 HTTP 请求中的排序与过滤参数解析为结构化的回调事件，供上层拼装 SQL
的 `ORDER BY` / `WHERE` 子句。

[English](README.md) | 简体中文

## 特性

- **排序解析**：`type,-create_time` 形式的排序串（`-` 降序、`+` 显式升序、无前缀升序）
- **过滤解析**，两种输入格式：
  - JSON 格式：`{"name__exact": "John"}`，支持单对象与对象数组
  - 自定义查询串格式：`name__exact:John,age__gte:30`（逗号分隔键值对、冒号分隔字段与值、竖线分隔多值）
- 21 个过滤操作符（借鉴 Tortoise ORM / Django Field lookups）+ 14 个日期提取操作符
- 字段名自动转换为 `snake_case`，查询串的 key/value 自动做 URL 解码

## 安装

query_parser 是独立 Go 模块（`query_parser/go.mod`），依赖
`github.com/go-kratos/kratos/v2`（JSON 编解码）与 `github.com/tx7do/go-utils`（命名转换）：

```bash
go get github.com/tx7do/go-utils/query_parser
```

```go
import "github.com/tx7do/go-utils/query_parser"
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/query_parser"
)

func main() {
	// 过滤：JSON 格式（单对象或对象数组均可）
	// 注意：JSON 对象按 map 遍历，回调顺序不保证与书写顺序一致
	err := query_parser.ParseFilterJSONString(
		`{"userName__icontains":"Tom","createTime__gte":"2024-01-01"}`,
		func(field, operator, value string) {
			fmt.Println("filter:", field, operator, value)
			// 依次输出（顺序不定）：
			// filter: user_name icontains Tom
			// filter: create_time gte 2024-01-01
		},
	)
	if err != nil {
		panic(err)
	}

	// 过滤：自定义查询串格式
	err = query_parser.ParseFilterQueryString(
		"userName__icontains:Tom,age__gte:30",
		func(field, operator, value string) {
			fmt.Println("filter:", field, operator, value)
		},
	)
	if err != nil {
		panic(err)
	}

	// 排序：逗号分隔的字符串，- 前缀降序、+ 前缀显式升序、无前缀升序
	err = query_parser.ParseOrderByString("type,-createTime,+id", func(field string, desc bool) {
		fmt.Println("orderby:", field, desc)
		// orderby: type false
		// orderby: createTime true
		// orderby: id false
	})
	if err != nil {
		panic(err)
	}
}
```

## 排序规则

排序操作本质上是 SQL 里面的 `Order By` 条件。

排序串是**逗号分隔的字符串**（不是 JSON 数组），每一段是一个字段名：

| 排序   | 示例                    | 说明                              |
|------|-----------------------|---------------------------------|
| 升序   | `type`                | 无前缀，默认升序                    |
| 显式升序 | `+createTime`         | `+` 前缀，显式声明升序              |
| 降序   | `-createTime`         | `-` 前缀为降序                    |
| 多字段  | `type,-createTime,+id` | 逗号分隔，按出现顺序依次回调          |

对应解析函数：

- `ParseOrderByString(orderBy, handler)`：解析单个逗号分隔的排序串
- `ParseOrderByStrings(orderBys, handler)`：解析多个排序字段（每个元素一个字段描述，不做逗号切分）
- `ParseOrderByField(orderBy, handler)`：解析单个排序字段

回调签名：`type OrderByHandler func(field string, desc bool)`。字段名按原样回调
（不做大小写或命名风格转换），由调用方决定如何映射到数据库列名。

```go
query_parser.ParseOrderByString("type,-createTime", func(field string, desc bool) {
	// "type"   -> asc
	// "-createTime" -> field == "createTime", desc == true
})
```

## 过滤规则

过滤器操作，其实质上是将用户的查询条件转换为数据库查询语句中的 `WHERE` 子句。通过这
种方式，用户可以根据需要筛选数据，获取更精确的结果。

一个完整的过滤器分为三个部分：

1. **字段名 (Field)**：要查询的字段。
2. **操作符 (Operator)**：用于指定查询的类型。
3. **值 (Value)**：要查询的具体值。

如果只是普通的查询，只需要传递字段名即可，但是如果需要一些特殊的查询，那么就需要加
入操作符了。

过滤器的操作符规则，借鉴并遵循了 Python 中一些 ORM 的规则，比如：

- [Tortoise ORM Filtering][1]
- [Django Field lookups][2]

过滤器通过 query 参数传递，它必须要遵循某一种格式或者说规则。在这里，实现了两种格式：

- **JSON 格式**：使用 JSON 对象来表示查询条件。
- **自定义字符串格式**：使用自定义的字符串来表示查询条件。

### JSON 格式

在这套规则里面，有 2 个分隔符：

- **双下划线** `__`：用于分隔字段名和操作符，如果没有操作符则视作等于操作。
- **点号** `.`：用于分隔字段名和 JSON 字段名。

```text
{字段名}__{操作符} : {查询值}
{字段名}.{JSON字段名}__{操作符} : {查询值}

{{字段名1}__{操作符1} : {查询值1}, {字段名2}__{操作符2} : {查询值2}}
[{{字段名1}__{操作符1} : {查询值1}}, {{字段名2}__{操作符2} : {查询值2}}]
```

顶层输入既可以是单个 JSON 对象，也可以是对象数组（分别按 `map[string]string` 与
`[]map[string]string` 两轮尝试反序列化）。解析入口：
`ParseFilterJSONString(query, handler)`；空串直接返回 `nil`，非法 JSON 返回错误。

### 自定义字符串格式

在这套规则里面，有 5 个分隔符：

- **逗号** `,`：用于分隔多个查询条件。
- **冒号** `:`：用于分隔 `字段名+操作符` 和查询值。
- **双下划线** `__`：用于分隔字段名和操作符。
- **竖线** `|`：用于分隔多个查询值。
- **点号** `.`：用于分隔字段名和 JSON 字段名。

```text
{字段名}__{操作符} : {查询值}
{字段名1}__{操作符1} : {查询值1}, {字段名2}__{操作符2} : {查询值2}
{字段名}.{JSON字段名}__{操作符} : {查询值}
```

解析入口：`ParseFilterQueryString(query, handler)`。每一对键值先按冒号拆成两段，然后
对 key 和 value 各做一次 URL 解码（`url.QueryUnescape`），解码失败或段数不对的键值对
直接跳过（不报错）。

### 操作符表

两种格式的字段名/操作符规则一致，下表对两种格式均适用。21 个操作符与源码
`filter.go` 的常量定义一一对应：

| 查找类型        | JSON 示例                                                      | 查询串示例                                       | SQL                                                                                                                                                       | 备注                                                                                                            |
|-------------|---------------------------------------------------------------|----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| not         | `{"name__not" : "tom"}`                                       | `name__not : tom`                            | `WHERE NOT ("name" = "tom")`                                                                                                                              |                                                                                                               |
| in          | `{"name__in" : "[\"tom\", \"jimmy\"]"}`                       | `name__in : tom \| jimmy`                    | `WHERE name IN ("tom", "jimmy")`                                                                                                                          |                                                                                                               |
| not_in      | `{"name__not_in" : "[\"tom\", \"jimmy\"]"}`                   | `name__not_in : tom \| jimmy`                | `WHERE name NOT IN ("tom", "jimmy")`                                                                                                                      |                                                                                                               |
| gte         | `{"create_time__gte" : "2023-10-25"}`                         | `create_time__gte : 2023-10-25`              | `WHERE "create_time" >= "2023-10-25"`                                                                                                                     |                                                                                                               |
| gt          | `{"create_time__gt" : "2023-10-25"}`                          | `create_time__gt : 2023-10-25`               | `WHERE "create_time" > "2023-10-25"`                                                                                                                      |                                                                                                               |
| lte         | `{"create_time__lte" : "2023-10-25"}`                         | `create_time__lte : 2023-10-25`              | `WHERE "create_time" <= "2023-10-25"`                                                                                                                     |                                                                                                               |
| lt          | `{"create_time__lt" : "2023-10-25"}`                          | `create_time__lt : 2023-10-25`               | `WHERE "create_time" < "2023-10-25"`                                                                                                                      |                                                                                                               |
| range       | `{"create_time__range" : "[\"2023-10-25\", \"2024-10-25\"]"}` | `create_time__range : 2023-10-25\|2024-10-25` | `WHERE "create_time" BETWEEN "2023-10-25" AND "2024-10-25"` <br>或<br> `WHERE "create_time" >= "2023-10-25" AND "create_time" <= "2024-10-25"`             | 需要注意的是: <br>1. 有些数据库的 BETWEEN 实现的开闭区间可能不一样。<br>2. 日期 `2005-01-01` 会被隐式转换为 `2005-01-01 00:00:00`，两个日期一致就会导致查询不到数据。 |
| isnull      | `{"name__isnull" : "True"}`                                   | `name__isnull : True`                        | `WHERE name IS NULL`                                                                                                                                      |                                                                                                               |
| not_isnull  | `{"name__not_isnull" : "False"}`                              | `name__not_isnull : False`                   | `WHERE name IS NOT NULL`                                                                                                                                  |                                                                                                               |
| contains    | `{"name__contains" : "L"}`                                    | `name__contains : L`                         | `WHERE name LIKE '%L%';`                                                                                                                                  |                                                                                                               |
| icontains   | `{"name__icontains" : "L"}`                                   | `name__icontains : L`                        | `WHERE name ILIKE '%L%';`                                                                                                                                 |                                                                                                               |
| startswith  | `{"name__startswith" : "La"}`                                 | `name__startswith : La`                      | `WHERE name LIKE 'La%';`                                                                                                                                  |                                                                                                               |
| istartswith | `{"name__istartswith" : "La"}`                                | `name__istartswith : La`                     | `WHERE name ILIKE 'La%';`                                                                                                                                 |                                                                                                               |
| endswith    | `{"name__endswith" : "a"}`                                    | `name__endswith : a`                         | `WHERE name LIKE '%a';`                                                                                                                                   |                                                                                                               |
| iendswith   | `{"name__iendswith" : "a"}`                                   | `name__iendswith : a`                        | `WHERE name ILIKE '%a';`                                                                                                                                  |                                                                                                               |
| exact       | `{"name__exact" : "a"}`                                       | `name__exact : a`                            | `WHERE name LIKE 'a';`                                                                                                                                    |                                                                                                               |
| iexact      | `{"name__iexact" : "a"}`                                      | `name__iexact : a`                           | `WHERE name ILIKE 'a';`                                                                                                                                   |                                                                                                               |
| regex       | `{"title__regex" : "^(An?\|The) +"}`                          | `title__regex : ^(An?\|The) +`               | MySQL: `WHERE title REGEXP BINARY '^(An?\|The) +'`  <br> Oracle: `WHERE REGEXP_LIKE(title, '^(An?\|The) +', 'c');`  <br> PostgreSQL: `WHERE title ~ '^(An?\|The) +';`  <br> SQLite: `WHERE title REGEXP '^(An?\|The) +';` |                                                                                                               |
| iregex      | `{"title__iregex" : "^(an?\|the) +"}`                         | `title__iregex : ^(an?\|the) +`              | MySQL: `WHERE title REGEXP '^(an?\|the) +'`  <br> Oracle: `WHERE REGEXP_LIKE(title, '^(an?\|the) +', 'i');`  <br> PostgreSQL: `WHERE title ~* '^(an?\|the) +';`  <br> SQLite: `WHERE title REGEXP '(?i)^(an?\|the) +';`   |                                                                                                               |
| search      |                                                               |                                              |                                                                                                                                                           | 全文搜索                                                                                                           |

### 日期提取操作符表

以下操作符将日期字段按指定部分提取后比较，两种格式同样适用（14 个，与源码
`filter.go` 的 `DatePart*` 常量一致）：

| 查找类型         | JSON 示例                              | 查询串示例                           | SQL                                               | 备注                   |
|--------------|--------------------------------------|---------------------------------|---------------------------------------------------|----------------------|
| date         | `{"pub_date__date" : "2023-01-01"}`  | `pub_date__date : 2023-01-01`   | `WHERE DATE(pub_date) = '2023-01-01'`             |                      |
| year         | `{"pub_date__year" : "2023"}`        | `pub_date__year : 2023`         | `WHERE EXTRACT('YEAR' FROM pub_date) = '2023'`    | 哪一年                  |
| iso_year     | `{"pub_date__iso_year" : "2023"}`    | `pub_date__iso_year : 2023`     | `WHERE EXTRACT('ISOYEAR' FROM pub_date) = '2023'` | ISO 8601 年           |
| quarter      | `{"pub_date__quarter" : "1"}`        | `pub_date__quarter : 1`         | `WHERE EXTRACT('QUARTER' FROM pub_date) = '1'`    | 一年中的季度                |
| month        | `{"pub_date__month" : "12"}`         | `pub_date__month : 12`          | `WHERE EXTRACT('MONTH' FROM pub_date) = '12'`     | 月份，1-12              |
| week         | `{"pub_date__week" : "7"}`           | `pub_date__week : 7`            | `WHERE EXTRACT('WEEK' FROM pub_date) = '7'`       | ISO 8601 周编号 一年中的周数    |
| week_day     | `{"pub_date__week_day" : "1"}`       | `pub_date__week_day : 1`        | ``                                                | 星期几                  |
| iso_week_day | `{"pub_date__iso_week_day" : "1"}`   | `pub_date__iso_week_day : 1`    | ``                                                | ISO 8601 星期几          |
| day          | `{"pub_date__day" : "3"}`            | `pub_date__day : 3`             | `WHERE EXTRACT('DAY' FROM pub_date) = '3'`        | 该月的某天(1-31)          |
| time         | `{"pub_date__time" : "12:59:59"}`    | `pub_date__time : 12:59:59`     | ``                                                | 小时:分钟:秒              |
| hour         | `{"pub_date__hour" : "12"}`          | `pub_date__hour : 12`           | `WHERE EXTRACT('HOUR' FROM pub_date) = '12'`      | 小时(0-23)             |
| minute       | `{"pub_date__minute" : "59"}`        | `pub_date__minute : 59`         | `WHERE EXTRACT('MINUTE' FROM pub_date) = '59'`    | 分钟 (0-59)            |
| second       | `{"pub_date__second" : "59"}`        | `pub_date__second : 59`         | `WHERE EXTRACT('SECOND' FROM pub_date) = '59'`    | 秒 (0-59)             |
| microsecond  | `{"pub_date__microsecond" : "500"}`  | `pub_date__microsecond : 500`   | ``                                                | 微秒                   |

## API 参考

### 回调类型

```go
type FilterHandler func(field, operator, value string)

type OrderByHandler func(field string, desc bool)
```

### 过滤解析（filter.go）

| 函数 | 说明 |
|------|------|
| `ParseFilterJSONString(query string, handler FilterHandler) error` | 解析 JSON 过滤串。入参可以是单个 JSON 对象（`map[string]string`）或对象数组（`[]map[string]string`）。空串返回 `nil` 且不回调；非法 JSON 返回错误 |
| `ParseFilterQueryString(query string, handler FilterHandler) error` | 解析自定义查询串。逗号分隔键值对、冒号分隔字段与值；key/value 均做 URL 解码；无法解析的键值对跳过（不报错），空串返回 `nil` |
| `ParseFilterField(key, value string, handler FilterHandler)` | 解析单个过滤条件：key 按 `__` 拆分字段名与操作符，字段名转 `snake_case` 后回调。key 或 value 为空时不回调 |
| `EncodeSpecialCharacters(input string) string` / `DecodeSpecialCharacters(input string) (string, error)` | 对字段名/值做 URL 编解码（`url.QueryEscape` / `url.QueryUnescape`），用于在查询串中传递含 `,` `:` `\|` 等分隔符的内容 |

### 排序解析（orderby.go）

| 函数 | 说明 |
|------|------|
| `ParseOrderByString(orderBy string, handler OrderByHandler) error` | 解析逗号分隔的排序串，逐字段回调。空串（或全空白）返回 `nil` 且不回调 |
| `ParseOrderByStrings(orderBys []string, handler OrderByHandler) error` | 解析多个排序字段：每个元素是一个独立的字段描述（不做逗号切分），逐个回调；空字符串元素跳过 |
| `ParseOrderByField(orderBy string, handler OrderByHandler)` | 解析单个排序字段：`-` 前缀降序，`+` 前缀显式升序，无前缀升序；空白串不回调 |

### 分隔符辅助函数

| 函数 | 分隔符 | 说明 |
|------|------|------|
| `SplitJsonFieldAndOperator(field string) []string` | `__` | 拆分 JSON 过滤器的字段名与操作符 |
| `SplitQueryFieldAndOperator(field string) []string` | `:` | 拆分查询串过滤器中"字段名+操作符"与值 |
| `SplitQueryQueries(field string) []string` | `,` | 拆分查询串中的多个键值对 |
| `SplitQueryValues(field string) []string` | `\|` | 拆分查询串中的多个值 |
| `SplitJSONField(field string) []string` | `.` | 拆分 JSONB 字段路径 |

相关分隔符常量：`JSONFilterFieldOperatorDelimiter = "__"`、`QueryFilterFieldOperatorDelimiter = ":"`、
`QueryFilterQueriesDelimiter = ","`、`QueryFilterValuesDelimiter = "|"`、`JsonFieldDelimiter = "."`。

## 行为约定与注意事项

- **字段名强制转 `snake_case`**：`ParseFilterField`（两种过滤格式最终都经由它回调）会把
  字段名（`__` 之前的首段）用 `stringcase.ToSnakeCase` 转换后再交给 handler。例如 JSON key
  `"userName__icontains"` 回调出 `field == "user_name"`、`operator == "icontains"`；
  含特殊字符的 `"na!me"` 转为 `"na_me"`（有测试佐证）。排序解析**不做**该转换，字段名按
  原样回调。
- **查询串 key/value 会做 URL 解码**：`ParseFilterQueryString` 对每对键值的 key 与 value
  调用 `DecodeSpecialCharacters`（`url.QueryUnescape`）。若字段或值本身含 `,` `:` `|`
  等分隔符，调用方可先用 `EncodeSpecialCharacters` 编码后再拼进查询串。编码失败（非法
  转义）的键值对被跳过。
- **无操作符视作等于**：字段名后没有 `__` 时，回调的 `operator` 为空字符串 `""`。
- **JSON 格式的值是字符串**：JSON 过滤串中所有值都按字符串解析（`map[string]string`），
  列表类操作符（`in` / `not_in` / `range`）的值是形如 `"[\"tom\", \"jimmy\"]"` 的字符串，
  由调用方自行二次解析。
- **错误语义不同**：`ParseFilterJSONString` 遇到非法 JSON 返回错误；`ParseFilterQueryString`
  容错跳过非法键值对，始终返回 `nil`。
- 操作符常量（`FilterNot`、`FilterIn` 等）为无类型字符串常量，可直接与回调中的 `operator`
  比较。

## 参考资料

- [Tortoise ORM Filtering][1]
- [Django Field lookups][2]
- [PostgreSQL Date/Time Functions and Operators][3]
- [PostgreSQL Regular Expressions][4]
- [PostgreSQL Date/Time Types][5]

[1]: https://tortoise.github.io/query.html#filtering

[2]: https://docs.djangoproject.com/en/4.2/ref/models/querysets/#field-lookups

[3]: https://www.postgresql.org/docs/current/functions-datetime.html

[4]: https://www.postgresql.org/docs/current/functions-matching.html#FUNCTIONS-REGEXP-TABLE

[5]: https://www.postgresql.org/docs/current/datatype-datetime.html
