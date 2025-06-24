# 查询解析器

## 排序规则

排序操作本质上是`SQL`里面的`Order By`条件。

| 序列 | 示例                 | 备注           |
|----|--------------------|--------------|
| 升序 | `["type"]`         |              |
| 降序 | `["-create_time"]` | 字段名前加`-`是为降序 |

## 过滤规则

过滤器操作，其实质上是将用户的查询条件转换为数据库查询语句中的`WHERE`子句。通过这种方式，用户可以根据需要筛选数据，获取更精确的结果。

一个完整的过滤器分为三个部分：

1. **字段名 (Field)**：要查询的字段。
2. **操作符 (Operator)**：用于指定查询的类型。
3. **值 (Value)**：要查询的具体值。

如果只是普通的查询，只需要传递`字段名`即可，但是如果需要一些特殊的查询，那么就需要加入`操作符`了。

过滤器的`操作符`规则，我借鉴并遵循了Python中一些ORM的规则，比如：

- [Tortoise ORM Filtering][1]。
- [Django Field lookups][2]

过滤器通过query参数传递，它必须要遵循某一种格式或者说规则。在这里，我实现了两种格式：

- **JSON格式**：使用`JSON`对象来表示查询条件。
- **自定义字符串格式**：使用`自定义的字符串`来表示查询条件。

### JSON格式

在这套规则里面，我们有2个分隔符：

- **双下划线** `__`：用于分隔`字段名`和`操作符`，如果没有操作符则视作等于操作。
- **点号** `.`：用于分隔`字段名`和`JSON字段名`。

```text
{字段名}__{操作符} : {查询值}
{字段名}.{JSON字段名}__{操作符} : {查询值}

{{字段名1}__{操作符1} : {查询值1}, {字段名2}__{操作符2} : {查询值2}}
[{{字段名1}__{操作符1} : {查询值1}}, {{字段名1}__{操作符2} : {查询值2}}]
```

