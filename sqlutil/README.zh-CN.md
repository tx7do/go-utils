# sqlutil

PostgreSQL 词法级的 SQL 审计工具，纯标准库实现（零第三方依赖）。

[English](README.md) | 简体中文

## 特性

- **`MaskSQL`**：字面量级脱敏——字符串字面量（`'…'`、`E'…'`、`$tag$…$tag$` 美元引用）
  与数值字面量整体替换为 `***`，保留 SQL 结构（关键字、双引号标识符、`$n` 占位符、
  注释、`::` 类型转换）。按 PG 词法规则实现：单引号串内 `\'` 与 `''` 均按转义处理
  （宁多掩不漏掩）、美元引用可跨行、块注释支持嵌套、注释与标识符中的引号不误判。
- **`ExtractTables`**：从（已脱敏的）SQL 文本提取被访问的表名，覆盖 FROM /
  INSERT INTO / UPDATE / DELETE FROM / JOIN 及 FROM 的逗号多表，跳过子查询括号，
  schema 限定名保留全名，按首个出现顺序去重。字符串与注释不拆词。
- **`MaskingRules`**：脱敏策略标识常量（`"sql_literals:v1"`），可随审计记录落库留痕。

## 安装

sqlutil 是独立 Go 模块（`sqlutil/go.mod`），无第三方运行时依赖：

```bash
go get github.com/tx7do/go-utils/sqlutil
```

```go
import "github.com/tx7do/go-utils/sqlutil"
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/sqlutil"
)

func main() {
	raw := `UPDATE users SET nickname = '喵', age = 30 WHERE id = 5`

	// 审计落库前脱敏
	masked := sqlutil.MaskSQL(raw)
	fmt.Println(masked)
	// 输出: UPDATE users SET nickname = ***, age = *** WHERE id = ***

	// 数据分类 / 访问面分析（建议对脱敏后的文本提取）
	tables := sqlutil.ExtractTables(masked)
	fmt.Println(tables)
	// 输出: [users]

	// 记录脱敏策略版本，便于日后审计回溯
	_ = sqlutil.MaskingRules // "sql_literals:v1"
}
```

## API 参考

### MaskSQL

```go
func MaskSQL(sql string) string
```

对 SQL 文本做字面量级脱敏，按 PostgreSQL 词法规则逐字符扫描：

**脱敏为 `***` 的内容**

- 单引号字符串字面量 `'…'`，含 `''` 与 `\'` 转义
- 转义字符串 `E'…'`（与普通串同一套扫描规则）
- 美元引用 `$tag$…$tag$`（标签可为空 `$$`，可跨行，整体脱敏）
- 数值字面量（整数/小数/科学计数/十六进制，如 `25`、`1.5e-3`、`0xFF`）

**原样保留的内容**

- 关键字与普通标识符
- 双引号标识符 `"sys_users"`（含 `""` 转义）
- `$n` 占位符（`$1`、`$12` 等，参数值本就不在 SQL 文本中）
- 行注释 `-- …` 与块注释 `/* … */`（PG 语义支持嵌套），注释中的引号不会误判
- `::` 类型转换（`'a'::text` → `***::text`）

**边界行为**

- 未闭合的单引号字符串：脱敏到文本末尾（宁多掩不漏掩）
- 未闭合的美元引用（找不到闭合标签）：该 `$` 按字面量原样保留，后续内容按普通
  规则继续扫描
- 未闭合的双引号标识符：原样保留到末尾
- 孤立的 `$`：原样保留

**指纹稳定性**：结构相同、字面量不同的 SQL 脱敏后输出完全一致，适合按脱敏文本
做 SQL digest 分组统计。

### MaskingRules

```go
const MaskingRules = "sql_literals:v1"
```

当前脱敏策略的标识常量，调用方可将其与脱敏后的 SQL 一并存入审计表，便于日后
策略演进时区分新旧数据。

### ExtractTables

```go
func ExtractTables(sql string) []string
```

从 SQL 文本提取被访问的表名（词法器与 `MaskSQL` 同源：字符串与注释整体跳过、
双引号标识符整体成词）：

- **触发的语句**：`FROM`、`INSERT INTO`、`UPDATE`、`DELETE FROM`、`JOIN`
  （含 `LEFT JOIN` / `RIGHT JOIN` / `INNER JOIN` 等前缀写法）以及 `FROM` 后的逗号多表
- **跳过子查询括号**：子查询内层的 `FROM` 会在主扫描中被再次捕获，表不会遗漏
- **schema 限定名保留全名**：`public."sys_users"` → `public.sys_users`
- **按首个出现顺序去重**（小写形式判重，保留首次出现的原始写法）
- 允许 `[schema.]table [AS alias]` 形式，别名不计入表名

**未覆盖的语句类型**：`TRUNCATE`、`CREATE` / `ALTER` / `DROP` 等 DDL、`GRANT`、
`SHOW` 等语句中的表名不会被提取；`WITH` 开头的 CTE 名本身不触发提取，但 CTE 体内
的 `FROM` 与外层引用的表会经由 `FROM` 捕获。

