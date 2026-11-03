# stringcase

Go 字符串命名风格（case）转换与分词工具：大小驼峰、蛇形、烤肉串互转，以及按命名
边界切分单词。

[English](README.md) | 简体中文

## 特性

- **驼峰命名**：`UpperCamelCase`（大驼峰/帕斯卡）与 `LowerCamelCase`（小驼峰），
  另有 `ToPascalCase` / `ToCamelCase` / `PascalCase` / `CamelCase` 别名
- **蛇形命名**：`ToSnakeCase`（snake_case）、`UpperSnakeCase`（UPPER_SNAKE_CASE）、
  `IsSnakeCase` 判定
- **烤肉串命名**：`KebabCase`（kebab-case）、`UpperKebabCase`（KEBAB-CASE）
- **分词**：`Split` 按命名边界切出单词，支持 `noSplit` 参数保护指定词不被拆分
- **字符工具**：非字母数字替换/切分、泛型 `ContainsFn`
- 任意输入风格可互相转换：`snake_case`、`kebab-case`、`camelCase`、`PascalCase`、
  空格/点号分隔、连续大写缩写（`HTTPStatusCode`）等都能识别

## 安装

stringcase 是根模块 `github.com/tx7do/go-utils` 的子包（无独立 go.mod），安装根模块
即可使用：

```bash
go get github.com/tx7do/go-utils
```

```go
import "github.com/tx7do/go-utils/stringcase"
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/stringcase"
)

func main() {
	// 大驼峰（帕斯卡命名）：每个单词首字母大写
	fmt.Println(stringcase.UpperCamelCase("hello world"))       // HelloWorld
	fmt.Println(stringcase.UpperCamelCase("HTTPStatusCode"))    // HttpStatusCode
	fmt.Println(stringcase.UpperCamelCase("123NumberPrefix"))   // 123NumberPrefix

	// 小驼峰：首单词首字母小写，后续单词首字母大写
	fmt.Println(stringcase.LowerCamelCase("HTTPRequest"))       // httpRequest
	fmt.Println(stringcase.LowerCamelCase("UserID"))            // userId

	// 蛇形命名
	fmt.Println(stringcase.ToSnakeCase("CamelCase"))            // camel_case
	fmt.Println(stringcase.ToSnakeCase("HTTPStatusCode"))       // http_status_code
	fmt.Println(stringcase.ToSnakeCase("userID"))               // user_id
	fmt.Println(stringcase.ToSnakeCase("md5Hash"))              // md5_hash
	fmt.Println(stringcase.UpperSnakeCase("FooID"))             // FOO_ID
	fmt.Println(stringcase.IsSnakeCase("user_id"))              // true
	fmt.Println(stringcase.IsSnakeCase("hello-world"))          // false

	// 烤肉串命名
	fmt.Println(stringcase.KebabCase("HTTPStatusCode"))         // http-status-code
	fmt.Println(stringcase.UpperKebabCase("HTTPStatusCode"))    // HTTP-STATUS-CODE

	// 分词
	fmt.Println(stringcase.Split("HTTPStatusCode"))             // [HTTP Status Code]
	fmt.Println(stringcase.Split("ParseURLDoParse"))            // [Parse URL Do Parse]
	fmt.Println(stringcase.Split("helloWorld"))                 // [hello World]
	fmt.Println(stringcase.Split("numbers123test"))             // [numbers 123 test]

	// noSplit 保护指定词不被拆分
	fmt.Println(stringcase.Split("parseHTTPURL", "HTTPURL"))    // [parse HTTPURL]
	fmt.Println(stringcase.Split("userID", "userID"))           // [userID]
}
```

## API 参考

### 驼峰命名（camel_case.go）

| 函数 | 签名 | 说明 |
|------|------|------|
| `UpperCamelCase` | `func(input string) string` | 转大驼峰（帕斯卡命名，如 `HelloWorld`） |
| `ToPascalCase` | `func(input string) string` | 同 `UpperCamelCase`（别名） |
| `PascalCase` | `func(input string) string` | 同 `UpperCamelCase`（别名） |
| `LowerCamelCase` | `func(input string) string` | 转小驼峰（如 `helloWorld`） |
| `ToCamelCase` | `func(input string) string` | 同 `LowerCamelCase`（别名） |
| `CamelCase` | `func(input string) string` | 同 `LowerCamelCase`（别名） |

行为要点：先按命名边界分词，再逐词规整——每个单词仅保留首字母的大小写声明、其余
字母统一小写，最后拼接。因此 `HTTPStatusCode` → `HttpStatusCode`（大驼峰）/
`httpStatusCode`（小驼峰），连续大写缩写会被"压平"。空输入返回空串。

```go
stringcase.UpperCamelCase("user ID")   // UserId
stringcase.LowerCamelCase("PascalCaseString") // pascalCaseString
```

### 蛇形命名（snake_case.go）

| 函数 | 签名 | 说明 |
|------|------|------|
| `ToSnakeCase` | `func(input string) string` | 转 snake_case（小写下划线），与 `SnakeCase` 等价 |
| `SnakeCase` | `func(s string) string` | 转 snake_case |
| `UpperSnakeCase` | `func(s string) string` | 转 UPPER_SNAKE_CASE |
| `IsSnakeCase` | `func(s string) bool` | 判断是否为合法 snake_case |

相邻"纯字母词 + 纯数字词"会合并为一个词（如 `ID123Test` → `id123_test`，
`Numbers123Test` → `numbers123_test`）。

