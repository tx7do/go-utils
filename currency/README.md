# currency — Currency & Ledger Unit Conversion

`currency` is a single, mandatory entry point for converting currency amounts between three representations:

1. **Major units** — the human-facing float amount (e.g. `1.23` USD);
2. **ISO minor units** — integer cents (ISO 4217 exponent table: USD/CNY → 2, JPY/KRW → 0, KWD/BHD → 3);
3. **Ledger units** — a uniform fixed-precision integer accounting unit with **4 decimal places** (1 major unit = `10000` ledger units), independent of each currency's ISO exponent.

> **Convention**: route every amount conversion through this package. Inline magic-number conversions (`×100` / `÷100`) are how silent rounding bugs and fund discrepancies happen.

## Features

- **ISO 4217 exponent table**: zero- and three-decimal currencies are listed explicitly; everything else defaults to the dominant 2-decimal tier.
- **Safe fallback**: unknown, lowercase, or empty codes normalize and fall back to 2 decimals — the least costly wrong answer.
- **Round-half-away-from-zero**: `ToMinor`/`ToBase` use `math.Round` semantics, so negative amounts (refunds, adjustments) are symmetric and don't drift directionally.
- **Uniform ledger unit**: one mental model for accounting — cross-currency aggregation only needs the exchange rate, never a per-currency tier lookup.
- **Boundary bridges**: `MinorToBase` / `BaseToMinor` connect ISO minor-unit counterparties (payment providers) to the full-precision ledger.
- **Zero dependencies**: pure standard library; only tests use `testify`.

## Installation

```bash
go get github.com/tx7do/go-utils/currency
```

```go
import "github.com/tx7do/go-utils/currency"
```

## Quick Start

```go
// ISO minor units (display formatting / payment-provider boundaries)
currency.ToMinor(0.3, "USD")   // 30
currency.ToMinor(120, "JPY")   // 120  (0-decimal)
currency.ToMinor(0.125, "KWD") // 125  (3-decimal)

currency.ToMajor(160, "USD")   // 1.6

// Ledger units (uniform 4-decimal accounting)
currency.ToBase(1.2345)        // 12345
currency.FromBase(12345)       // 1.2345

// Bridges between the two integer layers
currency.MinorToBase(30, "USD")     // 3000  (2 guard digits)
currency.BaseToMinor(3000, "USD")   // 30

// Normalization
currency.Normalize(" usd ")         // "USD"
currency.NormalizeOr("", "USD")     // "USD"
```

## API

| Function | Description |
|---|---|
| `Normalize(code) string` | Trim whitespace, upper-case; empty stays empty. |
| `NormalizeOr(code, fallback string) string` | Normalize; return fallback when unspecified. |
| `Exponent(code) int` | ISO 4217 decimal exponent; empty/unknown → `2`. |
| `Factor(code) float64` | `1 major = 10^exponent minor`; empty/unknown → `100`. |
| `ToMinor(amount float64, code) int64` | Major float → minor int, half away from zero. |
| `ToMajor(units int64, code) float64` | Minor int → major float. |
| `ToMajorInt(units int64, code) int64` | Minor int → whole major units (toward zero). |
| `ToBase(amount float64) int64` | Major float → ledger int (uniform, no currency arg). |
| `FromBase(units int64) float64` | Ledger int → major float. |
| `FromBaseInt(units int64) int64` | Ledger int → whole major units (toward zero). |
| `MinorToBase(minor int64, code) int64` | Minor → ledger (× 10^(4−exponent)). |
| `BaseToMinor(units int64, code) int64` | Ledger → minor (÷ 10^(4−exponent), toward zero). |
| `DefaultCode` | Reference currency constant (`"USD"`) for unspecified fallbacks. |
| `BaseDigits` / `BaseFactor` | Ledger precision constants: `4` / `10000`. |

## Precision Notes

- **Rounding**: conversions *up* (`ToMinor`, `ToBase`) round half away from zero — negative amounts are symmetric. Conversions *down* (`ToMajorInt`, `FromBaseInt`, `BaseToMinor`) truncate toward zero, matching integer-division semantics.
- **Float artifacts**: `0.29 * 100` in raw float math can yield `28.999999...`; `ToMinor` rounds through it. Conversions in the integer domain round-trip losslessly (see `TestBaseRoundTrip`).
- **`BaseToMinor` truncation**: `5` ledger units = `0.5` fils (KWD) → truncates to `0`. Issue only whole minor multiples at boundaries; internal accounting always keeps full ledger precision.

## Extending

Unknown codes default to exponent 2. To support a newly listed ISO 4217 currency in the zero- or three-decimal tiers, add its code to the `zeroDecimal` / `threeDecimal` tables — nothing else needs to change.
