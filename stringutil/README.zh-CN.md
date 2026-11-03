# stringutil

字符串实用工具包：类型转换、伪随机/加密随机字符串生成、JSON 字段值替换。

[English](README.md) | 简体中文

## 特性

- **类型转换**：Go 基本类型与字符串互转（含指针版本与带默认值版本）
- **伪随机字符串**：基于 `math/rand/v2` 的 PCG 伪随机数生成器，适合一般业务场景
- **加密随机字符串**：基于 `crypto/rand` 的密码学安全随机字符串，适合密码、Token、密钥
- **JSON 字段替换**：用正则替换 JSON 字符串中指定字段的值
- **可复现随机**：`RandomSeed` 接受外部 `*rand.Rand`，固定种子即可复现输出

## 安装

stringutil 是根模块 `github.com/tx7do/go-utils` 的子包（无独立 go.mod），安装根模块
即可使用：

```bash
go get github.com/tx7do/go-utils
```

```go
import "github.com/tx7do/go-utils/stringutil"
```

> 版本要求：代码使用了 `math/rand/v2`（Go 1.22 引入），实际至少需要 Go 1.22+；
> 根模块 `go.mod` 当前声明 `go 1.25.0`。

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/stringutil"
)

func main() {
	// 类型转换
	s := stringutil.IntToString(42)
	n, _ := stringutil.StringToInt("42")
	fmt.Println(s, n)

	// 伪随机字符串
	alpha, _ := stringutil.RandomAlphabetic(10)
	num, _ := stringutil.RandomNumeric(6)
	alphanum, _ := stringutil.RandomAlphaNumeric(16)
	fmt.Println(alpha, num, alphanum)

	// 加密安全随机字符串
	safe, _ := stringutil.CryptoRandomAlphaNumeric(32)
	fmt.Println(safe)

	// JSON 字段替换
	jsonStr := `{"tenantId": "old_value", "name": "test"}`
	result := stringutil.ReplaceJSONField("tenantId|tenant_id", "new_value", jsonStr)
	fmt.Println(result)
}
```

## API 参考

### 类型转换（convert.go）

#### 基本类型 → 字符串

| 函数 | 说明 |
|------|------|
| `BoolToString(v bool)` | `true` → `"true"` |
| `IntToString(v int)` | 整数转十进制字符串 |
| `Int8ToString` / `Int16ToString` / `Int32ToString` / `Int64ToString` | 各尺寸有符号整数 |
| `UintToString` / `Uint8ToString` / `Uint16ToString` / `Uint32ToString` / `Uint64ToString` | 各尺寸无符号整数 |
| `Float32ToString` / `Float64ToString` | 浮点数转字符串（`DoubleToString` 是 `Float64ToString` 的别名） |

#### 字符串 → 基本类型

所有函数会自动 `TrimSpace`，解析失败时返回 `error`。

| 函数 | 说明 |
|------|------|
| `StringToBool(s)` | 基于 `strconv.ParseBool`，接受的取值：`1`、`t`、`T`、`TRUE`、`true`、`True` → `true`；`0`、`f`、`F`、`FALSE`、`false`、`False` → `false`（其余报错） |
| `StringToInt(s)` / `StringToIntBase(s, base)` | 十进制或指定进制（base 2-36，0 表示按前缀推断） |
| `StringToInt8` ~ `StringToInt64` | 各尺寸有符号整数（均有 `*Base` 变体） |
| `StringToUint` ~ `StringToUint64` | 各尺寸无符号整数（均有 `*Base` 变体） |
| `StringToFloat32` / `StringToFloat64` | 浮点数解析（`StringToDouble` 是 `StringToFloat64` 的别名） |

#### 带默认值版本

解析失败时返回 `defaultValue` 而非 `error`：

```go
v := stringutil.StringToIntOrDefault("abc", 0)   // v == 0
v2 := stringutil.StringToIntOrDefault("123", 0)  // v2 == 123
```

可用的函数：`StringToBoolOrDefault`、`StringToIntOrDefault`、`StringToUintOrDefault`、
`StringToDoubleOrDefault`（等价于 `StringToFloat64OrDefault` 语义）等，以及对应的
`*BaseOrDefault` 变体。

#### 指针类型转换

`*T` ↔ `*string` 的转换，输入为 `nil` 时返回 `nil`：

```go
n := 42
sp := stringutil.IntPtrToStringPtr(&n) // *string