`IsSnakeCase` 判定规则（不满足任一条即返回 `false`）：

- 非空
- 不能以下划线开头或结尾
- 不能包含大写字母
- 不允许连续下划线
- 仅允许小写字母、数字和下划线

```go
stringcase.IsSnakeCase("snake_case") // true
stringcase.IsSnakeCase("md5_hash")   // true
stringcase.IsSnakeCase("123")        // true
stringcase.IsSnakeCase("Snake_case") // false（含大写）
stringcase.IsSnakeCase("a__b")       // false（连续下划线）
stringcase.IsSnakeCase("_a")         // false（下划线开头）
```

### 烤肉串命名（kebab_case.go）

| 函数 | 签名 | 说明 |
|------|------|------|
| `KebabCase` | `func(s string) string` | 转 kebab-case（小写连字符），如 `http-status-code` |
| `UpperKebabCase` | `func(s string) string` | 转 KEBAB-CASE（大写连字符），如 `HTTP-STATUS-CODE` |

### 分词（split.go）

```go
func Split(input string, noSplit ...string) []string
```

把输入当作命名标识符切分成单词：

- 非字母数字字符一律视为分隔符（见下文"字符工具"）
- 按小驼峰边界切分：`helloWorld` → `hello` / `World`
- 连续大写归为一个缩写词，直到遇到"大写后跟小写"的边界：`HTTPStatusCode` →
  `HTTP` / `Status` / `Code`，`ParseURLDoParse` → `Parse` / `URL` / `Do` / `Parse`
- 字母与数字的边界也会切开：`numbers123test` → `numbers` / `123` / `test`
- `noSplit` 可变参数列出不应被拆分的词（按前缀匹配保护），如
  `Split("parseHTTPURL", "HTTPURL")` → `[parse HTTPURL]`
- 输入为空串或非合法 UTF-8 时，返回只含原输入的单元素切片（`Split("")` → `[""]`）

### 字符工具（utils.go）

| 函数 | 签名 | 说明 |
|------|------|------|
| `ReplaceNonAlphanumeric` | `func(s string, replacement string) string` | 把所有非英文字母/数字的连续字符替换为 `replacement`；`replacement` 为空时使用 `_` |
| `SplitByNonAlphanumeric` | `func(input string) []string` | 按非字母数字字符切分（等价于把它们替换为空格后按空白切分），空段丢弃 |
| `SplitAndKeepDelimiters` | `func(input string) []string` | 切分但保留分隔符：单词与分隔符交替出现在结果中 |
| `ContainsFn` | `func[T any](slice []T, value T, predicate func(got, want T) bool) bool` | 泛型包含判断：对每个元素调用 `predicate(item, value)`，任一为 `true` 即返回 `true` |

```go
stringcase.ReplaceNonAlphanumeric("a-b c", "-")  // a-b-c
stringcase.ReplaceNonAlphanumeric("a-b c", "")   // a_b_c（空替换符回退为 _）
stringcase.SplitByNonAlphanumeric("hello_world 42!")   // [hello world 42]
stringcase.SplitAndKeepDelimiters("hello_world 42!")   // [hello _ world   42 !]

stringcase.ContainsFn([]string{"HTTPURL"}, "HTTP",
	func(got, want string) bool { return strings.HasPrefix(got, want) },
) // true（"HTTPURL" 以 "HTTP" 开头）
```

注意 `ContainsFn` 的谓词语义完全由调用方定义：`predicate(item, value)` 中第一个
参数是切片元素、第二个是待查值。上面 `HasPrefix` 的用法即本包内 `noSplit` 分词的
实际用法。

## 背景：常见命名风格

- **小驼峰（lowerCamelCase / camelCase）**：第一个单词首字母小写，后续单词首字母
  大写，单词间无分隔符。例：`myVariableName`。常见于 JavaScript、Java、C# 的变量
  与函数名。
- **大驼峰（UpperCamelCase / PascalCase）**：每个单词（包括第一个）首字母都大写。
  例：`MyVariableName`。常见于 C#、Java、TypeScript 的类名、接口名与类型名。
- **蛇形（snake_case）**：单词间用下划线 `_` 分隔。例：`first_name`。常见于
  Python、Ruby 的变量与函数名，以及数据库列名。
- **烤肉串（kebab-case）**：单词间用连字符 `-` 分隔，又称 `dash-case`、
  `hyphen-case`、`lisp-case`。例：`user-profile`。常见于 URL、文件名和
  HTML/CSS 类名。

## 注意事项

- 所有转换函数都先 `TrimSpace` 再分词；空输入返回空串（`Split` 例外，返回 `[""]`）。
- 单词内除首字母外统一小写（驼峰）或统一按目标风格处理（蛇形/烤肉串），因此
  `ParseURL.DoParse` 与 `ParseUrl.DoParse` 转换结果一致（`ParseUrlDoParse` /
  `parse_url_do_parse`）。
- 前导/尾随分隔符会被忽略：`__leading_underscores` → `LeadingUnderscores`。
- 数字开头的输入保留数字原位：`123NumberPrefix` 转驼峰后仍为 `123NumberPrefix`。
- 非字母数字字符（标点、符号、emoji 等）一律视为分隔符；非 ASCII 字母/数字（如中文）
  会被整体保留为一个词、原样通过转换（`ToSnakeCase("影師天下太平")` →
  `"影師天下太平"`），但命名边界识别（驼峰大小写切换、字母数字边界）只对 ASCII
  生效，中文与英文混排时的切分结果请以实际输出为准。
