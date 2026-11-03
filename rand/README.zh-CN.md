# rand

[English](README.md) | 简体中文

`rand` 是一个独立 Go 模块（`github.com/tx7do/go-utils/rand`），基于 Go 1.25 的 `math/rand/v2` 提供两套随机能力：

- **全局函数**：开箱即用，适合简单场景
- **`Randomizer` 实例**：可选择随机源与种子策略，支持固定种子复现，适合需要可控行为的场景

同时提供 `Seeder`（种子生成）、`Sha256Source`（基于 SHA-256 的随机源）与 O(1) 加权别名表。

## 功能特性

- 四种随机源：PCG（默认）、ChaCha8、SHA-256、Zipf
- 五种种子策略：UnixNano、MapHash、CryptoRand（默认）、RandomString、Fixed
- Options 模式构造：`New(opts ...Option)`
- 丰富的业务方法：区间、字符串、时间、权重、分布（正态/指数/对数正态/截断正态）、游戏骰子、概率命中、抖动等
- O(1) 加权随机：`NewAliasTable*` + `AliasChoice`

## 安装

本包是独立模块（有自己的 `go.mod`），不是根模块 `github.com/tx7do/go-utils` 的子包：

```bash
go get github.com/tx7do/go-utils/rand
```

```go
import utilsrand "github.com/tx7do/go-utils/rand"
```

> 包名为 `rand`，与标准库 `math/rand` 同名，导入时建议使用别名（如 `utilsrand`）。

## 快速开始

```go
package main

import (
	"fmt"

	utilsrand "github.com/tx7do/go-utils/rand"
)

func main() {
	// 1) 全局函数
	fmt.Println(utilsrand.RandomInt(1, 10))
	fmt.Println(utilsrand.RandomString(12))

	// 2) Options 模式（默认 PCG + CryptoRand 种子）
	r := utilsrand.New()
	fmt.Println(r.RangeUint64(100, 200))
	fmt.Println(r.PickString([]string{"a", "b", "c"}))

	// 3) 兼容旧 API：显式指定随机源与种子策略
	r2 := utilsrand.NewRandomizer(utilsrand.PCGRandType, utilsrand.UnixNanoSeed)
	fmt.Println(r2.RangeUint64(100, 200))

	// 4) Zipf 分布（四种随机源都能正常构造，Zipf 型额外启用 Zipf 采样）
	z := utilsrand.NewZipfRandomizer(utilsrand.UnixNanoSeed, 1.2, 1.0, 100)
	fmt.Println(z.ZipfUint64())

	// 5) 固定种子（可复现序列，测试用）
	r3 := utilsrand.New(
		utilsrand.WithRandType(utilsrand.PCGRandType),
		utilsrand.WithFixedSeed(42),
	)
	fmt.Println(r3.IntN(100))
}
```

## 全局函数（rand.go）

- 浮点：`Float32()`、`Float64()`（`[0, 1)`）
- 基础范围：`IntN`、`Int32N`、`Int64N`（`[0, n)`）
- 区间随机（闭区间）：`RandomInt` / `RandomInt32` / `RandomInt64` / `RandomUint` / `RandomUint32` / `RandomUint64` / `RandomDuration`
- 字符串：`RandomString(l)`（大小写字母+数字；`l <= 0` 返回空串）
- 通用选择：`RandomChoice[T](array []T, n int) []T`（随机选 n 个，不修改原数组）
- 洗牌：`Shuffle[T any](array []T)`（原位 Fisher–Yates）
- 权重选择：`WeightedChoice`、`NonWeightedChoice`
- 哈希随机值：`SHA256Value(serverSeed, clientSeed string, nonce uint64) uint64`

## Randomizer（randomizer.go）

### 构造

```go
func New(opts ...Option) *Randomizer                          // Options 模式，默认 PCG + CryptoRand
func Default() *Randomizer                                    // 进程内共享的默认实例
func NewRandomizer(randType RandType, seedType SeedType) *Randomizer       // 兼容旧 API
func NewRandomizerWithSeed(randType RandType, seed uint64) *Randomizer     // 兼容旧 API
func NewZipfRandomizer(seedType SeedType, s, q float64, v uint64) *Randomizer // 兼容旧 API
```

