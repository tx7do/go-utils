package currency

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// ErrCurrencyMismatch is returned when arithmetic or comparison is applied
// to two Money values of different currencies.
var ErrCurrencyMismatch = errors.New("currency: money values of different currencies")

// Money is an amount of a single currency stored in ledger units
// (1 major unit = 10000, see BaseFactor). All arithmetic runs on integers —
// floats never touch the stored value.
//
// The zero value is a zero amount of currency "" (unspecified); every
// operation normalizes the code, so currency mismatch detection works even
// with unnormalized inputs.
type Money struct {
	code string
	base int64
}

// NewMoney returns a zero-amount Money of the given currency.
func NewMoney(code string) Money {
	return Money{code: NormalizeOr(code, DefaultCode)}
}

// MoneyFromBase builds Money from integer ledger units.
func MoneyFromBase(code string, base int64) Money {
	return Money{code: NormalizeOr(code, DefaultCode), base: base}
}

// MoneyFromMinor builds Money from ISO minor units (e.g. cents).
func MoneyFromMinor(code string, minor int64) Money {
	return MoneyFromBase(code, MinorToBase(minor, code))
}

// MoneyFromMajor builds Money from a major-unit float, rounding half away
// from zero. Prefer the integer constructors when the source amount is
// already discrete.
func MoneyFromMajor(code string, amount float64) Money {
	return MoneyFromBase(code, ToBase(amount))
}

// Code returns the normalized currency code.
func (m Money) Code() string { return m.code }

// Base returns the raw ledger units.
func (m Money) Base() int64 { return m.base }

// Minor returns the amount in ISO minor units (truncated toward zero —
// sub-minor ledger precision is dropped at the ISO boundary).
func (m Money) Minor() int64 { return BaseToMinor(m.base, m.code) }

// Major returns the amount as a major-unit float.
func (m Money) Major() float64 { return FromBase(m.base) }

// IsZero reports whether the amount is exactly zero.
func (m Money) IsZero() bool { return m.base == 0 }

// IsPositive reports whether the amount is greater than zero.
func (m Money) IsPositive() bool { return m.base > 0 }

// IsNegative reports whether the amount is less than zero.
func (m Money) IsNegative() bool { return m.base < 0 }

// Neg returns -m.
func (m Money) Neg() Money { return Money{code: m.code, base: -m.base} }

// Abs returns |m|.
func (m Money) Abs() Money {
	if m.base < 0 {
		return m.Neg()
	}
	return m
}

// sameCurrencyOr panics-free helper: returns ErrCurrencyMismatch unless the
// normalized codes match.
func sameCurrencyOr(a, b Money) error {
	if a.code != b.code {
		return fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, a.code, b.code)
	}
	return nil
}

// Add returns a+b. Both values must share one currency.
func (m Money) Add(o Money) (Money, error) {
	if err := sameCurrencyOr(m, o); err != nil {
		return Money{}, err
	}
	return Money{code: m.code, base: m.base + o.base}, nil
}

// Sub returns a-b. Both values must share one currency.
func (m Money) Sub(o Money) (Money, error) {
	if err := sameCurrencyOr(m, o); err != nil {
		return Money{}, err
	}
	return Money{code: m.code, base: m.base - o.base}, nil
}

// Cmp compares two same-currency amounts: -1/0/+1.
func (m Money) Cmp(o Money) (int, error) {
	if err := sameCurrencyOr(m, o); err != nil {
		return 0, err
	}
	switch {
	case m.base < o.base:
		return -1, nil
	case m.base > o.base:
		return 1, nil
	default:
		return 0, nil
	}
}

// Mul scales the amount by an integer factor (e.g. quantity × unit price).
func (m Money) Mul(factor int64) Money {
	return Money{code: m.code, base: m.base * factor}
}

// Allocate splits the amount by integer ratios so the parts sum exactly to
// the original — the remainder is distributed by the largest-remainder
// rule (per-unit, deterministic). Negative amounts allocate symmetrically.
// Ratios must not all be zero, and negative ratios are rejected.
func (m Money) Allocate(ratios []int64) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, fmt.Errorf("currency: allocate requires at least one ratio")
	}
	total := int64(0)
	for i, r := range ratios {
		if r < 0 {
			return nil, fmt.Errorf("currency: allocate ratio[%d] is negative", i)
		}
		total += r
	}
	if total == 0 {
		return nil, fmt.Errorf("currency: allocate ratios sum to zero")
	}

	// Largest-remainder allocation with exact integer accumulation
	// (big.Int guards against base*ratio overflow).
	acc := big.NewInt(0)
	totalBig := big.NewInt(total)
	parts := make([]int64, len(ratios))
	prev := int64(0)
	for i, r := range ratios {
		acc.Add(acc, big.NewInt(m.base*r))
		target := new(big.Int).Quo(acc, totalBig).Int64() // truncating division
		parts[i] = target - prev
		prev = target
	}
	// Truncation residue (sign follows m.base) goes to the largest ratio.
	residue := m.base - prev
	if residue != 0 {
		maxIdx := 0
		for i, r := range ratios {
			if r > ratios[maxIdx] {
				maxIdx = i
			}
		}
		parts[maxIdx] += residue
	}

	out := make([]Money, len(parts))
	for i, p := range parts {
		out[i] = Money{code: m.code, base: p}
	}
	return out, nil
}

// Split divides the amount into n near-equal parts summing exactly to the
// original: the remainder (first units) goes to the leading parts. n must
// be positive; negative amounts split symmetrically.
func (m Money) Split(n int) ([]Money, error) {
	if n <= 0 {
		return nil, fmt.Errorf("currency: split count must be positive")
	}
	ratios := make([]int64, n)
	for i := range ratios {
		ratios[i] = 1
	}
	return m.Allocate(ratios)
}

