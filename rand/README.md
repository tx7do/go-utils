# rand

English | [简体中文](README.zh-CN.md)

`rand` is a standalone Go module (`github.com/tx7do/go-utils/rand`) built on Go 1.25's `math/rand/v2`, providing two layers of randomness:

- **Global functions**: ready to use, best for simple cases
- **`Randomizer` instances**: choose the random source and seeding strategy, with fixed-seed reproducibility for controlled behavior

It also provides `Seeder` (seed generation), `Sha256Source` (a SHA-256-based random source), and an O(1) weighted alias table.

## Features

- Four random sources: PCG (default), ChaCha8, SHA-256, Zipf
- Five seeding strategies: UnixNano, MapHash, CryptoRand (default), RandomString, Fixed
- Options-style construction: `New(opts ...Option)`
- Rich business helpers: ranges, strings, times, weights, distributions (normal/exponential/log-normal/truncated normal), game dice, probability hits, jitter, and more
- O(1) weighted sampling: `NewAliasTable*` + `AliasChoice`

## Installation

This package is a standalone module (it has its own `go.mod`), not a subpackage of the root module `github.com/tx7do/go-utils`:

```bash
go get github.com/tx7do/go-utils/rand
```

```go
import utilsrand "github.com/tx7do/go-utils/rand"
```

> The package name is `rand`, which collides with the standard library `math/rand`; import it with an alias (e.g. `utilsrand`).

## Quick Start

```go
package main

import (
	"fmt"

	utilsrand "github.com/tx7do/go-utils/rand"
)

func main() {
	// 1) Global functions
	fmt.Println(utilsrand.RandomInt(1, 10))
	fmt.Println(utilsrand.RandomString(12))

	// 2) Options style (defaults: PCG + CryptoRand seed)
	r := utilsrand.New()
	fmt.Println(r.RangeUint64(100, 200))
	fmt.Println(r.PickString([]string{"a", "b", "c"}))

	// 3) Legacy API: explicit random source and seeding strategy
	r2 := utilsrand.NewRandomizer(utilsrand.PCGRandType, utilsrand.UnixNanoSeed)
	fmt.Println(r2.RangeUint64(100, 200))

	// 4) Zipf distribution (all four sources construct fine; the Zipf type additionally enables Zipf sampling)
	z := utilsrand.NewZipfRandomizer(utilsrand.UnixNanoSeed, 1.2, 1.0, 100)
	fmt.Println(z.ZipfUint64())

	// 5) Fixed seed (reproducible sequence, useful in tests)
	r3 := utilsrand.New(
		utilsrand.WithRandType(utilsrand.PCGRandType),
		utilsrand.WithFixedSeed(42),
	)
	fmt.Println(r3.IntN(100))
}
```

## Global Functions (rand.go)

- Floats: `Float32()`, `Float64()` (`[0, 1)`)
- Basic ranges: `IntN`, `Int32N`, `Int64N` (`[0, n)`)
- Closed-interval randoms: `RandomInt` / `RandomInt32` / `RandomInt64` / `RandomUint` / `RandomUint32` / `RandomUint64` / `RandomDuration`
- Strings: `RandomString(l)` (letters + digits; returns "" for `l <= 0`)
- Generic picking: `RandomChoice[T](array []T, n int) []T` (picks n elements without modifying the input)
- Shuffling: `Shuffle[T any](array []T)` (in-place Fisher–Yates)
- Weighted picks: `WeightedChoice`, `NonWeightedChoice`
- Hash-based value: `SHA256Value(serverSeed, clientSeed string, nonce uint64) uint64`

## Randomizer (randomizer.go)

### Construction

```go
func New(opts ...Option) *Randomizer                          // Options style; defaults PCG + CryptoRand
func Default() *Randomizer                                    // process-wide shared default instance
func NewRandomizer(randType RandType, seedType SeedType) *Randomizer       // legacy API
func NewRandomizerWithSeed(randType RandType, seed uint64) *Randomizer     // legacy API
func NewZipfRandomizer(seedType SeedType, s, q float64, v uint64) *Randomizer // legacy API
```