| 查找类型        | 示例                                                            | SQL                                                                                                                                                                                                                       | 备注                                                                                                            |
|-------------|---------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| not         | `{"name__not" : "tom"}`                                       | `WHERE NOT ("name" = "tom")`                                                                                                                                                                                              |                                                                                                               |
| in          | `{"name__in" : "[\"tom\", \"jimmy\"]"}`                       | `WHERE name IN ("tom", "jimmy")`                                                                                                                                                                                          |                                                                                                               |
| not_in      | `{"name__not_in" : "[\"tom\", \"jimmy\"]"}`                   | `WHERE name NOT IN ("tom", "jimmy")`                                                                                                                                                                                      |                                                                                                               |
| gte         | `{"create_time__gte" : "2023-10-25"}`                         | `WHERE "create_time" >= "2023-10-25"`                                                                                                                                                                                     |                                                                                                               |
| gt          | `{"create_time__gt" : "2023-10-25"}`                          | `WHERE "create_time" > "2023-10-25"`                                                                                                                                                                                      |                                                                                                               |
| lte         | `{"create_time__lte" : "2023-10-25"}`                         | `WHERE "create_time" <= "2023-10-25"`                                                                                                                                                                                     |                                                                                                               |
| lt          | `{"create_time__lt" : "2023-10-25"}`                          | `WHERE "create_time" < "2023-10-25"`                                                                                                                                                                                      |                                                                                                               |
| range       | `{"create_time__range" : "[\"2023-10-25\", \"2024-10-25\"]"}` | `WHERE "create_time" BETWEEN "2023-10-25" AND "2024-10-25"` <br>或<br> `WHERE "create_time" >= "2023-10-25" AND "create_time" <= "2024-10-25"`                                                                             | 需要注意的是: <br>1. 有些数据库的BETWEEN实现的开闭区间可能不一样。<br>2. 日期`2005-01-01`会被隐式转换为：`2005-01-01 00:00:00`，两个日期一致就会导致查询不到数据。 |
| isnull      | `{"name__isnull" : "True"}`                                   | `WHERE name IS NULL`                                                                                                                                                                                                      |                                                                                                               |
| not_isnull  | `{"name__not_isnull" : "False"}`                              | `WHERE name IS NOT NULL`                                                                                                                                                                                                  |                                                                                                               |
| contains    | `{"name__contains" : "L"}`                                    | `WHERE name LIKE '%L%';`                                                                                                                                                                                                  |                                                                                                               |
| icontains   | `{"name__icontains" : "L"}`                                   | `WHERE name ILIKE '%L%';`                                                                                                                                                                                                 |                                                                                                               |
| startswith  | `{"name__startswith" : "La"}`                                 | `WHERE name LIKE 'La%';`                                                                                                                                                                                                  |                                                                                                               |
| istartswith | `{"name__istartswith" : "La"}`                                | `WHERE name ILIKE 'La%';`                                                                                                                                                                                                 |                                                                                                               |
| endswith    | `{"name__endswith" : "a"}`                                    | `WHERE name LIKE '%a';`                                                                                                                                                                                                   |                                                                                                               |
| iendswith   | `{"name__iendswith" : "a"}`                                   | `WHERE name ILIKE '%a';`                                                                                                                                                                                                  |                                                                                                               |
| exact       | `{"name__exact" : "a"}`                                       | `WHERE name LIKE 'a';`                                                                                                                                                                                                    |                                                                                                               |
| iexact      | `{"name__iexact" : "a"}`                                      | `WHERE name ILIKE 'a';`                                                                                                                                                                                                   |                                                                                                               |
| regex       | `{"title__regex" : "^(An?\|The) +"}`                          | MySQL: `WHERE title REGEXP BINARY '^(An?\|The) +'`  <br> Oracle: `WHERE REGEXP_LIKE(title, '^(An?\|The) +', 'c');`  <br> PostgreSQL: `WHERE title ~ '^(An?\|The) +';`  <br> SQLite: `WHERE title REGEXP '^(An?\|The) +';` |                                                                                                               |
| iregex      | `{"title__iregex" : "^(an?\|the) +"}`                         | MySQL: `WHERE title REGEXP '^(an?\|the) +'`  <br> Oracle: `WHERE REGEXP_LIKE(title, '^(an?\|the) +', 'i');`  <br> PostgreSQL: `WHERE title ~* '^(an?\|the) +';`  <br> SQLite: `WHERE title REGEXP '(?i)^(an?\|the) +';`   |                                                                                                               |
| search      |                                                               |                                                                                                                                                                                                                           |                                                                                                               |

以及将日期提取出来的查找类型：

| 查找类型         | 示例                                   | SQL                                               | 备注                   |
|--------------|--------------------------------------|---------------------------------------------------|----------------------|
| date         | `{"pub_date__date" : "2023-01-01"}`  | `WHERE DATE(pub_date) = '2023-01-01'`             |                      |
| year         | `{"pub_date__year" : "2023"}`        | `WHERE EXTRACT('YEAR' FROM pub_date) = '2023'`    | 哪一年                  |
| iso_year     | `{"pub_date__iso_year" : "2023"}`    | `WHERE EXTRACT('ISOYEAR' FROM pub_date) = '2023'` | ISO 8601 一年中的周数      |
| month        | `{"pub_date__month" : "12"}`         | `WHERE EXTRACT('MONTH' FROM pub_date) = '12'`     | 月份，1-12              |
| day          | `{"pub_date__day" : "3"}`            | `WHERE EXTRACT('DAY' FROM pub_date) = '3'`        | 该月的某天(1-31)          |
| week         | `{"pub_date__week" : "7"}`           | `WHERE EXTRACT('WEEK' FROM pub_date) = '7'`       | ISO 8601 周编号 一年中的周数	 |
| week_day     | `{"pub_date__week_day" : "tom"}`     | ``                                                | 星期几                  |
| iso_week_day | `{"pub_date__iso_week_day" : "tom"}` | ``                                                |                      |
| quarter      | `{"pub_date__quarter" : "1"}`        | `WHERE EXTRACT('QUARTER' FROM pub_date) = '1'`    | 一年中的季度	              |
| time         | `{"pub_date__time" : "12:59:59"}`    | ``                                                |                      |
| hour         | `{"pub_date__hour" : "12"}`          | `WHERE EXTRACT('HOUR' FROM pub_date) = '12'`      | 小时(0-23)             |
| minute       | `{"pub_date__minute" : "59"}`        | `WHERE EXTRACT('MINUTE' FROM pub_date) = '59'`    | 分钟 (0-59)            |
| second       | `{"pub_date__second" : "59"}`        | `WHERE EXTRACT('SECOND' FROM pub_date) = '59'`    | 秒 (0-59)             |