### Options

| 选项 | 说明 |
|------|------|
| `WithRandType(t RandType)` | 随机源类型：`PCGRandType`（默认）/ `ChaCha8RandType` / `SHA256RandType`（测试用，非高性能）/ `ZipfRandType` |
| `WithSeedType(t SeedType)` | 种子策略：`CryptoRandSeed`（默认）/ `UnixNanoSeed` / `MapHashSeed` / `RandomStringSeed` / `FixedSeed` |
| `WithFixedSeed(seed uint64)` | 固定种子，优先级最高（可复现随机序列） |
| `WithZipfParams(s, q float64, v uint64)` | Zipf 分布参数（默认 s=1.1, q=1.0, v=100） |
| `WithSecureMode()` | 加密安全模式：启用后 `SecureInt64` 使用 `crypto/rand` |

> 四种 `RandType`（含 `ZipfRandType`）都能正常构造 `Randomizer`；Zipf 型仅在 `ZipfUint64`/`Uint64` 上使用 Zipf 采样，其余方法与 PCG 型一致。

### 基础方法

- 浮点/整数：`Float32`、`Float64`、`Int`、`Int32`、`Int64`、`Uint32`、`Uint64`、`IntN`、`Int32N`、`Int64N`、`UintN`、`Uint32N`、`Uint64N`
- 区间（闭区间）：`RangeInt`、`RangeInt32`、`RangeInt64`、`RangeUint`、`RangeUint32`、`RangeUint64`、`RangeFloat32`、`RangeFloat64`
- 权重：`WeightedChoice`、`NonWeightedChoice`、`WeightedBool(trueWeight, falseWeight int) bool`
- 字符串：`RandomString(l)`（需 `l >= 0`）、`StringWithCharset(length, charset)`、`Hex(length)`、`LetterString(length)`
- 分布：`Normal(mean, stdDev)`、`Exp(lambda)`、`TruncatedNormal(mean, std, low, high)`、`LogNormal(mean, std)`、`ZipfUint64`、`Perm(n)`
- 其他：`Bool()`、`TimeBetween(min, max time.Time)`、`SecureInt64() int64`、`Shuffle(n int, swap func(i, j int))`

### 常用业务方法（节选）

| 类别 | 方法 |
|------|------|
| 选择 | `Pick([]any) any`、`PickString([]string) string`、`PickInt([]int) int`、`RandomPickUniqueStr/RandomPickUniqueInt/RandomPickUnique(list, count)` |
| 数据模拟 | `UUID()`、`PhoneNumber()`（中国大陆号段）、`Email()`、`IDCard()`（简化版，非真实算法）、`RandomIPv4/IPv6`、`RandomLatLng`、`RandomColorHex/RGB`、`RandomFilePath` |
| 概率 | `ProbabilityHit(p float64) bool`（百分比 0~100）、`ProbPermilleHit(permille int) bool`（千分比 0~1000）、`CritHit(rate)` |
| 抖动 | `JitterDuration(base, jitterPercent)`、`JitterInt(base, percent)`、`IntOffset`、`FloatOffset` |
| 骰子/Roll | `Roll(min, max)`、`D6`、`D10`、`D20`、`D100` |
| 其他 | `RandomSign`、`RandomAngle`、`RandomAngleRad`、`CircleRandomPoint`、`RectRandomPoint`、`AttrAssign` 等 |

### O(1) 加权选择（别名表）

```go
weights := []int{10, 20, 30, 40}
table := r.NewAliasTable(weights)          // 另有 NewAliasTableFloat32 / NewAliasTableFloat64
idx := r.AliasChoice(table)                // O(1) 按权重返回原始索引；表为 nil/空时返回 -1
```

构建时过滤无效权重（int 型 `<= 0`，float 型 `<= 1e-7`），并保留原始索引映射。权重集合不变、需要高频抽样时，别名表比每次 O(n) 的 `WeightedChoice` 更快。

