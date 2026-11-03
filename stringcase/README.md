# stringcase

Go naming-case conversion and word-splitting utilities: UpperCamelCase /
lowerCamelCase, snake_case, kebab-case, plus identifier word segmentation.

English | [简体中文](README.zh-CN.md)

## Features

- **Camel case**: `UpperCamelCase` (Pascal / UpperCamelCase) and `LowerCamelCase`
  (lowerCamelCase), with `ToPascalCase` / `ToCamelCase` / `PascalCase` / `CamelCase` aliases
- **Snake case**: `ToSnakeCase` (snake_case), `UpperSnakeCase` (UPPER_SNAKE_CASE),
  and the `IsSnakeCase` predicate
- **Kebab case**: `KebabCase` (kebab-case) and `UpperKebabCase` (KEBAB-CASE)
- **Word splitting**: `Split` cuts an identifier into words along naming boundaries,
  with a `noSplit` parameter to protect words that must stay whole
- **Character utilities**: non-alphanumeric replacement/splitting and a generic `ContainsFn`
- Any input style converts to any other: `snake_case`, `kebab-case`, `camelCase`,
  `PascalCase`, space/dot separated words, and consecutive-uppercase acronyms
  (`HTTPStatusCode`) are all recognized

## Installation

stringcase is a subpackage of the root module `github.com/tx7do/go-utils`
(no go.mod of its own); installing the root module is enough:

```bash
go get github.com/tx7do/go-utils
```

```go
import "github.com/tx7do/go-utils/stringcase"
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/stringcase"
)

func main() {
	// UpperCamelCase (Pascal case): every word capitalized.
	fmt.Println(stringcase.UpperCamelCase("hello world"))       // HelloWorld
	fmt.Println(stringcase.UpperCamelCase("HTTPStatusCode"))    // HttpStatusCode
	fmt.Println(stringcase.UpperCamelCase("123NumberPrefix"))   // 123NumberPrefix

	// LowerCamelCase: first word lower-case, following words capitalized.
	fmt.Println(stringcase.LowerCamelCase("HTTPRequest"))       // httpRequest
	fmt.Println(stringcase.LowerCamelCase("UserID"))            // userId

	// Snake case.
	fmt.Println(stringcase.ToSnakeCase("CamelCase"))            // camel_case
	fmt.Println(stringcase.ToSnakeCase("HTTPStatusCode"))       // http_status_code
	fmt.Println(stringcase.ToSnakeCase("userID"))               // user_id
	fmt.Println(stringcase.ToSnakeCase("md5Hash"))              // md5_hash
	fmt.Println(stringcase.UpperSnakeCase("FooID"))             // FOO_ID
	fmt.Println(stringcase.IsSnakeCase("user_id"))              // true
	fmt.Println(stringcase.IsSnakeCase("hello-world"))          // false

	// Kebab case.
	fmt.Println(stringcase.KebabCase("HTTPStatusCode"))         // http-status-code
	fmt.Println(stringcase.UpperKebabCase("HTTPStatusCode"))    // HTTP-STATUS-CODE

	// Word splitting.
	fmt.Println(stringcase.Split("HTTPStatusCode"))             // [HTTP Status Code]
	fmt.Println(stringcase.Split("ParseURLDoParse"))            // [Parse URL Do Parse]
	fmt.Println(stringcase.Split("helloWorld"))                 // [hello World]
	fmt.Println(stringcase.Split("numbers123test"))             // [numbers 123 test]

	// noSplit protects the given words from being split.
	fmt.Println(stringcase.Split("parseHTTPURL", "HTTPURL"))    // [parse HTTPURL]
	fmt.Println(stringcase.Split("userID", "userID"))           // [userID]
}
```

## API Reference

### Camel case (camel_case.go)

| Function | Signature | Description |
|----------|-----------|-------------|
| `UpperCamelCase` | `func(input string) string` | Converts to UpperCamelCase / Pascal case (e.g. `HelloWorld`) |
| `ToPascalCase` | `func(input string) string` | Alias of `UpperCamelCase` |
| `PascalCase` | `func(input string) string` | Alias of `UpperCamelCase` |
| `LowerCamelCase` | `func(input string) string` | Converts to lowerCamelCase (e.g. `helloWorld`) |
| `ToCamelCase` | `func(input string) string` | Alias of `LowerCamelCase` |
| `CamelCase` | `func(input string) string` | Alias of `LowerCamelCase` |

Behavior: the input is split into words first, then each word is normalized — only
the first letter keeps its case declaration and the rest of the word is lower-cased —
before everything is joined. Consecutive-uppercase acronyms are therefore "flattened":
`HTTPStatusCode` → `HttpStatusCode` (UpperCamelCase) / `httpStatusCode`
(LowerCamelCase). Empty input returns an empty string.

```go
stringcase.UpperCamelCase("user ID")   // UserId
stringcase.LowerCamelCase("PascalCaseString") // pascalCaseString
```

### Snake case (snake_case.go)

| Function | Signature | Description |
|----------|-----------|-------------|
| `ToSnakeCase` | `func(input string) string` | Converts to snake_case (lower-case with underscores); equivalent to `SnakeCase` |
| `SnakeCase` | `func(s string) string` | Converts to snake_case |
| `UpperSnakeCase` | `func(s string) string` | Converts to UPPER_SNAKE_CASE |
| `IsSnakeCase` | `func(s string) bool` | Reports whether the string is valid snake_case |

An adjacent "letters-only word + digits-only word" pair is merged into one word
(e.g. `ID123Test` → `id123_test`, `Numbers123Test` → `numbers123_test`).

