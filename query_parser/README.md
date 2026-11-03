# query_parser

Query parser: turns sorting and filtering parameters from HTTP requests into
structured callback events, for the upper layer to assemble SQL `ORDER BY` /
`WHERE` clauses.

English | [简体中文](README.zh-CN.md)

## Features

- **Order-by parsing**: comma-separated strings such as `type,-create_time`
  (`-` descending, `+` explicitly ascending, no prefix ascending)
- **Filter parsing** in two input formats:
  - JSON format: `{"name__exact": "John"}`, either a single object or an array of objects
  - Custom query-string format: `name__exact:John,age__gte:30` (commas separate pairs,
    colons separate field from value, pipes separate multiple values)
- 21 filter operators (inspired by Tortoise ORM / Django field lookups) + 14 date-extraction operators
- Field names are automatically converted to `snake_case`; query-string keys/values are URL-decoded

## Installation

query_parser is an independent Go module (`query_parser/go.mod`) and depends on
`github.com/go-kratos/kratos/v2` (JSON codec) and `github.com/tx7do/go-utils` (case conversion):

```bash
go get github.com/tx7do/go-utils/query_parser
```

```go
import "github.com/tx7do/go-utils/query_parser"
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/query_parser"
)

func main() {
	// Filter: JSON format (single object or array of objects).
	// Note: a JSON object is walked as a Go map, so callback order is
	// not guaranteed to match the written order.
	err := query_parser.ParseFilterJSONString(
		`{"userName__icontains":"Tom","createTime__gte":"2024-01-01"}`,
		func(field, operator, value string) {
			fmt.Println("filter:", field, operator, value)
			// prints (in unspecified order):
			// filter: user_name icontains Tom
			// filter: create_time gte 2024-01-01
		},
	)
	if err != nil {
		panic(err)
	}

	// Filter: custom query-string format
	err = query_parser.ParseFilterQueryString(
		"userName__icontains:Tom,age__gte:30",
		func(field, operator, value string) {
			fmt.Println("filter:", field, operator, value)
		},
	)
	if err != nil {
		panic(err)
	}

	// Order by: comma-separated string; `-` prefix descending,
	// `+` prefix explicitly ascending, no prefix ascending.
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

## Order-By Rules

Ordering is essentially the `ORDER BY` clause of a SQL statement.

The order-by input is a **comma-separated string** (not a JSON array); each
segment is a field name:

| Order            | Example                | Description                                  |
|------------------|------------------------|----------------------------------------------|
| Ascending        | `type`                 | No prefix, ascending by default               |
| Explicit ascend  | `+createTime`          | `+` prefix declares ascending explicitly      |
| Descending       | `-createTime`          | `-` prefix means descending                   |
| Multiple fields  | `type,-createTime,+id` | Comma-separated; callbacks fire in order      |

Parsing functions:

- `ParseOrderByString(orderBy, handler)`: parses one comma-separated order-by string
- `ParseOrderByStrings(orderBys, handler)`: parses multiple order-by fields (one field spec per element, no comma splitting)
- `ParseOrderByField(orderBy, handler)`: parses a single order-by field

Callback signature: `type OrderByHandler func(field string, desc bool)`. Field names
are passed through verbatim (no case or naming-style conversion); mapping them to
database columns is up to the caller.

```go
query_parser.ParseOrderByString("type,-createTime", func(field string, desc bool) {
	// "type"   -> asc
	// "-createTime" -> field == "createTime", desc == true
})
```

## Filter Rules

Filtering essentially converts the user's query conditions into the `WHERE` clause
of a database query, so users can narrow data down to exactly what they need.

A complete filter has three parts:

1. **Field**: the field to query.
2. **Operator**: the kind of comparison to perform.
3. **Value**: the concrete value to query.

A plain equality query only needs the field name; special queries require an operator.

The operator rules borrow from and follow Python ORM conventions such as:

- [Tortoise ORM Filtering][1]
- [Django Field lookups][2]

Filters are passed via query parameters and must follow one of two formats:

- **JSON format**: query conditions expressed as a JSON object.
- **Custom string format**: query conditions expressed as a custom string.

### JSON Format

This format uses 2 delimiters:

- **Double underscore** `__`: separates the field name from the operator; absence of an operator means equality.
- **Dot** `.`: separates the field name from a JSON (embedded) field name.

```text
{field}__{operator} : {value}
{field}.{jsonField}__{operator} : {value}