// Sum sums same-currency amounts (mismatch → ErrCurrencyMismatch).
func Sum(first Money, rest ...Money) (Money, error) {
	total, err := first.Add(Money{code: first.code})
	if err != nil {
		return Money{}, err
	}
	for _, m := range rest {
		if total, err = total.Add(m); err != nil {
			return Money{}, err
		}
	}
	return total, nil
}

// Min returns the smaller of two same-currency amounts.
func Min(a, b Money) (Money, error) {
	cmp, err := a.Cmp(b)
	if err != nil {
		return Money{}, err
	}
	if cmp < 0 {
		return a, nil
	}
	return b, nil
}

// Max returns the larger of two same-currency amounts.
func Max(a, b Money) (Money, error) {
	cmp, err := a.Cmp(b)
	if err != nil {
		return Money{}, err
	}
	if cmp > 0 {
		return a, nil
	}
	return b, nil
}

// LessThan reports a < b (same currency).
func (m Money) LessThan(o Money) (bool, error) {
	cmp, err := m.Cmp(o)
	return cmp < 0, err
}

// GreaterThan reports a > b (same currency).
func (m Money) GreaterThan(o Money) (bool, error) {
	cmp, err := m.Cmp(o)
	return cmp > 0, err
}

// MulRatio scales the amount by the exact integer ratio num/den, rounding
// half away from zero on ledger units. This is the ledger-grade rate/fee
// application: no floats participate ("2.5%" = MulRatio(25, 1000)).
func (m Money) MulRatio(num, den int64) (Money, error) {
	if den == 0 {
		return Money{}, fmt.Errorf("currency: mul ratio denominator is zero")
	}
	acc := new(big.Int).Mul(big.NewInt(m.base), big.NewInt(num))
	q, r := new(big.Int).QuoRem(acc, big.NewInt(den), new(big.Int))
	// half away from zero: |2r| >= |den| moves away from zero
	twice := new(big.Int).Mul(r, big.NewInt(2))
	if twice.Abs(twice).Cmp(big.NewInt(absI64(den))) >= 0 {
		step := big.NewInt(1)
		if acc.Sign() < 0 {
			step = big.NewInt(-1)
		}
		q.Add(q, step)
	}
	return Money{code: m.code, base: q.Int64()}, nil
}

// Percent returns pct percent of the amount ("2.5%" = Percent(250)).
// Equivalent to MulRatio(pct, 100).
func (m Money) Percent(pct int64) (Money, error) {
	return m.MulRatio(pct, 100)
}

// RoundToMinor rounds away sub-minor ledger dust (half away from zero),
// keeping the currency and the ledger-unit domain: after allocation or
// ratio math, callers often need amounts clean at display precision.
func (m Money) RoundToMinor() Money {
	return MoneyFromMinor(m.code, m.Minor())
}

func absI64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// String renders the amount at display precision followed by the code,
// e.g. "1,234.56 USD", "120 JPY".
func (m Money) String() string {
	return FormatBase(m.base, m.code) + " " + m.code
}

// GoString satisfies fmt for debugging: same as String but quoted.
func (m Money) GoString() string { return fmt.Sprint(m.String()) }

// FormatBase renders integer ledger units at display precision: thousands
// grouping, at least the ISO exponent decimals, extra ledger digits kept
// only when non-zero (e.g. "0.001" for sub-cent USD amounts, "120" for JPY,
// "120.05" for JPY holding sub-yen ledger dust).
func FormatBase(base int64, code string) string {
	exp := Exponent(code)
	neg := base < 0
	v := base
	if neg {
		v = -v
	}
	intPart := v / int64(BaseFactor)
	frac4 := int(v % int64(BaseFactor))

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteString(groupThousands(int64(intPart)))

	if frac4 == 0 {
		if exp > 0 {
			b.WriteByte('.')
			b.WriteString(strings.Repeat("0", exp))
		}
		return b.String()
	}

	digits := fmt.Sprintf("%04d", frac4)
	minLen := exp
	if minLen < 1 {
		minLen = 1
	}
	// 去尾部零但绝不越过非零位（值不变），且不短于指数位精度
	for len(digits) > minLen && digits[len(digits)-1] == '0' {
		digits = digits[:len(digits)-1]
	}
	b.WriteByte('.')
	b.WriteString(digits)
	return b.String()
}

// FormatMinor renders ISO minor units at display precision: thousands
// grouping plus exactly the currency's exponent decimals
// ("1,234.56" for USD, "120" for JPY, "1.234" for KWD).
func FormatMinor(minor int64, code string) string {
	exp := Exponent(code)
	neg := minor < 0
	v := minor
	if neg {
		v = -v
	}
	factor := int64(math.Pow10(exp))

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	if exp == 0 {
		b.WriteString(groupThousands(v))
		return b.String()
	}
	b.WriteString(groupThousands(v / factor))
	frac := v % factor
	b.WriteByte('.')
	b.WriteString(fmt.Sprintf("%0*d", exp, frac))
	return b.String()
}

// FormatSymbolMinor renders like FormatMinor prefixed with the currency
// symbol ("$1,234.56", "¥120"). Alphabetic symbols get a space separator
// ("kr 12.00", "XYZ 12.00"); graphic symbols attach directly.
func FormatSymbolMinor(minor int64, code string) string {
	sym := Symbol(code)
	needsSpace := false
	for _, r := range sym {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			needsSpace = true
			break
		}
	}
	if needsSpace {
		return sym + " " + FormatMinor(minor, code)
	}
	return sym + FormatMinor(minor, code)
}