### Options

| Option | Description |
|--------|-------------|
| `WithRandType(t RandType)` | Random source: `PCGRandType` (default) / `ChaCha8RandType` / `SHA256RandType` (for tests, not high performance) / `ZipfRandType` |
| `WithSeedType(t SeedType)` | Seeding strategy: `CryptoRandSeed` (default) / `UnixNanoSeed` / `MapHashSeed` / `RandomStringSeed` / `FixedSeed` |
| `WithFixedSeed(seed uint64)` | Fixed seed, highest priority (reproducible sequences) |
| `WithZipfParams(s, q float64, v uint64)` | Zipf parameters (defaults s=1.1, q=1.0, v=100) |
| `WithSecureMode()` | Cryptographically secure mode: `SecureInt64` then uses `crypto/rand` |

> All four `RandType` values (including `ZipfRandType`) construct a valid `Randomizer`; the Zipf type only uses Zipf sampling in `ZipfUint64`/`Uint64`, while all other methods behave like the PCG type.

### Basic Methods

- Floats/integers: `Float32`, `Float64`, `Int`, `Int32`, `Int64`, `Uint32`, `Uint64`, `IntN`, `Int32N`, `Int64N`, `UintN`, `Uint32N`, `Uint64N`
- Closed intervals: `RangeInt`, `RangeInt32`, `RangeInt64`, `RangeUint`, `RangeUint32`, `RangeUint64`, `RangeFloat32`, `RangeFloat64`
- Weights: `WeightedChoice`, `NonWeightedChoice`, `WeightedBool(trueWeight, falseWeight int) bool`
- Strings: `RandomString(l)` (requires `l >= 0`), `StringWithCharset(length, charset)`, `Hex(length)`, `LetterString(length)`
- Distributions: `Normal(mean, stdDev)`, `Exp(lambda)`, `TruncatedNormal(mean, std, low, high)`, `LogNormal(mean, std)`, `ZipfUint64`, `Perm(n)`
- Others: `Bool()`, `TimeBetween(min, max time.Time)`, `SecureInt64() int64`, `Shuffle(n int, swap func(i, j int))`

### Representative Business Methods (selected)

| Category | Methods |
|----------|---------|
| Picking | `Pick([]any) any`, `PickString([]string) string`, `PickInt([]int) int`, `RandomPickUniqueStr/RandomPickUniqueInt/RandomPickUnique(list, count)` |
| Fake data | `UUID()`, `PhoneNumber()` (mainland China prefixes), `Email()`, `IDCard()` (simplified, not a real algorithm), `RandomIPv4/IPv6`, `RandomLatLng`, `RandomColorHex/RGB`, `RandomFilePath` |
| Probability | `ProbabilityHit(p float64) bool` (percent 0–100), `ProbPermilleHit(permille int) bool` (permille 0–1000), `CritHit(rate)` |
| Jitter | `JitterDuration(base, jitterPercent)`, `JitterInt(base, percent)`, `IntOffset`, `FloatOffset` |
| Dice/Roll | `Roll(min, max)`, `D6`, `D10`, `D20`, `D100` |
| Others | `RandomSign`, `RandomAngle`, `RandomAngleRad`, `CircleRandomPoint`, `RectRandomPoint`, `AttrAssign`, etc. |

### O(1) Weighted Picking (Alias Table)

```go
weights := []int{10, 20, 30, 40}
table := r.NewAliasTable(weights)          // also NewAliasTableFloat32 / NewAliasTableFloat64
idx := r.AliasChoice(table)                // O(1) weighted pick returning the original index; -1 for nil/empty tables
```

