# ddl_parser

English | [简体中文](README.zh-CN.md)

A lightweight SQL DDL parser focused on extracting `CREATE TABLE` definitions from SQL scripts: it parses the table name, columns (name, type, nullability, primary key, default, comment, auto-increment, unique) and MySQL-style table attributes (ENGINE / CHARSET / COLLATE / COMMENT) into plain Go structs, useful for code generation, schema diffing and similar tasks.

> Note: the import path is `github.com/tx7do/go-utils/ddl_parser`, but the package name is `ddlparser` (no underscore).

## Features

- Single-statement parsing with `ParseCreateTable` (returns `*TableDef`) and whole-script parsing with `ParseCreateTables` (returns `[]*TableDef`)
- Multi-dialect basics: MySQL (`ENGINE` / `CHARSET` / `COLLATE` / table and column `COMMENT` / `AUTO_INCREMENT`), PostgreSQL (`SERIAL`, array types, generated columns), SQLite (`AUTOINCREMENT`, `WITHOUT ROWID`)
- Column attributes: `NOT NULL` / `NULL`, `PRIMARY KEY` (inline and table-level), `DEFAULT`, `COMMENT`, `AUTO_INCREMENT` / `AUTOINCREMENT`, `UNIQUE`
- Case-insensitive keywords; identifiers wrapped in backticks / double quotes are supported (quoted content is preserved as-is)
- Statements are split on `;` (semicolons inside quotes or parentheses do not split); non-`CREATE TABLE` statements are silently skipped
- Pure text/regex implementation, no SQL engine, zero third-party runtime dependencies

## Installation

This package is an independent Go module; install it on its own:

```bash
go get github.com/tx7do/go-utils/ddl_parser
```

```go
import "github.com/tx7do/go-utils/ddl_parser"
```

## Quick Start

Complete, compilable example:

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

Parsing a single statement:

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

## API Reference

### Parsing functions (parser.go)

| Function | Signature | Description |
| --- | --- | --- |
| `ParseCreateTable` | `func ParseCreateTable(sql string) (*TableDef, error)` | Parses one `CREATE TABLE` statement; returns an error if the input is not a CREATE TABLE, or the table name / parentheses are missing or unbalanced |
| `ParseCreateTables` | `func ParseCreateTables(sql string) ([]*TableDef, error)` | Splits on `;` (quote/parenthesis aware) and parses each statement; **non-`CREATE TABLE` statements are silently skipped**; returns an error as a whole if any CREATE TABLE fails to parse |

### Result structs (types.go)

`ColumnDef`:

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Column name (quotes stripped) |
| `Type` | `string` | Raw type text, e.g. `varchar(255)`, `decimal(10, 2)` |
| `Nullable` | `bool` | Defaults to `true`; becomes `false` on `NOT NULL`, `PRIMARY KEY` or `AUTO_INCREMENT` |
| `PrimaryKey` | `bool` | Inline `PRIMARY KEY`, columns referenced by a table-level `PRIMARY KEY (...)`, or MySQL `AUTO_INCREMENT` (implies primary key) |
| `Default` | `string` | Raw default token, e.g. `30`, `'unknown'` (quotes kept), `current_timestamp` |
| `Comment` | `string` | Content of the column `COMMENT '...'` (quotes stripped) |
| `AutoIncrement` | `bool` | `AUTO_INCREMENT` / `AUTOINCREMENT` |
| `Unique` | `bool` | Inline `UNIQUE` |

`TableDef`:

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Table name (quotes stripped) |
| `Columns` | `[]ColumnDef` | Column definitions |
| `Indexes` | `[]string` | Reserved field; table-level index definitions are currently skipped, so this is always empty |
| `Engine` | `string` | MySQL `ENGINE=` (e.g. `innodb`) |
| `Charset` | `string` | `DEFAULT CHARSET=` / `CHARACTER SET` (e.g. `utf8mb4`) |
| `Comment` | `string` | Table-level `COMMENT='...'` (`=` is optional) |
| `Collation` | `string` | `COLLATE=` / `COLLATION=` (e.g. `utf8mb4_unicode_ci`) |

## Notes / Caveats

- **Only `CREATE TABLE` is handled**: `ParseCreateTables` silently ignores other statements (`INSERT` / `SELECT` / `ALTER` etc.) without error; feeding a non-CREATE TABLE statement directly to `ParseCreateTable` returns an error.
- **Case-insensitive**: the input is normalized first — block comments `/* */` and line comments `--` are removed, and everything outside quotes is lowercased. Therefore **unquoted identifiers and types in the output are lowercase too** (e.g. `InnoDB` → `innodb`, `INT` → `int`); quoted content is preserved verbatim.
- **Table-level constraints are skipped**: `PRIMARY KEY (...)` (only its column names are used to mark `PrimaryKey` on the columns), `FOREIGN KEY`, `CONSTRAINT`, `INDEX` / `KEY`, `FULLTEXT`, `SPATIAL`, `UNIQUE KEY` / `UNIQUE INDEX` are not emitted as columns; unparseable column fragments are also skipped on a best-effort basis.
- **Not a full SQL parser**: no type validation, no DDL execution, no nested complex syntax; `CHECK` constraints, generated columns etc. are tolerated (unrecognized tokens are ignored).
- An empty column list `CREATE TABLE t ()` is valid and yields zero columns.

## Tests

Run tests from the module directory:

```bash
cd ddl_parser
go test ./...
```