{{field1}__{operator1} : {value1}, {field2}__{operator2} : {value2}}
[{{field1}__{operator1} : {value1}}, {{field2}__{operator2} : {value2}}]
```

The top-level input may be a single JSON object or an array of objects (the input is
deserialized as `map[string]string` first, then as `[]map[string]string`). Entry
point: `ParseFilterJSONString(query, handler)`; an empty string returns `nil`, invalid
JSON returns an error.

### Custom String Format

This format uses 5 delimiters:

- **Comma** `,`: separates multiple query conditions.
- **Colon** `:`: separates `field+operator` from the value.
- **Double underscore** `__`: separates the field name from the operator.
- **Pipe** `|`: separates multiple values.
- **Dot** `.`: separates the field name from a JSON (embedded) field name.

```text
{field}__{operator} : {value}
{field1}__{operator1} : {value1}, {field2}__{operator2} : {value2}
{field}.{jsonField}__{operator} : {value}
```

Entry point: `ParseFilterQueryString(query, handler)`. Each pair is split on the
colon into two parts; key and value are each URL-decoded (`url.QueryUnescape`). Pairs
with the wrong number of parts or failed decoding are skipped silently (no error).

### Operator Table

Field/operator rules are identical for both formats, so the table below applies to
both. The 21 operators correspond one-to-one with the constants in `filter.go`:

| Lookup      | JSON example                                                   | Query-string example                          | SQL                                                                                                                                                       | Notes                                                                                                           |
|-------------|----------------------------------------------------------------|-----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| not         | `{"name__not" : "tom"}`                                       | `name__not : tom`                            | `WHERE NOT ("name" = "tom")`                                                                                                                              |                                                                                                               |
| in          | `{"name__in" : "[\"tom\", \"jimmy\"]"}`                       | `name__in : tom \| jimmy`                    | `WHERE name IN ("tom", "jimmy")`                                                                                                                          |                                                                                                               |
| not_in      | `{"name__not_in" : "[\"tom\", \"jimmy\"]"}`                   | `name__not_in : tom \| jimmy`                | `WHERE name NOT IN ("tom", "jimmy")`                                                                                                                      |                                                                                                               |
| gte         | `{"create_time__gte" : "2023-10-25"}`                         | `create_time__gte : 2023-10-25`              | `WHERE "create_time" >= "2023-10-25"`                                                                                                                     |                                                                                                               |
| gt          | `{"create_time__gt" : "2023-10-25"}`                          | `create_time__gt : 2023-10-25`               | `WHERE "create_time" > "2023-10-25"`                                                                                                                      |                                                                                                               |
| lte         | `{"create_time__lte" : "2023-10-25"}`                         | `create_time__lte : 2023-10-25`              | `WHERE "create_time" <= "2023-10-25"`                                                                                                                     |                                                                                                               |
| lt          | `{"create_time__lt" : "2023-10-25"}`                          | `create_time__lt : 2023-10-25`               | `WHERE "create_time" < "2023-10-25"`                                                                                                                      |                                                                                                               |
| range       | `{"create_time__range" : "[\"2023-10-25\", \"2024-10-25\"]"}` | `create_time__range : 2023-10-25\|2024-10-25` | `WHERE "create_time" BETWEEN "2023-10-25" AND "2024-10-25"` <br>or<br> `WHERE "create_time" >= "2023-10-25" AND "create_time" <= "2024-10-25"`             | Notes: <br>1. BETWEEN boundary semantics differ between databases.<br>2. The date `2005-01-01` is implicitly `2005-01-01 00:00:00`; two identical dates may match nothing. |
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
| search      |                                                               |                                              |                                                                                                                                                           | Full-text search                                                                                               |

### Date-Extraction Operator Table

These operators extract a part of a date field before comparing; both formats apply
(14 operators, matching the `DatePart*` constants in `filter.go`):

| Lookup       | JSON example                         | Query-string example           | SQL                                               | Notes                    |
|--------------|--------------------------------------|--------------------------------|---------------------------------------------------|--------------------------|
| date         | `{"pub_date__date" : "2023-01-01"}`  | `pub_date__date : 2023-01-01`  | `WHERE DATE(pub_date) = '2023-01-01'`             |                          |
| year         | `{"pub_date__year" : "2023"}`        | `pub_date__year : 2023`        | `WHERE EXTRACT('YEAR' FROM pub_date) = '2023'`    | Which year               |
| iso_year     | `{"pub_date__iso_year" : "2023"}`    | `pub_date__iso_year : 2023`    | `WHERE EXTRACT('ISOYEAR' FROM pub_date) = '2023'` | ISO 8601 year            |
| quarter      | `{"pub_date__quarter" : "1"}`        | `pub_date__quarter : 1`        | `WHERE EXTRACT('QUARTER' FROM pub_date) = '1'`    | Quarter of the year      |
| month        | `{"pub_date__month" : "12"}`         | `pub_date__month : 12`         | `WHERE EXTRACT('MONTH' FROM pub_date) = '12'`     | Month, 1-12              |
| week         | `{"pub_date__week" : "7"}`           | `pub_date__week : 7`           | `WHERE EXTRACT('WEEK' FROM pub_date) = '7'`       | ISO 8601 week number     |
| week_day     | `{"pub_date__week_day" : "1"}`       | `pub_date__week_day : 1`       | ``                                                | Day of week              |
| iso_week_day | `{"pub_date__iso_week_day" : "1"}`   | `pub_date__iso_week_day : 1`   | ``                                                | ISO 8601 day of week     |
| day          | `{"pub_date__day" : "3"}`            | `pub_date__day : 3`            | `WHERE EXTRACT('DAY' FROM pub_date) = '3'`        | Day of month (1-31)      |
| time         | `{"pub_date__time" : "12:59:59"}`    | `pub_date__time : 12:59:59`    | ``                                                | hour:minute:second       |
| hour         | `{"pub_date__hour" : "12"}`          | `pub_date__hour : 12`          | `WHERE EXTRACT('HOUR' FROM pub_date) = '12'`      | Hour (0-23)              |
| minute       | `{"pub_date__minute" : "59"}`        | `pub_date__minute : 59`        | `WHERE EXTRACT('MINUTE' FROM pub_date) = '59'`    | Minute (0-59)            |
| second       | `{"pub_date__second" : "59"}`        | `pub_date__second : 59`        | `WHERE EXTRACT('SECOND' FROM pub_date) = '59'`    | Second (0-59)            |
| microsecond  | `{"pub_date__microsecond" : "500"}`  | `pub_date__microsecond : 500`  | ``                                                | Microsecond              |

## API Reference

### Callback types

```go
type FilterHandler func(field, operator, value string)

