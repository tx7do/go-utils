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
| `Money` + constructors/arithmetic | Typed amount (code + integer ledger units): `MoneyFromMinor` etc., `Add`/`Sub`/`Mul`/`Cmp`/`Allocate`/`Split` — mismatch-safe. |
| `FormatMinor(minor, code) string` | Display render: thousands grouping + exponent decimals. |
| `FormatBase(base, code) string` | Ledger-unit render, keeps sub-minor precision only when non-zero. |
| `FormatSymbolMinor(minor, code) string` | Like `FormatMinor` with symbol prefix (`"$1,234.56"`, `"¥120"`). |
| `ConvertMinor(minor, from, to, rate) (int64, error)` | Apply an explicit exchange rate between currencies. |
| `ConvertBase(base, from, to, rate) (int64, error)` | Ledger-unit variant of `ConvertMinor`. |
| `ToMinorFloor/Ceil`, `ToBaseFloor/Ceil` | Directional rounding variants for fee/settlement boundaries. |
| `LookupInfo/Symbol/NumericCode/DisplayName/IsKnown` | ISO 4217 metadata: numeric code, symbol, English name for ~30 currencies. |
| `BaseDigits` / `BaseFactor` | Ledger precision constants: `4` / `10000`. |

## Money Type

`Money` pairs a currency code with an integer ledger-unit amount. Arithmetic
runs on integers only; currency mismatches return `ErrCurrencyMismatch`
instead of silently mixing.

```go
price := currency.MoneyFromMinor("USD", 399) // $3.99
total, err := price.Mul(5)                   // $19.95

parts, err := total.Split(3) // exact-sum split, remainder to leading parts
// parts sum == total — guaranteed in the ledger-unit domain
```

- Constructors: `NewMoney` / `MoneyFromBase` / `MoneyFromMinor` / `MoneyFromMajor`
- Arithmetic: `Add` / `Sub` / `Mul` / `Neg` / `Abs` (mismatch-safe)
- Comparison: `Cmp` / `IsZero` / `IsPositive` / `IsNegative`
- Division: `Allocate(ratios)` largest-remainder exact split; `Split(n)` equal parts
- Display: `String()` → `"1,234.56 USD"`; `FormatBase` / `FormatMinor` /
  `FormatSymbolMinor` with thousands grouping and exponent-aware decimals
  (`"120"` JPY, `"1.234"` KWD, `"0.001"` sub-cent USD)

## Exchange Rates & Metadata

```go
out, err := currency.ConvertMinor(100, "USD", "JPY", 150) // ¥150
out, err = currency.ConvertBase(base, "USD", "JPY", 150)

info := currency.LookupInfo("JPY") // {Numeric:"392", Symbol:"¥", Name:"Japanese Yen"}
currency.Symbol("USD")             // "$"
currency.NumericCode("EUR")        // "978"
currency.DisplayName("jpy")        // "Japanese Yen"
currency.IsKnown("XYZ")            // false (conversions still work, 2-decimal fallback)
```

Rates are explicit floats supplied by the caller — this package never fetches
or caches them. Treat converted values as settlement estimates; ledger-grade
conversion should apply rates in the integer domain.

## Rounding Policies

The default conversions round half away from zero. Directional variants for
fee/settlement boundaries: `ToMinorFloor` / `ToMinorCeil` /
`ToBaseFloor` / `ToBaseCeil`.

## Precision Notes

- **Rounding**: conversions *up* (`ToMinor`, `ToBase`) round half away from zero — negative amounts are symmetric. Conversions *down* (`ToMajorInt`, `FromBaseInt`, `BaseToMinor`) truncate toward zero, matching integer-division semantics.
- **Float artifacts**: `0.29 * 100` in raw float math can yield `28.999999...`; `ToMinor` rounds through it. Conversions in the integer domain round-trip losslessly (see `TestBaseRoundTrip`).
- **`BaseToMinor` truncation**: `5` ledger units = `0.5` fils (KWD) → truncates to `0`. Issue only whole minor multiples at boundaries; internal accounting always keeps full ledger precision.

## Extending

Unknown codes default to exponent 2. To support a newly listed ISO 4217 currency in the zero- or three-decimal tiers, add its code to the `zeroDecimal` / `threeDecimal` tables — nothing else needs to change.
