# id

[English](README.md) | 简体中文

`id` 是一个独立 Go 模块（`github.com/tx7do/go-utils/id`），提供多种唯一 ID 生成能力：通用主键、分布式时序 ID、订单号和机器标识。

## 功能特性

- UUID/GUID：`NewGUIDv4`、`NewGUIDv7`（支持是否带横线）
- 短 ID：`NewShortUUID`、`NewKSUID`、`NewXID`
- MongoDB ID：`NewMongoObjectID`
- 分布式 ID：`NewSnowflakeID` / `GenerateSnowflakeID`、`NewSonyflakeID` / `GenerateSonyflakeID`
- Snowflake 节点：`NewSnowflakeNode` 返回可复用的节点实例，支持 `Generate` / `GenerateString`
- 订单号：随机、自增、商户维度、带前缀 Snowflake/Sonyflake
- 机器标识：`UnifyMachineID`、`FormatMachineID`（跨平台，基于内嵌的 `id/machineid` 子包）

## 安装

本包是独立模块（有自己的 `go.mod`），不是根模块 `github.com/tx7do/go-utils` 的子包：

```bash
go get github.com/tx7do/go-utils/id
```

```go
import "github.com/tx7do/go-utils/id"
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/id"
)

func main() {
	guid := id.NewGUIDv4(false)   // 32 位无横线
	guidV7 := id.NewGUIDv7(true)  // 带横线的 UUIDv7（时钟不可靠时回退到 v4）
	shortID := id.NewShortUUID()
	ksuid := id.NewKSUID()
	xid := id.NewXID()
	mongoID := id.NewMongoObjectID()

	// Snowflake（按 workerId 缓存节点，进程内复用）
	sfID, err := id.NewSnowflakeID(1)
	if err != nil {
		log.Fatal(err)
	}

	// Sonyflake（进程内单例）
	sonyID, err := id.NewSonyflakeID()
	if err != nil {
		log.Fatal(err)
	}

	// 机器标识（32 位小写十六进制）
	machineID, err := id.UnifyMachineID()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(guid, guidV7, shortID, ksuid, xid, mongoID, sfID, sonyID, machineID)
}
```

## API 参考

### UUID / 短 ID

```go
func NewGUIDv4(withHyphen bool) string
func NewGUIDv7(withHyphen bool) string   // UUIDv7 生成失败时回退为 v4
func NewShortUUID() string
func NewKSUID() string
func NewXID() string
func NewMongoObjectID() string
```

### Snowflake / Sonyflake

```go
func NewSnowflakeID(workerId int64) (int64, error)
func GenerateSnowflakeID(workerId int64) int64   // 忽略错误，失败返回 0

type SnowflakeNode struct{ /* ... */ }

func NewSnowflakeNode(workerId int64) (*SnowflakeNode, error)
func (sfNode *SnowflakeNode) Generate() int64
func (sfNode *SnowflakeNode) GenerateString() string

func NewSonyflakeID() (uint64, error)            // 进程内共享单个 Sonyflake 实例
func GenerateSonyflakeID() uint64                // 忽略错误，失败返回 0
```

- `NewSnowflakeID` / `GenerateSnowflakeID` 按 `workerId` 在进程内缓存 `SnowflakeNode`（`sync.Map`），同一 `workerId` 复用同一节点
- 需要长期持有独立节点实例（例如自定义序列化场景）时，直接使用 `NewSnowflakeNode`
- workerId 的有效范围是 0–1023（上游 `bwmarrin/snowflake` 的 10 位节点数约束），越界时 `NewSnowflakeNode` 返回错误

### 订单号

```go
func GenerateOrderIdWithRandom(prefix string, tm *time.Time) string          // 前缀 + 14 位时间戳 + 4 位随机数
func GenerateOrderIdWithIncreaseIndex(prefix string, tm *time.Time) string   // 前缀 + 14 位时间戳 + 自增索引
func GenerateOrderIdWithTenantId(tenantID string) string                     // 14 位时间戳 + 5 位商户 ID + 4 位随机数
func GenerateOrderIdWithPrefixSonyflake(prefix string) string
func GenerateOrderIdWithPrefixSnowflake(workerId int64, prefix string) string
```

- `tm` 传 `nil` 时使用当前时间
- 自增索引是进程内的 `uint32` 计数器：从 0 递增，到达 1000 时归零（`atomic.CompareAndSwap`），非持久化
- `GenerateOrderIdWithTenantId`：商户 ID 超过 5 位截断，不足 5 位右侧补 `'0'`

### Machine ID

```go
type FormatOption struct {
	UpperCase  bool // true: 大写，false: 小写
	WithHyphen bool // true: 带横线，false: 不带横线
}

func UnifyMachineID() (string, error)                  // 等价于 FormatMachineID{小写、无横线}
func FormatMachineID(opt FormatOption) (string, error)
```

- `UnifyMachineID` 返回 32 位小写十六进制字符串
- `FormatMachineID` 支持大小写与 GUID 横线格式（8-4-4-4-12）
- 底层返回值清洗后不足 32 位十六进制字符时，自动降级为其 SHA-256 哈希前 32 位，保证输出长度稳定

## 注意事项

- `UnifyMachineID` / `FormatMachineID` 的底层机器 ID 在进程内只读取一次（`sync.Once` 缓存），之后始终返回缓存值；运行期修改系统机器 ID 不会反映到本包
- `Generate*` 系列函数会忽略底层错误（失败返回零值），业务中建议优先使用 `New*` 并处理 `error`
- Snowflake/Sonyflake 依赖单调时钟，多进程部署时请确保各进程使用不同的 `workerId`
- 本包依赖多个成熟第三方库（`google/uuid`、`bwmarrin/snowflake`、`sony/sonyflake`、`segmentio/ksuid`、`rs/xid`、`lithammer/shortuuid`、`mongo-driver` 等）

## 平台支持

机器标识底层依赖 [`id/machineid`](./machineid/README.zh-CN.md) 子包，支持 Windows、Linux、macOS、BSD/Solaris、AIX、z/OS，无需管理员权限。

## 测试

```bash
go test ./...
```

## 许可证

MIT