type OrderByHandler func(field string, desc bool)
```

### Filter parsing (filter.go)

| Function | Description |
|----------|-------------|
| `ParseFilterJSONString(query string, handler FilterHandler) error` | Parses a JSON filter string. The input may be a single JSON object (`map[string]string`) or an array of objects (`[]map[string]string`). An empty string returns `nil` with no callbacks; invalid JSON returns an error |
| `ParseFilterQueryString(query string, handler FilterHandler) error` | Parses a custom query string. Commas separate pairs, colons separate field from value; key and value are URL-decoded; unparseable pairs are skipped (no error), empty string returns `nil` |
| `ParseFilterField(key, value string, handler FilterHandler)` | Parses a single condition: the key is split on `__` into field name and operator, the field name is converted to `snake_case`, then the handler fires. No callback when key or value is empty |
| `EncodeSpecialCharacters(input string) string` / `DecodeSpecialCharacters(input string) (string, error)` | URL encode/decode (via `url.QueryEscape` / `url.QueryUnescape`) for field names/values that contain separators such as `,` `:` `\|` |

### Order-by parsing (orderby.go)

| Function | Description |
|----------|-------------|
| `ParseOrderByString(orderBy string, handler OrderByHandler) error` | Parses one comma-separated order-by string, firing the handler per field. Empty (or blank) input returns `nil` with no callbacks |
| `ParseOrderByStrings(orderBys []string, handler OrderByHandler) error` | Parses multiple order-by fields: each element is one standalone field spec (no comma splitting); empty elements are skipped |
| `ParseOrderByField(orderBy string, handler OrderByHandler)` | Parses a single order-by field: `-` prefix descending, `+` prefix explicitly ascending, no prefix ascending; blank input is ignored |

### Delimiter helpers

| Function | Delimiter | Description |
|----------|-----------|-------------|
| `SplitJsonFieldAndOperator(field string) []string` | `__` | Splits the JSON filter field name from its operator |
| `SplitQueryFieldAndOperator(field string) []string` | `:` | Splits `field+operator` from the value in a query-string pair |
| `SplitQueryQueries(field string) []string` | `,` | Splits multiple pairs in a query string |
| `SplitQueryValues(field string) []string` | `\|` | Splits multiple values in a query-string pair |
| `SplitJSONField(field string) []string` | `.` | Splits a JSONB field path |

Related delimiter constants: `JSONFilterFieldOperatorDelimiter = "__"`,
`QueryFilterFieldOperatorDelimiter = ":"`, `QueryFilterQueriesDelimiter = ","`,
`QueryFilterValuesDelimiter = "|"` and `JsonFieldDelimiter = "."`.

## Behavior Conventions and Notes

- **Field names are forced to `snake_case`**: `ParseFilterField` (which both filter
  formats ultimately call) converts the field name — the first segment before `__` —
  with `stringcase.ToSnakeCase` before invoking the handler. For example the JSON key
  `"userName__icontains"` yields `field == "user_name"`, `operator == "icontains"`;
  `"na!me"` becomes `"na_me"` (covered by tests). Order-by parsing does **not** apply
  this conversion; field names are passed through verbatim.
- **Query-string keys/values are URL-decoded**: `ParseFilterQueryString` calls
  `DecodeSpecialCharacters` (`url.QueryUnescape`) on the key and the value of every
  pair. If a field or value itself contains `,` `:` `|` etc., pre-encode it with
  `EncodeSpecialCharacters` before assembling the query string. Pairs whose decoding
  fails (invalid escape) are skipped.
- **No operator means equality**: a field name without `__` yields an empty
  `operator` string in the callback.
- **JSON values are strings**: all values in the JSON filter are parsed as strings
  (`map[string]string`). List-style operators (`in` / `not_in` / `range`) carry values
  like `"[\"tom\", \"jimmy\"]"`; secondary parsing is left to the caller.
- **Error semantics differ**: `ParseFilterJSONString` returns an error for invalid
  JSON; `ParseFilterQueryString` skips invalid pairs and always returns `nil`.
- The operator constants (`FilterNot`, `FilterIn`, ...) are untyped string constants
  and can be compared directly against the `operator` value in callbacks.

## References

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
