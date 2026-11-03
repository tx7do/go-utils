# sqlutil

PostgreSQL lexer-level SQL auditing utilities, implemented with the standard
library only (zero third-party runtime dependencies).

English | [简体中文](README.zh-CN.md)

## Features

- **`MaskSQL`**: literal-level masking — string literals (`'…'`, `E'…'`,
  `$tag$…$tag$` dollar-quoted strings) and numeric literals are replaced with
  `***` in their entirety, while the SQL structure is preserved (keywords,
  double-quoted identifiers, `$n` placeholders, comments, `::` casts).
  Implemented per PostgreSQL lexical rules: both `\'` and `''` inside single-quoted
  strings are treated as escapes (over-mask rather than under-mask), dollar quoting
  may span lines, block comments nest, and quotes inside comments or identifiers
  are never misread.
- **`ExtractTables`**: extracts the table names accessed by a (masked) SQL text,
  covering FROM / INSERT INTO / UPDATE / DELETE FROM / JOIN and comma-separated
  FROM lists, skipping subquery parentheses, keeping schema-qualified names in
  full, deduplicated in order of first appearance. Strings and comments never
  split words.
- **`MaskingRules`**: a masking-policy identifier constant (`"sql_literals:v1"`)
  that can be persisted alongside audit records for traceability.

## Installation

sqlutil is an independent Go module (`sqlutil/go.mod`) with no third-party
runtime dependencies:

```bash
go get github.com/tx7do/go-utils/sqlutil
```

```go
import "github.com/tx7do/go-utils/sqlutil"
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/sqlutil"
)

func main() {
	raw := `UPDATE users SET nickname = '喵', age = 30 WHERE id = 5`

	// Mask literals before persisting to the audit log.
	masked := sqlutil.MaskSQL(raw)
	fmt.Println(masked)
	// Output: UPDATE users SET nickname = ***, age = *** WHERE id = ***

	// Data classification / access-surface analysis (extract from masked text).
	tables := sqlutil.ExtractTables(masked)
	fmt.Println(tables)
	// Output: [users]

	// Record the masking policy version for later audits.
	_ = sqlutil.MaskingRules // "sql_literals:v1"
}
```

## API Reference

### MaskSQL

```go
func MaskSQL(sql string) string
```

Masks SQL text at literal level by scanning character-by-character under
PostgreSQL lexical rules.

**Masked with `***`**

- Single-quoted string literals `'…'`, including `''` and `\'` escapes
- Escape strings `E'…'` (same scanning rules as plain strings)
- Dollar-quoted strings `$tag$…$tag$` (empty tag `$$` allowed, may span lines, masked as a whole)
- Numeric literals (integer/decimal/scientific/hexadecimal, e.g. `25`, `1.5e-3`, `0xFF`)

**Preserved verbatim**

- Keywords and bare identifiers
- Double-quoted identifiers `"sys_users"` (including `""` escapes)
- `$n` placeholders (`$1`, `$12`, ...; parameter values are not part of the SQL text anyway)
- Line comments `-- …` and block comments `/* … */` (PG semantics, nesting supported);
  quotes inside comments are never misread
- `::` casts (`'a'::text` → `***::text`)

**Edge behavior**

- Unterminated single-quoted string: masked through the end of the input
  (over-mask rather than under-mask)
- Unterminated dollar quote (no closing tag found): the `$` is kept verbatim and
  scanning continues under the normal rules
- Unterminated double-quoted identifier: preserved verbatim through the end
- A lone `$`: kept verbatim

**Fingerprint stability**: structurally identical SQL with different literals
masks to identical output — suitable for digest-style grouping on masked text.

### MaskingRules

```go
const MaskingRules = "sql_literals:v1"
```

Identifier constant of the current masking policy. Store it alongside masked SQL
in audit tables so that data masked under older policies can be told apart when
the policy evolves.

### ExtractTables

```go
func ExtractTables(sql string) []string
```

Extracts accessed table names from SQL text (the lexer shares its scanners with
`MaskSQL`: strings and comments are skipped whole, double-quoted identifiers
become single words):

- **Statements covered**: `FROM`, `INSERT INTO`, `UPDATE`, `DELETE FROM`, `JOIN`
  (including `LEFT JOIN` / `RIGHT JOIN` / `INNER JOIN` spellings) and comma-separated
  tables after `FROM`
- **Subquery parentheses are skipped**: the inner `FROM` of a subquery is captured
  again by the main scan, so no table is missed
- **Schema-qualified names keep their full form**: `public."sys_users"` → `public.sys_users`
- **Deduplicated in order of first appearance** (case-insensitively; the first
  original spelling is kept)
- Accepts `[schema.]table [AS alias]`; aliases are not counted as tables

