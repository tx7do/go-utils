# ddl_parser

[English](README.md) | 简体中文

轻量级 SQL DDL 解析器，专注于从 SQL 脚本中提取 `CREATE TABLE` 定义：解析表名、列（名称、类型、可空性、主键、默认值、注释、自增、唯一）以及 MySQL 风格的表属性（ENGINE / CHARSET / COLLATE / COMMENT），返回纯 Go 结构体，便于代码生成、表结构比对等场景使用。

> 注意：import 路径为 `github.com/tx7do/go-utils/ddl_parser`，但包名是 `ddlparser`（无下划线）。

## Features

- 单条语句解析 `ParseCreateTable`（返回 `*TableDef`）与整段脚本解析 `ParseCreateTables`（返回 `[]*TableDef`）
- 多方言基础语法：MySQL（`ENGINE` / `CHARSET` / `COLLATE` / 表与列 `COMMENT` / `AUTO_INCREMENT`）、PostgreSQL（`SERIAL`、数组类型、生成列）、SQLite（`AUTOINCREMENT`、`WITHOUT ROWID`）
- 列属性：`NOT NULL` / `NULL`、`PRIMARY KEY`（行内与表级）、`DEFAULT`、`COMMENT`、`AUTO_INCREMENT` / `AUTOINCREMENT`、`UNIQUE`
- 关键字大小写不敏感；支持反引号 / 双引号包裹的标识符（引号内内容保持原样）
- 语句按 `;` 切分（引号与括号内的分号不切分）；非 `CREATE TABLE` 语句静默跳过
- 纯文本/正则实现，无 SQL 引擎依赖，运行时零第三方依赖

## Installation

本包为独立 Go 模块，单独安装：

```bash
go get github.com/tx7do/go-utils/ddl_parser
```

```go
import "github.com/tx7do/go-utils/ddl_parser"
```

## Quick Start

完整可编译示例：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/ddl_parser"
)

func main() {
	sql := `
	CREATE TABLE users (id INT PRIMARY KEY, name TEXT);
	CREATE TABLE orders (id INT PRIMARY KEY, user_id INT);
	`

	tables, err := ddlparser.ParseCreateTables(sql)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tables {
		fmt.Println(t.Name) // users / orders
	}
}
```

单条语句解析：

```go
table, err := ddlparser.ParseCreateTable(`
	CREATE TABLE IF NOT EXISTS products (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Product ID',
		name VARCHAR(100) NOT NULL COMMENT 'Product Name',
		price DECIMAL(10, 2) NOT NULL DEFAULT 0.00
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Products Table';
`)
// table.Name     => "products"
// table.Engine   => "innodb"
// table.Charset  => "utf8mb4"
// table.Comment  => "Products Table"
// table.Columns[0].PrimaryKey  => true
// table.Columns[1].Comment     => "Product Name"
// table.Columns[2].Default     => "0.00"
```

## API 参考

### 解析函数（parser.go）

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `ParseCreateTable` | `func ParseCreateTable(sql string) (*TableDef, error)` | 解析单条 `CREATE TABLE` 语句；输入不是 CREATE TABLE、缺表名、缺括号或括号不匹配时返回错误 |
| `ParseCreateTables` | `func ParseCreateTables(sql string) ([]*TableDef, error)` | 按 `;` 切分（引号/括号内不切）后逐条解析；**非 `CREATE TABLE` 语句静默跳过**；任一 CREATE TABLE 解析失败则整体返回错误 |

### 返回结构（types.go）

`ColumnDef`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Name` | `string` | 列名（去除引号） |
| `Type` | `string` | 原始类型文本，如 `varchar(255)`、`decimal(10, 2)` |
| `Nullable` | `bool` | 默认 `true`；遇到 `NOT NULL`、`PRIMARY KEY`、`AUTO_INCREMENT` 时为 `false` |
| `PrimaryKey` | `bool` | 行内 `PRIMARY KEY`、表级 `PRIMARY KEY (...)` 引用的列，或 MySQL `AUTO_INCREMENT`（隐含主键） |
| `Default` | `string` | 默认值原文，如 `30`、`'unknown'`（保留引号）、`current_timestamp` |
| `Comment` | `string` | 列 `COMMENT '...'` 内容（去除引号） |
| `AutoIncrement` | `bool` | `AUTO_INCREMENT` / `AUTOINCREMENT` |
| `Unique` | `bool` | 行内 `UNIQUE` |

`TableDef`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Name` | `string` | 表名（去除引号） |
| `Columns` | `[]ColumnDef` | 列定义列表 |
| `Indexes` | `[]string` | 预留字段；当前表级索引定义被跳过，该字段始终为空 |
| `Engine` | `string` | MySQL `ENGINE=`（如 `innodb`） |
| `Charset` | `string` | `DEFAULT CHARSET=` / `CHARACTER SET`（如 `utf8mb4`） |
| `Comment` | `string` | 表级 `COMMENT='...'`（`=` 可省略） |
| `Collation` | `string` | `COLLATE=` / `COLLATION=`（如 `utf8mb4_unicode_ci`） |

## Notes / Caveats

- **只处理 `CREATE TABLE`**：`ParseCreateTables` 对其它语句（`INSERT` / `SELECT` / `ALTER` 等）静默忽略、不报错；`ParseCreateTable` 直接喂非 CREATE TABLE 语句会返回错误。
- **大小写不敏感**：解析前会做标准化——去除块注释 `/* */` 与行注释 `--`，引号外内容全部转小写。因此**输出中未加引号的标识符与类型也是小写**（如 `InnoDB` → `innodb`、`INT` → `int`）；引号内内容保留原样。
- **表级约束跳过**：`PRIMARY KEY (...)`（仅取其列名回填到列的 `PrimaryKey`）、`FOREIGN KEY`、`CONSTRAINT`、`INDEX` / `KEY`、`FULLTEXT`、`SPATIAL`、`UNIQUE KEY` / `UNIQUE INDEX` 都不作为列输出；无法解析的列片段也会被容错跳过。
- **不是完整 SQL 解析器**：不做类型校验、不执行 DDL、不处理嵌套复杂语法；`CHECK`、生成列等以容错方式处理（忽略无法识别的 token）。
- 空列定义 `CREATE TABLE t ()` 合法，得到 0 列。

## Tests

在模块目录内运行：

```bash
cd ddl_parser
go test ./...
```