`IsSnakeCase` rules (any violation returns `false`):

- Not empty
- Must not start or end with an underscore
- Must not contain uppercase letters
- No consecutive underscores
- Only lowercase letters, digits and underscores are allowed

```go
stringcase.IsSnakeCase("snake_case") // true
stringcase.IsSnakeCase("md5_hash")   // true
stringcase.IsSnakeCase("123")        // true
stringcase.IsSnakeCase("Snake_case") // false (uppercase letter)
stringcase.IsSnakeCase("a__b")       // false (consecutive underscores)
stringcase.IsSnakeCase("_a")         // false (leading underscore)
```

### Kebab case (kebab_case.go)

| Function | Signature | Description |
|----------|-----------|-------------|
| `KebabCase` | `func(s string) string` | Converts to kebab-case (lower-case with hyphens), e.g. `http-status-code` |
| `UpperKebabCase` | `func(s string) string` | Converts to KEBAB-CASE (uppercase with hyphens), e.g. `HTTP-STATUS-CODE` |

### Word splitting (split.go)

```go
func Split(input string, noSplit ...string) []string
```

Splits the input as a naming identifier into words:

- Every non-alphanumeric character acts as a delimiter (see "Character utilities" below)
- Splits along lowerCamelCase boundaries: `helloWorld` → `hello` / `World`
- Consecutive uppercase letters stay together as an acronym until an
  "uppercase followed by lowercase" boundary: `HTTPStatusCode` →
  `HTTP` / `Status` / `Code`, and `ParseURLDoParse` → `Parse` / `URL` / `Do` / `Parse`
- Letter/digit boundaries split too: `numbers123test` → `numbers` / `123` / `test`
- The variadic `noSplit` lists words that must not be split (protected by prefix
  matching): `Split("parseHTTPURL", "HTTPURL")` → `[parse HTTPURL]`
- For an empty string or invalid UTF-8 input, a single-element slice containing the
  original input is returned (`Split("")` → `[""]`)

### Character utilities (utils.go)

| Function | Signature | Description |
|----------|-----------|-------------|
| `ReplaceNonAlphanumeric` | `func(s string, replacement string) string` | Replaces every run of non-alphanumeric characters with `replacement`; an empty `replacement` falls back to `_` |
| `SplitByNonAlphanumeric` | `func(input string) []string` | Splits on non-alphanumeric characters (equivalent to replacing them with spaces then splitting on whitespace); empty segments dropped |
| `SplitAndKeepDelimiters` | `func(input string) []string` | Splits but keeps the delimiters: words and separators alternate in the result |
| `ContainsFn` | `func[T any](slice []T, value T, predicate func(got, want T) bool) bool` | Generic containment check: calls `predicate(item, value)` per element; returns `true` on the first match |

```go
stringcase.ReplaceNonAlphanumeric("a-b c", "-")  // a-b-c
stringcase.ReplaceNonAlphanumeric("a-b c", "")   // a_b_c (empty replacement falls back to _)
stringcase.SplitByNonAlphanumeric("hello_world 42!")   // [hello world 42]
stringcase.SplitAndKeepDelimiters("hello_world 42!")   // [hello _ world   42 !]

stringcase.ContainsFn([]string{"HTTPURL"}, "HTTP",
	func(got, want string) bool { return strings.HasPrefix(got, want) },
) // true ("HTTPURL" has prefix "HTTP")
```

Note that the predicate semantics of `ContainsFn` are entirely up to the caller:
in `predicate(item, value)` the first argument is the slice element and the second
is the value being searched for. The `HasPrefix` usage above is exactly how this
package implements `noSplit` word protection.

## Background: Common Naming Conventions

- **lowerCamelCase (camelCase)**: the first word starts with a lowercase letter and
  following words start with uppercase letters, with no separators. Example:
  `myVariableName`. Common for variables and functions in JavaScript, Java and C#.
- **UpperCamelCase (PascalCase)**: every word — including the first — starts with an
  uppercase letter. Example: `MyVariableName`. Common for classes, interfaces and
  types in C#, Java and TypeScript.
- **snake_case**: words separated by underscores `_`. Example: `first_name`. Common
  for variables and functions in Python and Ruby, and for database column names.
- **kebab-case**: words separated by hyphens `-`, also known as `dash-case`,
  `hyphen-case` or `lisp-case`. Example: `user-profile`. Common for URLs, file
  names and HTML/CSS class names.

## Notes

- All conversion functions `TrimSpace` the input before splitting; empty input
  returns an empty string (`Split` is the exception, returning `[""]`).
- Within a word only the first letter keeps its case (camel) or everything is mapped
  to the target style (snake/kebab), so `ParseURL.DoParse` and `ParseUrl.DoParse`
  produce the same result (`ParseUrlDoParse` / `parse_url_do_parse`).
- Leading/trailing delimiters are ignored: `__leading_underscores` → `LeadingUnderscores`.
- Numeric prefixes keep their position: `123NumberPrefix` stays `123NumberPrefix`
  under UpperCamelCase.
- Non-alphanumeric characters (punctuation, symbols, emoji, ...) always act as
  delimiters. Non-ASCII letters/digits (e.g. Chinese) are kept as one word and pass
  through conversions unchanged (`ToSnakeCase("影師天下太平")` → `"影師天下太平"`),
  but naming-boundary detection (camel humps, letter/digit boundaries) only works
  on ASCII; verify mixed Chinese/English output against your expectations.
