# rand

`rand` 包提供两套随机能力：

- **全局函数**：开箱即用，适合简单场景。
- **Randomizer 实例**：可选择随机源和种子策略，适合需要可控行为的场景。

同时提供 `Seeder`（种子生成）和基于 SHA-256 的随机源扩展。

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

	// 2) Randomizer（PCG + UnixNanoSeed）
	r := utilsrand.NewRandomizer(utilsrand.PCGRandType, utilsrand.UnixNanoSeed)
	fmt.Println(r.RangeUint64(100, 200))

	// 3) Zipf 分布
	z := utilsrand.NewZipfRandomizer(utilsrand.UnixNanoSeed, 1.2, 1.0, 100)
	fmt.Println(z.ZipfUint64())
}
```

## 主要能力

### 1) 全局随机函数（`rand.go`）

- 浮点：`Float32()`、`Float64()`（区间 `[0, 1)`）
- 基础范围：`IntN`、`Int32N`、`Int64N`
- 区间随机（闭区间）：
  - `RandomInt` / `RandomInt32` / `RandomInt64`
  - `RandomUint` / `RandomUint32` / `RandomUint64`
  - `RandomDuration`
- 其它：`RandomString`、`RandomChoice`、`Shuffle`
- 权重选择：`WeightedChoice`、`NonWeightedChoice`
- 哈希随机值：`SHA256Value(serverSeed, clientSeed, nonce)`

### 2) `Randomizer`（`randomizer.go`）

构造函数：

```go
r := NewRandomizer(randType, seedType)
```

支持的 `RandType`：

- `PCGRandType`（默认推荐）
- `ChaCha8RandType`
- `SHA256RandType`
- `ZipfRandType`（注意：`NewRandomizer` 对该类型返回 `nil`，请使用 `NewZipfRandomizer`）

支持的 `SeedType` 见下文 `Seeder`。

`Randomizer` 提供与全局函数对应的方法：

- `RangeInt/RangeInt32/RangeInt64`
- `RangeUint/RangeUint32/RangeUint64`
- `RangeFloat32/RangeFloat64`
- `WeightedChoice` / `NonWeightedChoice`
- `RandomString` / `Shuffle`
- `ZipfUint64`

## Seeder（`seeder.go`）

用于生成 `int64` 种子，构造方式：

```go
s := NewSeeder(seedType)
seed := s.Seed()
```

可选 `SeedType`：

- `UnixNanoSeed`：时间戳 + 系统随机 + goroutine 数量扰动
- `MapHashSeed`：基于 `maphash`
- `CryptoRandSeed`：`crypto/rand`，失败时降级到 `UnixNano()`
- `RandomStringSeed`：随机字符串映射到 `int64`，失败时降级到 `UnixNano()`
- `FixedSeed`：固定种子（测试用），需通过 `Seed(manualSeed...)` 传入

`Seeder` 额外方法：

- `CryptoRand32() [32]byte`：生成 32 字节安全随机值（读取失败会 `panic`）

## 边界语义（重要）

### 区间函数

- 全局 `RandomInt/RandomInt32/RandomInt64/RandomUint* / RandomDuration`：当 `min >= max` 返回 `max`。
- `Randomizer.RangeInt`：当 `min >= max` 返回 `min`。
- `Randomizer.RangeInt32/RangeInt64/RangeUint*`：当 `min >= max` 返回 `max`。

### 权重选择

两套 API 在极端输入下语义不完全一致，请按需选择：

- 全局 `WeightedChoice([]int{})` 返回 `-1`；
  - 单元素返回 `0`
  - 所有权重 <= 0 返回 `0`
- `Randomizer.WeightedChoice([]int{})` 返回 `-1`；
  - 单元素且 `<=0` 返回 `-1`
  - 所有权重 <= 0 返回 `-1`
- 全局 `NonWeightedChoice(nil)` 返回 `-1`，空切片 `[]int{}` 返回 `0`
- `Randomizer.NonWeightedChoice([]int{})` 返回 `-1`

## 使用建议

- 业务随机：优先 `PCGRandType + UnixNanoSeed`。
- 需要更稳定、可复现实验：使用 `FixedSeed` 并固定手工种子。
- 需要偏斜分布：使用 `NewZipfRandomizer`。
- 安全敏感场景：优先 `CryptoRandSeed` / `CryptoRand32`。

## 注意事项

- `NewRandomizer(ZipfRandType, ...)` 当前会返回 `nil`，不要直接调用。
- `Randomizer.RandomString(l)` 需保证 `l >= 0`。
- 全局函数与 `Randomizer` 的边界返回值并非完全一致，建议在业务层做统一封装。
