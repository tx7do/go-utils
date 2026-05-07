
# ID 生成器

`id` package 提供多种唯一 ID 生成能力，覆盖通用主键、分布式时序 ID、订单号和机器标识。

## 功能概览

- UUID/GUID：`NewGUIDv4`、`NewGUIDv7`（支持是否带横线）
- 短 ID：`NewShortUUID`、`NewKSUID`、`NewXID`
- 分布式 ID：`NewSnowflakeID`、`NewSonyflakeID`
- MongoDB ID：`NewMongoObjectID`
- 订单号：随机、自增、商户维度、带前缀 Snowflake/Sonyflake
- 机器标识：`UnifyMachineID`、`FormatMachineID`

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/id"
)

func main() {
	guid := id.NewGUIDv4(false)
	guidV7 := id.NewGUIDv7(true)
	shortID := id.NewShortUUID()
	ksuid := id.NewKSUID()
	xid := id.NewXID()

	sfID, err := id.NewSnowflakeID(1)
	if err != nil {
		log.Fatal(err)
	}

	sonyID, err := id.NewSonyflakeID()
	if err != nil {
		log.Fatal(err)
	}

	machineID, err := id.UnifyMachineID()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(guid, guidV7, shortID, ksuid, xid, sfID, sonyID, machineID)
}
```

## 主要 API

### UUID / 短 ID

```go
func NewGUIDv4(withHyphen bool) string
func NewGUIDv7(withHyphen bool) string
func NewShortUUID() string
func NewKSUID() string
func NewXID() string
func NewMongoObjectID() string
```

### Snowflake / Sonyflake

```go
func NewSnowflakeID(workerId int64) (int64, error)
func GenerateSnowflakeID(workerId int64) int64

func NewSonyflakeID() (uint64, error)
func GenerateSonyflakeID() uint64
```

### 订单号

```go
func GenerateOrderIdWithRandom(prefix string, tm *time.Time) string
func GenerateOrderIdWithIncreaseIndex(prefix string, tm *time.Time) string
func GenerateOrderIdWithTenantId(tenantID string) string
func GenerateOrderIdWithPrefixSonyflake(prefix string) string
func GenerateOrderIdWithPrefixSnowflake(workerId int64, prefix string) string
```

### Machine ID

```go
type FormatOption struct {
	UpperCase  bool
	WithHyphen bool
}

func UnifyMachineID() (string, error)
func FormatMachineID(opt FormatOption) (string, error)
```

- `UnifyMachineID`：返回 32 位小写十六进制字符串
- `FormatMachineID`：支持大小写与 GUID 横线格式（8-4-4-4-12）
- 当底层返回值格式异常时，会自动降级为哈希结果，保证稳定输出

## Machine ID 平台说明

底层依赖 `id/machineid` 子包，支持多平台（Windows、Linux、macOS、BSD、AIX、Z/OS）。

## 说明

- 本包依赖多个成熟第三方库（如 `uuid`、`snowflake`、`sonyflake`、`ksuid` 等）
- `Generate*` 系列方法通常会忽略底层错误，业务中建议优先使用 `New*` 并处理 `error`

## 测试

```sh
go test ./...
```

