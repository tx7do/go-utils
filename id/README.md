
# ID生成器

本 package 提供多种分布式唯一 ID 生成方案，适用于订单号、分布式主键、设备标识等多种场景。

## 支持的 ID 类型

- **UUID/GUID**：标准 128 位唯一标识符，支持 v1/v4。
- **Snowflake**：高性能分布式时序 ID，支持自定义 WorkerID。
- **KSUID**：K-Sortable Unique ID，严格时序，适合日志等场景。
- **ShortUUID**：短 ID，适合 URL、短链等。
- **XID**：高并发下的短 ID，趋势有序。
- **MachineID**：获取本机唯一标识，支持多平台（Windows、Linux、macOS）。

## 快速开始

以 Go 代码为例：

```go
import "github.com/yourorg/go-utils/id"

// 生成 UUID
id1 := id.NewUUID()

// 生成 Snowflake ID
id2 := id.NewSnowflakeID()

// 生成 KSUID
id3 := id.NewKSUID()

// 生成 ShortUUID
id4 := id.NewShortUUID()

// 生成 XID
id5 := id.NewXID()

// 获取本机唯一标识
mid, err := id.MachineID()
```

## MachineID 说明

MachineID 用于唯一标识一台物理/虚拟机，常用于授权、追踪等场景。

- Windows: 读取注册表 MachineGuid。
- Linux: 读取 /etc/machine-id 或 /var/lib/dbus/machine-id。
- macOS: 读取 IOPlatformUUID（通过 ioreg 命令）。

如需更高健壮性，已实现多路径兼容。

## 自定义格式

部分 ID 支持自定义格式：

- 可选大写/小写输出。
- 可选是否带横线（如 UUID）。
- 可选前缀、后缀等。

示例：

```go
id := id.NewUUIDWithFormat(id.UUIDFormat{Upper: true, Hyphen: false})
```

## 依赖与测试

本库为纯 Go 实现，无第三方依赖。

运行全部单元测试：

```sh
go test ./...
```

---


## 订单ID

- 电商平台：202506041234567890（时间戳 + 随机数，19-20 位）。
- 支付系统：PAY20250604123456789（业务前缀 + 时间戳 + 序号）。
- 微信支付：1589123456789012345（类似 Snowflake 的纯数字 ID）。
- 美团订单：202506041234567890123（时间戳 + 商户 ID + 随机数）。


## UUID

| 特性    | GUID/UUID    | KSUID      | ShortUUID | XID      | Snowflake      |
|-------|--------------|------------|-----------|----------|----------------|
| 长度    | 36/32字符（不含-） | 27字符       | 22字符      | 20字符     | 19（数字位数）       |
| 有序性   | 无序（UUIDv4）   | 严格时序       | 无序        | 趋势有序     | 严格时序           |
| 时间精度  | 无（UUIDv4）    | 毫秒级        | 无         | 秒级       | 毫秒级            |
| 分布式安全 | 高（随机数）       | 高          | 高         | 高        | 高（需配置WorkerID） |
| 性能    | 中等           | 中等         | 较低（编码开销）  | 极高       | 极高             |
| 时钟依赖  | 无            | 有（需处理时钟回拨） | 无         | 有（但影响较小） | 强依赖（需严格同步）     |
| 适用场景  | 跨系统兼容        | 时序索引       | 短ID、URL   | 高并发、短ID  | 分布式时序ID        |

### 选择建议

- **GUID/UUID**: 适用于需要跨系统兼容的场景，特别是当不需要有序性时。
- **KSUID**: 适合需要严格时序的应用，如事件日志、时间序列数据。
- **ShortUUID**: 当需要短ID且不关心有序性时的理想选择，适用于URL、短链接等。
- **XID**: 高并发场景下的短ID选择，适合需要一定有序性的应用。
- **Snowflake**: 适合分布式系统，特别是需要严格时序和高性能的场景，如大规模分布式应用。
