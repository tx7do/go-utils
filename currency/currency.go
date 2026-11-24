// Package currency provides a single entry point for currency amount
// conversion across two unit layers:
//
//  1. ISO minor units (cents): the Exponent table describes the number of
//     decimal places of each currency (CNY/USD: 2; JPY/KRW/VND: 0;
//     KWD/BHD/TND: 3). ToMinor/ToMajor serve display formatting and
//     payment-provider settlement boundaries.
//  2. A fixed-precision ledger unit: ToBase/FromBase convert between major
//     units and integer ledger units with a uniform 4 decimal places
//     (1 major unit = 10000 ledger units), independent of each currency's
//     ISO exponent.
//
// Convention: route every amount conversion through this package. Inline
// magic-number conversions (×100 / ÷100) are how silent rounding bugs and
// fund discrepancies happen.
//
// Unknown or unspecified currency codes (empty, lowercase, or not in the
// tables) fall back to 2 decimal places — the dominant ISO 4217 tier, so a
// wrong fallback is the least costly one. Add newly supported currencies
// directly to the exponent tables.
//
// Beyond raw conversion the package offers:
//   - Money: a typed amount (currency + integer ledger units) with
//     mismatch-safe arithmetic, exact-sum allocation/split and comparison
//     (money.go);
//   - display formatting with thousands grouping, symbols and ISO metadata
//     (format.go / meta.go);
//   - explicit exchange-rate application (rate.go) and directional
//     rounding variants for fee/settlement boundaries.
package currency

import (
	"math"
	"strings"
)

// DefaultCode is the reference currency for callers that need a fallback
// when no currency code is provided.
const DefaultCode = "USD"

// zeroDecimal / threeDecimal list the ISO 4217 currencies whose minor-unit
// exponent differs from the dominant 2-decimal tier. Every currency not
// listed here (including all mainstream 2-decimal currencies) defaults to
// exponent 2.
var (
	zeroDecimal = map[string]struct{}{
		"BIF": {}, "CLP": {}, "DJF": {}, "GNF": {}, "ISK": {}, "JPY": {},
		"KMF": {}, "KRW": {}, "PYG": {}, "RWF": {}, "UGX": {}, "VND": {},
		"VUV": {}, "XAF": {}, "XOF": {}, "XPF": {},
	}
	threeDecimal = map[string]struct{}{
		"BHD": {}, "IQD": {}, "JOD": {}, "KWD": {}, "LYD": {}, "OMR": {}, "TND": {},
	}
)

// Normalize trims surrounding whitespace and upper-cases a currency code.
// An empty input stays empty (meaning "unspecified").
func Normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// NormalizeOr normalizes a currency code and returns fallback when the
// normalized result is empty (unspecified).
func NormalizeOr(code, fallback string) string {
	if c := Normalize(code); c != "" {
		return c
	}
	return fallback
}

// Exponent returns the ISO 4217 decimal exponent of the currency
// (minor unit = 10^-exponent major units). Empty or unknown codes return 2.
func Exponent(code string) int {
	switch Normalize(code) {
	case "":
		return 2
	default:
		if _, ok := zeroDecimal[Normalize(code)]; ok {
			return 0
		}
		if _, ok := threeDecimal[Normalize(code)]; ok {
			return 3
		}
		return 2
	}
}

// Factor returns the conversion factor "1 major unit = N minor units"
// (10^exponent). Empty or unknown codes return 100.
func Factor(code string) float64 {
	return math.Pow10(Exponent(code))
}

// ToMinor converts a major-unit float amount to integer minor units,
// rounding half away from zero (same semantics as math.Round, so negative
// amounts — refunds, adjustments — are symmetric).
//
//	ToMinor(0.3, "USD") = 30
//	ToMinor(120, "JPY") = 120
//	ToMinor(0.125, "KWD") = 125
func ToMinor(amount float64, code string) int64 {
	return int64(math.Round(amount * Factor(code)))
}

// ToMajor converts integer minor units to a major-unit float.
//
//	ToMajor(160, "USD") = 1.6
//	ToMajor(120, "JPY") = 120
func ToMajor(units int64, code string) float64 {
	return float64(units) / Factor(code)
}

// ToMajorInt converts integer minor units to whole major units (truncated
// toward zero), matching the semantics of integer division. For systems
// that keep balances in whole major units but read from a minor-unit source.
func ToMajorInt(units int64, code string) int64 {
	return int64(ToMajor(units, code))
}

// ===== Fixed-precision ledger unit =====
//
// The ledger unit is a uniform fixed-precision representation of major
// units with 4 decimal places (1 major unit = 10000 ledger units),
// deliberately decoupled from each currency's ISO exponent:
//
//   - 2-decimal currencies (CNY/USD...): 1 minor unit = 100 ledger units —
//     sub-minor precision (0.001 / 0.0001) stays exactly representable, so
//     rate splits and proportional allocations never lose precision;
//   - 0-decimal currencies (JPY/KRW): 1 unit = 10000 ledger units, two
//     guard digits;
//   - 3-decimal currencies (KWD): 1 fils = 10 ledger units, still lossless.
//
// The uniform 1e4 factor (instead of per-currency exponent+2) keeps one
// mental model for the ledger: cross-currency aggregation and conversion
// only need the exchange rate, never a per-currency tier lookup. The
// exponent table then only serves display formatting and payment-provider
// settlement boundaries.

// BaseDigits is the number of decimal places the ledger unit keeps below
// one major unit (4: 0.0001 steps).
const BaseDigits = 4

// BaseFactor is the major-unit → ledger-unit conversion factor, always
// 10^BaseDigits = 10000.
const BaseFactor = 10000

// ToBase converts a major-unit float amount to integer ledger units
// (rounding half away from zero). The ledger unit is uniform across
// currencies, so there is no currency parameter.
//
//	ToBase(0.0001) = 1
//	ToBase(1.2345) = 12345
func ToBase(amount float64) int64 {
	return int64(math.Round(amount * BaseFactor))
}

// FromBase converts integer ledger units to a major-unit float.
//
//	FromBase(12345) = 1.2345
func FromBase(units int64) float64 {
	return float64(units) / BaseFactor
}

// FromBaseInt converts integer ledger units to whole major units (truncated
// toward zero), matching ToMajorInt semantics.
func FromBaseInt(units int64) int64 {
	return units / BaseFactor
}

// MinorToBase converts ISO minor units to ledger units
// (× 10^(BaseDigits - ISO exponent)). Used at payment-provider boundaries
// where counterparties settle in minor units. For 2-decimal currencies this
// is always × 100.
func MinorToBase(minor int64, code string) int64 {
	return minor * int64(math.Pow10(BaseDigits-Exponent(code)))
}

// BaseToMinor converts ledger units to ISO minor units
// (÷ 10^(BaseDigits - ISO exponent), truncated toward zero). Only for
// issuing settlement amounts to payment providers; internal accounting
// always keeps full ledger precision.
func BaseToMinor(units int64, code string) int64 {
	return units / int64(math.Pow10(BaseDigits-Exponent(code)))
}