**Statement types not covered**: table names in `TRUNCATE`, DDL such as
`CREATE` / `ALTER` / `DROP`, `GRANT`, `SHOW` and similar statements are not
extracted. A `WITH` CTE name itself does not trigger extraction, but the `FROM`
inside the CTE body and the tables referenced in the outer query are captured.

## Examples

The following examples are taken from the unit tests (`sqlmask_test.go`,
`extract_tables_test.go`).

### MaskSQL input/output

| Input | Output | Notes |
|-------|--------|-------|
| `SELECT id FROM t WHERE email = 'a@b.c'` | `SELECT id FROM t WHERE email = ***` | String literal masked |
| `WHERE n = 'O''Brien'` | `WHERE n = ***` | `''` escape masked as a whole |
| `WHERE n = 'O\'Brien'` | `WHERE n = ***` | `\'` escape masked as a whole |
| `WHERE a = E'x\ny'` | `WHERE a = ***` | E escape string |
| `WHERE age > 25 AND score = 1.5e-3` | `WHERE age > *** AND score = ***` | Numeric literals (incl. scientific) |
| <code>WHERE id = $1 AND t."uid" = $12</code> | unchanged | `$n` placeholders and quoted identifiers |
| `SELECT 1 -- don't leak 'x'` newline `FROM t` | `SELECT *** -- don't leak 'x'` newline `FROM t` | Line comment kept; its quotes not misread |
| `/* 'a' /* 'b' */ 'c' */ SELECT 'd'` | `/* 'a' /* 'b' */ 'c' */ SELECT ***` | Nested block comment kept |
| `DO $$ BEGIN RAISE 'hi'; END $$` | `DO ***` | Empty-tag dollar quote masked as a whole |
| `$fn$ body 'q' $fn$` | `***` | Tagged dollar quote masked as a whole (may span lines) |
| `INSERT INTO t (a,b) VALUES ('x', 42) RETURNING id` | `INSERT INTO t (a,b) VALUES (***, ***) RETURNING id` | INSERT VALUES |
| `WHERE x = 'a'::text AND n = 5::int` | `WHERE x = ***::text AND n = ***::int` | `::` casts preserved |
| `SET a = ''` | `SET a = ***` | Empty string masked |
| `WHERE a = 'abc` (unterminated) | `WHERE a = ***` | Unterminated string masked to the end |
| `UPDATE users SET nickname = '喵', age = 30 WHERE id = 5` | `UPDATE users SET nickname = ***, age = *** WHERE id = ***` | UPDATE SET |
| `WHERE a = 0xFF` | `WHERE a = ***` | Hexadecimal number |

### ExtractTables input/output

| Input | Output |
|-------|--------|
| <code>SELECT "sys_users"."id" FROM "sys_users" WHERE … LIMIT *** OFFSET ***</code> | `[sys_users]` |
| <code>INSERT INTO "sys_users" ("name", "age") VALUES ($1, $2) RETURNING "id"</code> | `[sys_users]` |
| <code>INSERT INTO "sys_users" … ON CONFLICT (id) DO UPDATE SET name = $3</code> | `[sys_users]` |
| <code>UPDATE "sys_users" SET "name" = $1 WHERE "id" = $2</code> | `[sys_users]` |
| <code>DELETE FROM "sys_users" WHERE "id" = $1</code> | `[sys_users]` |
| <code>SELECT * FROM "sys_user_org_units" AS u0 JOIN "sys_org_units" AS u1 ON … LEFT JOIN "sys_positions" p ON …</code> | `[sys_user_org_units sys_org_units sys_positions]` |
| <code>SELECT * FROM "a", "b" WHERE x = y</code> | `[a b]` |
| <code>SELECT * FROM (SELECT "id" FROM "sys_users") AS t WHERE t."id" = ***</code> | `[sys_users]` |
| <code>SELECT * FROM "sys_users" JOIN "sys_roles" ON 1=1 JOIN "sys_users" ON 2=2</code> | `[sys_users sys_roles]` (deduplicated) |
| `SELECT * FROM public."sys_users"` | `[public.sys_users]` (schema kept) |
| `select * from sys_users where id = 1` | `[sys_users]` (lower-case bare words work too) |
| `SHOW server_version` | `[]` |

## Notes

- **Mask first, then extract**: `ExtractTables` is designed for text that has
  already been through `MaskSQL` (tests assert identical results for both raw and
  masked input, but production pipelines should keep the fixed order "mask →
  extract" so literal contents cannot disturb the lexer).
- The masking policy is "over-mask rather than under-mask": `\` inside plain
  strings is treated as an escape even though PG's `standard_conforming_strings`
  makes it a literal backslash — at worst one extra character is swallowed,
  never a leak.
- The tools target PostgreSQL lexical rules; MySQL backtick identifiers and
  SQL Server bracketed identifiers are not given special treatment.
- `MaskSQL` does not change SQL semantics, it only replaces literal text; it is
  not a substitute for parameterized queries and other primary injection defenses.
