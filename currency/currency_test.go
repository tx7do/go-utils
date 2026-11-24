package currency

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert.Equal(t, "USD", Normalize("usd"))
	assert.Equal(t, "JPY", Normalize("  jpy "))
	assert.Empty(t, Normalize(""), "unspecified stays empty")
	assert.Empty(t, Normalize("   "))
}

func TestNormalizeOr(t *testing.T) {
	assert.Equal(t, "USD", NormalizeOr("", "USD"))
	assert.Equal(t, "EUR", NormalizeOr("eur ", "USD"))
}

func TestExponent(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{"", 2},          // unspecified → 2 (default tier)
		{"USD", 2},       // mainstream 2-decimal, not in tables → default
		{"CNY", 2},       //
		{"EUR", 2},       //
		{"JPY", 0},       // zero-decimal tier
		{"KRW", 0},       //
		{"VND", 0},       //
		{"CLP", 0},       //
		{"KWD", 3},       // three-decimal tier
		{"BHD", 3},       //
		{"OMR", 3},       //
		{"usd", 2},       // case normalization
		{" jpy ", 0},     // whitespace normalization
		{"XXX", 2},       // unknown code → 2
		{"CUSTOMCOIN", 2},// unlisted code → 2
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Exponent(c.code), "Exponent(%q)", c.code)
	}
}

func TestFactor(t *testing.T) {
	assert.Equal(t, 100.0, Factor(""))
	assert.Equal(t, 100.0, Factor("USD"))
	assert.Equal(t, 1.0, Factor("JPY"))
	assert.Equal(t, 1000.0, Factor("KWD"))
}

func TestToMinor(t *testing.T) {
	// 2-decimal: 0.3 major = 30 minor; float artifacts must be absorbed by
	// Round (the classic 0.29 truncation bug)
	assert.Equal(t, int64(30), ToMinor(0.3, "USD"))
	assert.Equal(t, int64(29), ToMinor(0.29, "USD"))
	assert.Equal(t, int64(10000), ToMinor(100, "CNY"))
	// 0-decimal: 1 unit stays 1
	assert.Equal(t, int64(120), ToMinor(120, "JPY"))
	assert.Equal(t, int64(7), ToMinor(7.3, "KRW"))
	// 3-decimal
	assert.Equal(t, int64(125), ToMinor(0.125, "KWD"))
	// negatives (refunds): symmetric, half away from zero
	assert.Equal(t, int64(-30), ToMinor(-0.3, "USD"))
	assert.Equal(t, int64(-1), ToMinor(-0.005, "USD"))
}

func TestToMajor(t *testing.T) {
	assert.Equal(t, 1.6, ToMajor(160, "USD"))
	assert.Equal(t, 120.0, ToMajor(120, "JPY"))
	assert.Equal(t, 0.125, ToMajor(125, "KWD"))
	assert.InDelta(t, 9989.76, ToMajor(998976, "USD"), 1e-9)
}

func TestToMajorInt(t *testing.T) {
	// matches integer division (truncate toward zero)
	assert.Equal(t, int64(99), ToMajorInt(9999, "USD"))
	assert.Equal(t, int64(120), ToMajorInt(120, "JPY"))
	assert.Equal(t, int64(-1), ToMajorInt(-150, "USD"))
}

// TestToMinorToMajorRoundTrip minor → major → minor must be lossless
// (the foundation of amount conservation).
func TestToMinorToMajorRoundTrip(t *testing.T) {
	for _, code := range []string{"USD", "JPY", "KWD"} {
		for _, units := range []int64{1, 29, 100, 998976, 1 << 40} {
			assert.Equal(t, units, ToMinor(ToMajor(units, code), code), "round trip %s %d", code, units)
		}
	}
}

// ===== Fixed-precision ledger unit (4 decimals, factor 1e4) =====

func TestToBase(t *testing.T) {
	// full 4-decimal precision
	assert.Equal(t, int64(1), ToBase(0.0001))
	assert.Equal(t, int64(12345), ToBase(1.2345))
	assert.Equal(t, int64(30), ToBase(0.003))
	assert.Equal(t, int64(10000), ToBase(1))
	// float artifacts absorbed by Round
	assert.Equal(t, int64(2999), ToBase(0.2999-1e-13))
	// negatives: symmetric, half away from zero
	assert.Equal(t, int64(-1), ToBase(-0.0001))
	assert.Equal(t, int64(-5000), ToBase(-0.5))
}

func TestFromBase(t *testing.T) {
	assert.Equal(t, 1.2345, FromBase(12345))
	assert.Equal(t, 0.0001, FromBase(1))
}

// TestBaseRoundTrip ledger → major → ledger must be lossless.
func TestBaseRoundTrip(t *testing.T) {
	for _, units := range []int64{1, 29, 100, 12345, 99897600, 1 << 40} {
		assert.Equal(t, units, ToBase(FromBase(units)), "round trip %d", units)
	}
}

func TestFromBaseInt(t *testing.T) {
	// truncate toward zero, same semantics as ToMajorInt
	assert.Equal(t, int64(1), FromBaseInt(12345))
	assert.Equal(t, int64(0), FromBaseInt(9999))
	assert.Equal(t, int64(0), FromBaseInt(-1))
}

func TestMinorBaseConversion(t *testing.T) {
	// 2-decimal: minor ↔ ledger = ×100 (2 guard digits)
	assert.Equal(t, int64(3000), MinorToBase(30, "USD"))
	assert.Equal(t, int64(30), BaseToMinor(3000, "CNY"))
	// 0-decimal (JPY): unit ↔ ledger = ×1e4
	assert.Equal(t, int64(1200000), MinorToBase(120, "JPY"), "120 units = 120e4 ledger")
	assert.Equal(t, int64(120), BaseToMinor(1200000, "JPY"))
	// 3-decimal (KWD): fils ↔ ledger = ×10
	assert.Equal(t, int64(10), MinorToBase(1, "KWD"), "1 fils = 10 ledger")
	assert.Equal(t, int64(1), BaseToMinor(10, "KWD"))
	// BaseToMinor truncation: 5 ledger = 0.5 fils, dropped when issued to a
	// minor-unit boundary (internal accounting never passes through here)
	assert.Equal(t, int64(0), BaseToMinor(5, "KWD"))
	// lossless round trip (ledger → minor → ledger, whole minor multiples)
	for _, units := range []int64{100, 1000, 3000, 1200000} {
		assert.Equal(t, units, MinorToBase(BaseToMinor(units, "USD"), "USD"), "round trip USD %d", units)
	}
}