> `Randomizer` 没有 `RandomChoice` 的泛型方法（Go 方法不支持类型参数）；单元素选择用 `Pick/PickString/PickInt`，多元素去重选择用 `RandomPickUnique*`。

## Seeder（seeder.go）

```go
s := utilsrand.NewSeeder(utilsrand.CryptoRandSeed)
seed := s.Seed()                       // int64 种子
seed = s.Seed(12345)                   // FixedSeed 类型时使用手工种子
```

`SeedType`：

- `UnixNanoSeed`：时间戳 ^ crypto/rand ^ goroutine 数量扰动（读取失败降级为纯时间戳）
- `MapHashSeed`：基于 `maphash`
- `CryptoRandSeed`：`crypto/rand`，失败时降级到 `UnixNano()`
- `RandomStringSeed`：随机字符串映射到 `int64`，失败时降级到 `UnixNano()`
- `FixedSeed`：固定种子（测试用），需通过 `Seed(manualSeed...)` 传入

其他：`CryptoRand32() [32]byte`（32 字节安全随机值，读取失败会 panic）、包级函数 `Seed(seedType SeedType) int64`。

## Sha256Source（sha256_source.go）

实现 `math/rand/v2` 的 `Source` 接口，可用于 `mathRand.New(src)`：

```go
src := utilsrand.NewSha256Source(seedBytes) // 每次调用递增 nonce，输出 SHA-256(seed||nonce) 的前 8 字节
```

## 边界语义（重要）

### 区间函数

- 全局 `RandomInt/RandomInt32/RandomInt64/RandomUint*/RandomDuration`：当 `min >= max` 返回 `max`
- `Randomizer.RangeInt`：当 `min >= max` 返回 `min`
- `Randomizer.RangeInt32/RangeInt64/RangeUint*/RangeFloat32/RangeFloat64`：当 `min >= max` 返回 `max`

### 权重选择

两套 API 在极端输入下语义不一致，请按需选择：

- 全局 `WeightedChoice([]int{})` 返回 `-1`
  - 单元素返回 `0`
  - 所有权重 <= 0 返回 `0`
- `Randomizer.WeightedChoice([]int{})` 返回 `-1`
  - 单元素且 `<= 0` 返回 `-1`
  - 所有权重 <= 0 返回 `-1`
- 全局 `NonWeightedChoice(nil)` 返回 `-1`，空切片 `[]int{}` 返回 `0`
- `Randomizer.NonWeightedChoice([]int{})` 返回 `-1`

### Shuffle 签名

- `Randomizer.Shuffle(n int, swap func(i, j int))`：与 `math/rand` 的 `Shuffle` 一致，通过回调交换
- 全局 `Shuffle[T any](array []T)`：直接原位打乱切片

两者签名不同，不要混用。

## 使用建议

- 业务随机：默认 `New()`（PCG + CryptoRand）即可
- 需要稳定可复现的序列：`WithFixedSeed`
- 需要偏斜分布：`WithRandType(ZipfRandType)` 或 `NewZipfRandomizer`
- 密码级随机整数：`WithSecureMode()` + `SecureInt64()`（未开启安全模式时 `SecureInt64` 退化为普通 `Int64`）
- 高频加权抽样：`NewAliasTable*` + `AliasChoice`

## 注意事项

- 包名与标准库 `math/rand` 同名，导入请使用别名
- `Randomizer.RandomString(l)` 需保证 `l >= 0`（负值会 panic）；全局 `RandomString` 对 `l <= 0` 返回空串
- 全局函数与 `Randomizer` 的边界返回值并非完全一致，建议在业务层做统一封装
- `Randomizer` 实例（包括 `Default()` 返回的共享实例）内部直接包装 `math/rand/v2` 的 `*Rand`，本身不加锁；多 goroutine 并发调用同一实例需外部同步（如加互斥锁），或为每个 goroutine 分配独立实例。全局函数使用带锁的默认源，可安全并发使用
