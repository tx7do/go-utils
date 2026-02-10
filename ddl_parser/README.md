# ddl_parser

Small SQL DDL parser focused on extracting CREATE TABLE definitions.

## Usage

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/ddl_parser"
)

func main() {
	sql := `
	CREATE TABLE users (id INT PRIMARY KEY, name TEXT);
	CREATE TABLE orders (id INT PRIMARY KEY, user_id INT);
	`

	tables, err := ddlparser.ParseCreateTables(sql)
	if err != nil {
		panic(err)
	}

	for _, t := range tables {
		fmt.Println(t.Name)
	}
}
```

## Tests

Run tests from the module directory:

```powershell
cd D:\GoProject\go-utils\ddl_parser
go test ./...
```