Table construction filters invalid weights (int `<= 0`, float `<= 1e-7`) and keeps the original index mapping. When the weight set is static and sampling is frequent, the alias table outperforms the O(n)-per-call `WeightedChoice`.

> `Randomizer` has no generic `RandomChoice` method (Go methods cannot declare type parameters); use `Pick/PickString/PickInt` for single picks and `RandomPickUnique*` for multi-picks.

## Seeder (seeder.go)

```go
s := utilsrand.NewSeeder(utilsrand.CryptoRandSeed)
seed := s.Seed()                       // int64 seed
seed = s.Seed(12345)                   // with FixedSeed, uses the manual seed
```

`SeedType`:

- `UnixNanoSeed`: timestamp ^ crypto/rand ^ goroutine count jitter (falls back to the timestamp alone if the random read fails)
- `MapHashSeed`: based on `maphash`
- `CryptoRandSeed`: `crypto/rand`, falling back to `UnixNano()` on failure
- `RandomStringSeed`: a random string mapped to `int64`, falling back to `UnixNano()` on failure
- `FixedSeed`: fixed seed (for tests), passed via `Seed(manualSeed...)`

Others: `CryptoRand32() [32]byte` (a 32-byte secure random value; panics if the read fails) and the package-level `Seed(seedType SeedType) int64`.

## Sha256Source (sha256_source.go)

Implements the `math/rand/v2` `Source` interface, usable with `mathRand.New(src)`:

```go
src := utilsrand.NewSha256Source(seedBytes) // increments a nonce per call; outputs the first 8 bytes of SHA-256(seed||nonce)
```

## Edge Semantics (Important)

### Interval Functions

- Global `RandomInt/RandomInt32/RandomInt64/RandomUint*/RandomDuration`: return `max` when `min >= max`.
- `Randomizer.RangeInt`: returns `min` when `min >= max`.
- `Randomizer.RangeInt32/RangeInt64/RangeUint*/RangeFloat32/RangeFloat64`: return `max` when `min >= max`.

### Weighted Picks

The two APIs differ on degenerate inputs; choose deliberately:

- Global `WeightedChoice([]int{})` returns `-1`;
  - a single element returns `0`
  - all weights <= 0 return `0`
- `Randomizer.WeightedChoice([]int{})` returns `-1`;
  - a single element <= 0 returns `-1`
  - all weights <= 0 return `-1`
- Global `NonWeightedChoice(nil)` returns `-1`; an empty slice `[]int{}` returns `0`
- `Randomizer.NonWeightedChoice([]int{})` returns `-1`

### Shuffle Signatures

- `Randomizer.Shuffle(n int, swap func(i, j int))`: mirrors `math/rand`'s `Shuffle`, swapping via a callback
- Global `Shuffle[T any](array []T)`: shuffles the slice in place

The signatures differ; do not mix them up.

## Recommendations

- General business randomness: plain `New()` (PCG + CryptoRand) is fine
- Stable, reproducible sequences: `WithFixedSeed`
- Skewed distributions: `WithRandType(ZipfRandType)` or `NewZipfRandomizer`
- Cryptographically secure random integers: `WithSecureMode()` + `SecureInt64()` (without secure mode, `SecureInt64` degrades to a plain `Int64`)
- Frequent weighted sampling: `NewAliasTable*` + `AliasChoice`

## Notes

- The package name collides with the standard library `math/rand`; import it with an alias
- `Randomizer.RandomString(l)` requires `l >= 0` (negative values panic); the global `RandomString` returns "" for `l <= 0`
- Edge-case return values are not fully consistent between global functions and `Randomizer`; consider wrapping them uniformly in your business layer
- A `Randomizer` instance (including the shared instance from `Default()`) directly wraps a `math/rand/v2` `*Rand` and is not internally locked; concurrent calls from multiple goroutines require external synchronization (e.g. a mutex) or a separate instance per goroutine. The global functions use a locked default source and are safe for concurrent use