## 示例

以下示例均取自单元测试（`sqlmask_test.go`、`extract_tables_test.go`）。

### MaskSQL 输入输出

| 输入 | 输出 | 说明 |
|------|------|------|
| `SELECT id FROM t WHERE email = 'a@b.c'` | `SELECT id FROM t WHERE email = ***` | 字符串字面量脱敏 |
| `WHERE n = 'O''Brien'` | `WHERE n = ***` | `''` 转义整体脱敏 |
| `WHERE n = 'O\'Brien'` | `WHERE n = ***` | `\'` 转义整体脱敏 |
| `WHERE a = E'x\ny'` | `WHERE a = ***` | E 转义串 |
| `WHERE age > 25 AND score = 1.5e-3` | `WHERE age > *** AND score = ***` | 数值字面量（含科学计数） |
| <code>WHERE id = $1 AND t."uid" = $12</code> | 原样保留 | `$n` 占位符与双引号标识符 |
| `SELECT 1 -- don't leak 'x'` 换行 `FROM t` | `SELECT *** -- don't leak 'x'` 换行 `FROM t` | 行注释原样且其中引号不误判 |
| `/* 'a' /* 'b' */ 'c' */ SELECT 'd'` | `/* 'a' /* 'b' */ 'c' */ SELECT ***` | 嵌套块注释原样 |
| `DO $$ BEGIN RAISE 'hi'; END $$` | `DO ***` | 空标签美元引用整体脱敏 |
| `$fn$ body 'q' $fn$` | `***` | 带标签美元引用整体脱敏（可跨行） |
| `INSERT INTO t (a,b) VALUES ('x', 42) RETURNING id` | `INSERT INTO t (a,b) VALUES (***, ***) RETURNING id` | INSERT VALUES |
| `WHERE x = 'a'::text AND n = 5::int` | `WHERE x = ***::text AND n = ***::int` | `::` 类型转换保留 |
| `SET a = ''` | `SET a = ***` | 空串脱敏 |
| `WHERE a = 'abc`（未闭合） | `WHERE a = ***` | 未闭合串脱敏到末尾 |
| `UPDATE users SET nickname = '喵', age = 30 WHERE id = 5` | `UPDATE users SET nickname = ***, age = *** WHERE id = ***` | UPDATE SET |
| `WHERE a = 0xFF` | `WHERE a = ***` | 十六进制数值 |

### ExtractTables 输入输出

| 输入 | 输出 |
|------|------|
| <code>SELECT "sys_users"."id" FROM "sys_users" WHERE … LIMIT *** OFFSET ***</code> | `[sys_users]` |
| <code>INSERT INTO "sys_users" ("name", "age") VALUES ($1, $2) RETURNING "id"</code> | `[sys_users]` |
| <code>INSERT INTO "sys_users" … ON CONFLICT (id) DO UPDATE SET name = $3</code> | `[sys_users]` |
| <code>UPDATE "sys_users" SET "name" = $1 WHERE "id" = $2</code> | `[sys_users]` |
| <code>DELETE FROM "sys_users" WHERE "id" = $1</code> | `[sys_users]` |
| <code>SELECT * FROM "sys_user_org_units" AS u0 JOIN "sys_org_units" AS u1 ON … LEFT JOIN "sys_positions" p ON …</code> | `[sys_user_org_units sys_org_units sys_positions]` |
| <code>SELECT * FROM "a", "b" WHERE x = y</code> | `[a b]` |
| <code>SELECT * FROM (SELECT "id" FROM "sys_users") AS t WHERE t."id" = ***</code> | `[sys_users]` |
| <code>SELECT * FROM "sys_users" JOIN "sys_roles" ON 1=1 JOIN "sys_users" ON 2=2</code> | `[sys_users sys_roles]`（去重） |
| `SELECT * FROM public."sys_users"` | `[public.sys_users]`（保留 schema 全名） |
| `select * from sys_users where id = 1` | `[sys_users]`（小写裸词同样支持） |
| `SHOW server_version` | `[]` |

## 注意事项

- **先脱敏再提取**：`ExtractTables` 的设计前提是输入已经过 `MaskSQL` 脱敏
  （测试对原始与脱敏文本均断言了相同结果，但生产链路建议固定"脱敏 → 提取"顺序，
  避免字面量中的内容干扰词法）。
- 脱敏策略是"宁多掩不漏掩"：普通串中的 `\` 也按转义处理，在
  `standard_conforming_strings` 开启的 PG 下可能多吞一个字符，但不会漏脱敏。
- 本工具面向 PostgreSQL 方言的词法规则；MySQL 反引号标识符、SQL Server 方括号
  标识符等未按特殊词法处理。
- `MaskSQL` 不改写 SQL 语义，仅替换字面量文本；不能替代参数化查询等根本性防注入手段。