### 自定义字符串格式

在这套规则里面，我们有5个分隔符：

- **逗号** `,`：用于分隔多个`查询条件`，如果没有操作符则视作等于操作。
- **冒号** `:`：用于分隔`字段名+操作符` 和 `查询值`。
- **双下划线** `__`：用于分隔`字段名`和`操作符`。
- **竖线** `|`：用于分隔多个的`查询值`。
- **点号** `.`：用于分隔`字段名`和`JSON字段名`。

```text
{字段名}__{操作符} : {查询值}
{字段名1}__{操作符1} : {查询值1}, {字段名2}__{操作符2} : {查询值2}
{字段名}.{JSON字段名}__{操作符} : {查询值}
```

| 查找类型        | 示例                                             | SQL                                                                                                                                                                                                                       | 备注                                                                                                            |
|-------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| not         | `name__not : tom`                              | `WHERE NOT ("name" = "tom")`                                                                                                                                                                                              |                                                                                                               |
| in          | `name__in" : tom \| jimm`                      | `WHERE name IN ("tom", "jimmy")`                                                                                                                                                                                          |                                                                                                               |
| not_in      | `name__not_in : tom \| jimm`                   | `WHERE name NOT IN ("tom", "jimmy")`                                                                                                                                                                                      |                                                                                                               |
| gte         | `create_time__gte : 2023-10-25`                | `WHERE "create_time" >= "2023-10-25"`                                                                                                                                                                                     |                                                                                                               |
| gt          | `create_time__gt : 2023-10-25`                 | `WHERE "create_time" > "2023-10-25"`                                                                                                                                                                                      |                                                                                                               |
| lte         | `create_time__lte : 2023-10-25`                | `WHERE "create_time" <= "2023-10-25"`                                                                                                                                                                                     |                                                                                                               |
| lt          | `create_time__lt : 2023-10-25`                 | `WHERE "create_time" < "2023-10-25"`                                                                                                                                                                                      |                                                                                                               |
| range       | `create_time__range : 2023-10-25\| 2024-10-25` | `WHERE "create_time" BETWEEN "2023-10-25" AND "2024-10-25"` <br>或<br> `WHERE "create_time" >= "2023-10-25" AND "create_time" <= "2024-10-25"`                                                                             | 需要注意的是: <br>1. 有些数据库的BETWEEN实现的开闭区间可能不一样。<br>2. 日期`2005-01-01`会被隐式转换为：`2005-01-01 00:00:00`，两个日期一致就会导致查询不到数据。 |
| isnull      | `name__isnull : True`                          | `WHERE name IS NULL`                                                                                                                                                                                                      |                                                                                                               |
| not_isnull  | `name__not_isnull : False`                     | `WHERE name IS NOT NULL`                                                                                                                                                                                                  |                                                                                                               |
| contains    | `name__contains : L`                           | `WHERE name LIKE '%L%';`                                                                                                                                                                                                  |                                                                                                               |
| icontains   | `name__icontains : L`                          | `WHERE name ILIKE '%L%';`                                                                                                                                                                                                 |                                                                                                               |
| startswith  | `name__startswith : La`                        | `WHERE name LIKE 'La%';`                                                                                                                                                                                                  |                                                                                                               |
| istartswith | `name__istartswith : La`                       | `WHERE name ILIKE 'La%';`                                                                                                                                                                                                 |                                                                                                               |
| endswith    | `name__endswith : a`                           | `WHERE name LIKE '%a';`                                                                                                                                                                                                   |                                                                                                               |
| iendswith   | `name__iendswith : a`                          | `WHERE name ILIKE '%a';`                                                                                                                                                                                                  |                                                                                                               |
| exact       | `name__exact : a`                              | `WHERE name LIKE 'a';`                                                                                                                                                                                                    |                                                                                                               |
| iexact      | `name__iexact : a`                             | `WHERE name ILIKE 'a';`                                                                                                                                                                                                   |                                                                                                               |
| regex       | `title__regex : ^(An?\|The) +`                 | MySQL: `WHERE title REGEXP BINARY '^(An?\|The) +'`  <br> Oracle: `WHERE REGEXP_LIKE(title, '^(An?\|The) +', 'c');`  <br> PostgreSQL: `WHERE title ~ '^(An?\|The) +';`  <br> SQLite: `WHERE title REGEXP '^(An?\|The) +';` |                                                                                                               |
| iregex      | `title__iregex : ^(an?\|the) +`                | MySQL: `WHERE title REGEXP '^(an?\|the) +'`  <br> Oracle: `WHERE REGEXP_LIKE(title, '^(an?\|the) +', 'i');`  <br> PostgreSQL: `WHERE title ~* '^(an?\|the) +';`  <br> SQLite: `WHERE title REGEXP '(?i)^(an?\|the) +';`   |                                                                                                               |
| search      |                                                |                                                                                                                                                                                                                           |                                                                                                               |