s := "100"
np := stringutil.StringPtrToIntPtr(&s) // *int，解析失败返回 nil
```

| 函数系列 | 说明 |
|---------|------|
| `BoolPtrToStringPtr` ~ `DoublePtrToStringPtr` | `*T` → `*string` |
| `StringPtrToBoolPtr` ~ `StringPtrToDoublePtr` | `*string` → `*T`（均有 `*Base` 变体） |

### 伪随机字符串（randomstringutils.go）

基于 `math/rand/v2` 的 PCG 伪随机数生成器，适合一般业务场景。

包级随机源（以当前 Unix 纳秒时间戳播种，`Random*` 系列函数均使用它）：

```go
var RANDOM = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
```

#### 便捷函数

| 函数 | 字符集 | 示例输出 |
|------|--------|---------|
| `RandomAlphabetic(count)` | a-z, A-Z | `KxTmBqRp` |
| `RandomNumeric(count)` | 0-9 | `83472615` |
| `RandomAlphaNumeric(count)` | a-z, A-Z, 0-9 | `aB3kR9xM` |
| `RandomAscii(count)` | ASCII 32-126 | `H_I;E` |
| `RandomNonAlphaNumeric(count)` | 无过滤 | 可包含任意 Unicode 字符（含代理对区） |
| `RandomAlphaNumericCustom(count, letters, numbers)` | 自定义过滤 | — |

#### 通用函数

```go
import "math/rand/v2"

// 使用预定义字符集
out, err := stringutil.Random(10, 0, 0, false, false, 'a', 'b', 'c', '1', '2', '3')

// 可复现的随机字符串（固定种子）：需要自行导入 math/rand/v2
random := rand.New(rand.NewPCG(42, 0))
out, err := stringutil.RandomSeed(10, 0, 0, true, true, nil, random)
```

`Random` 与 `RandomSeed` 的参数：`count` 生成长度；`start` / `end` 字符区间
（含头不含尾，两者均为 0 且未提供 `chars` 时忽略）；`letters` / `numbers` 字符过滤；
`chars` 可选自定义字符集（提供时 `start`/`end` 为该切片的索引区间）。
`start == 0 && end == 0` 且无 `chars` 时，未开启过滤默认区间 `[0, MaxInt32)`，
开启任意过滤则为 `[' ', 'z'+1)`。

#### 错误条件

`RandomSeed` 在以下情况返回 `error`：

- `count < 0`：请求长度为负数
- `chars` 非 `nil` 但长度为 0：空字符集
- `end <= start`（当 `start`/`end` 被显式提供时）
- `end > len(chars)`（当 `chars` 非 `nil` 时）

### 加密安全随机字符串（cryptorandomstringutils.go）

基于 `crypto/rand` 的密码学安全随机数生成器，适合生成密码、Token、密钥等安全敏感场景。

API 与伪随机版本完全对应，函数名以 `Crypto` 为前缀：

| 函数 | 字符集 |
|------|--------|
| `CryptoRandomAlphabetic(count)` | a-z, A-Z |
| `CryptoRandomNumeric(count)` | 0-9 |
| `CryptoRandomAlphaNumeric(count)` | a-z, A-Z, 0-9 |
| `CryptoRandomAscii(count)` | ASCII 32-126 |
| `CryptoRandomNonAlphaNumeric(count)` | 无过滤 |
| `CryptoRandomAlphaNumericCustom(count, letters, numbers)` | 自定义过滤 |
| `CryptoRandom(count, start, end, letters, numbers, chars...)` | 通用接口 |

错误条件与伪随机版本相同（熵源失败除外，见注意事项）。

### JSON 字段替换（replace.go）

```go
func ReplaceJSONField(fieldNames, newValue, jsonStr string) string
```

使用正则表达式替换 JSON 字符串中指定字段的值。

- `fieldNames`：要替换的字段名，用竖线 `|` 分隔（例如 `"tenantId|tenant_id"`）
- `newValue`：新值（将被包裹在引号中）
- 匹配不区分大小写

```go
jsonStr := `{"tenantId": "old", "Name": "test"}`
result := stringutil.ReplaceJSONField("tenantId|tenant_id", "new", jsonStr)
// result: `{"tenantId": "new", "Name": "test"}`
```

## 使用建议

- 一般业务随机：使用 `Random*` 系列函数（性能更好）
- 密码、Token、密钥等安全敏感场景：使用 `CryptoRandom*` 系列函数
- 需要可复现的随机输出：使用 `RandomSeed` 并传入固定种子的 `*rand.Rand`
- 类型转换需要容错：优先使用 `*OrDefault` 系列函数

## 注意事项

- **`CryptoRandom*` 系列在熵源失败时会 `panic`**：内部 `getCryptoRandomInt` 调用
  `rand.Int(rand.Reader, ...)` 失败时直接 panic（`crypto/rand` 读取系统熵失败极其
  罕见，但意味着该错误不会被作为 `error` 返回）。
- 包级变量 `RANDOM` 是全局共享的 `*rand.Rand`，`Random*` 系列函数并发调用时会共用
  同一随机源。`math/rand/v2` 的 `*rand.Rand` 本身**不是并发安全的**，多 goroutine
  同时调用存在数据竞争风险；高并发或需要独立随机序列时，请改用 `RandomSeed` 传入
  自有的随机源并自行加锁隔离。
- `RandomNonAlphaNumeric` / `CryptoRandomNonAlphaNumeric` 不过滤字符，输出可能包含
  任意 Unicode 字符（含部分代理对区间），写入日志或存储前请自行确认字符集安全性。
- `ReplaceJSONField` 基于正则实现，只适合处理"字段值为简单字符串"的 JSON；嵌套结构、
  转义字符复杂时请使用标准 JSON 库。