以及将日期提取出来的查找类型：

| 查找类型         | 示例                             | SQL                                               | 备注                   |
|--------------|--------------------------------|---------------------------------------------------|----------------------|
| date         | `pub_date__date : 2023-01-01`  | `WHERE DATE(pub_date) = '2023-01-01'`             |                      |
| year         | `pub_date__year : 2023`        | `WHERE EXTRACT('YEAR' FROM pub_date) = '2023'`    | 哪一年                  |
| iso_year     | `pub_date__iso_year : 2023`    | `WHERE EXTRACT('ISOYEAR' FROM pub_date) = '2023'` | ISO 8601 一年中的周数      |
| month        | `pub_date__month : 12`         | `WHERE EXTRACT('MONTH' FROM pub_date) = '12'`     | 月份，1-12              |
| day          | `pub_date__day : 3`            | `WHERE EXTRACT('DAY' FROM pub_date) = '3'`        | 该月的某天(1-31)          |
| week         | `pub_date__week : 7`           | `WHERE EXTRACT('WEEK' FROM pub_date) = '7'`       | ISO 8601 周编号 一年中的周数	 |
| week_day     | `pub_date__week_day : tom`     | ``                                                | 星期几                  |
| iso_week_day | `pub_date__iso_week_day : tom` | ``                                                |                      |
| quarter      | `pub_date__quarter : 1`        | `WHERE EXTRACT('QUARTER' FROM pub_date) = '1'`    | 一年中的季度	              |
| time         | `pub_date__time : 12:59:59`    | ``                                                |                      |
| hour         | `pub_date__hour : 12`          | `WHERE EXTRACT('HOUR' FROM pub_date) = '12'`      | 小时(0-23)             |
| minute       | `pub_date__minute : 59`        | `WHERE EXTRACT('MINUTE' FROM pub_date) = '59'`    | 分钟 (0-59)            |
| second       | `pub_date__second : 59`        | `WHERE EXTRACT('SECOND' FROM pub_date) = '59'`    | 秒 (0-59)             |

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
