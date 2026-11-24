package currency

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// groupThousands inserts "," every three digits: 1234567 → "1,234,567".
func groupThousands(v int64) string {
	s := fmt.Sprintf("%d", v)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// ConvertMinor converts an ISO minor-unit amount between currencies using
// an explicit exchange rate: rate expresses how many `to` major units one
// `from` major unit is worth. The result rounds half away from zero.
//
// The rate is a float64 — application code sourcing rates from feeds must
// treat the output as a display/settlement estimate, not ledger truth;
// ledger-grade conversions should pre-scale the rate into exact integers
// and apply them in the ledger-unit domain.
func ConvertMinor(minor int64, from, to string, rate float64) (int64, error) {
	if err := validateRate(rate); err != nil {
		return 0, err
	}
	majorFrom := ToMajor(minor, from)
	return ToMinor(majorFrom*rate, to), nil
}

// ConvertBase converts a ledger-unit amount between currencies (see
// ConvertMinor for the rate semantics and caveats). Sub-minor precision of
// the source amount participates in the computation; the result rounds
// half away from zero to target ledger units.
func ConvertBase(base int64, from, to string, rate float64) (int64, error) {
	if err := validateRate(rate); err != nil {
		return 0, err
	}
	majorFrom := FromBase(base)
	return ToBase(majorFrom * rate), nil
}

// ConvertMinorExact is the ledger-grade variant of ConvertMinor: the rate
// is an exact integer ratio (num/den of `to` major per `from` major), so
// no float participates anywhere. Rounding is half away from zero at the
// target minor unit.
//
//	USD 100¢ → JPY at 150/1  = 15000¢? no: 1.00 USD × 150 = 150 JPY = 150 minor
//	USD 100¢ → KWD at 3/10   = 0.30 KWD = 300 fils
func ConvertMinorExact(minor int64, from, to string, num, den int64) (int64, error) {
	if den == 0 {
		return 0, fmt.Errorf("currency: rate denominator is zero")
	}
	// minor_to = minor_from × num × 10^exp(to) / (den × 10^exp(from))
	acc := new(big.Int).Mul(big.NewInt(minor), big.NewInt(num))
	acc.Mul(acc, big.NewInt(int64(math.Pow10(Exponent(to)))))
	denom := big.NewInt(den)
	denom.Mul(denom, big.NewInt(int64(math.Pow10(Exponent(from)))))
	q, r := new(big.Int).QuoRem(acc, denom, new(big.Int))
	twice := new(big.Int).Mul(r, big.NewInt(2))
	if twice.Abs(twice).Cmp(big.NewInt(absI64D(denom.Int64()))) >= 0 {
		step := big.NewInt(1)
		if acc.Sign() < 0 {
			step = big.NewInt(-1)
		}
		q.Add(q, step)
	}
	return q.Int64(), nil
}

func absI64D(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func validateRate(rate float64) error {
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return fmt.Errorf("currency: rate must be finite, got %v", rate)
	}
	if rate <= 0 {
		return fmt.Errorf("currency: rate must be positive, got %v", rate)
	}
	return nil
}

// ToMinorFloor / ToMinorCeil are boundary variants of ToMinor: fees and
// PSP settlement amounts often require directional rounding instead of the
// half-away-from-zero default.
func ToMinorFloor(amount float64, code string) int64 {
	return applyRound(amount*Factor(code), math.Floor)
}

// ToMinorCeil rounds a major-unit float up (toward +inf) to ISO minor
// units.
func ToMinorCeil(amount float64, code string) int64 {
	return applyRound(amount*Factor(code), math.Ceil)
}

// ToBaseFloor / ToBaseCeil are the ledger-unit directional variants of
// ToBase.
func ToBaseFloor(amount float64) int64 {
	return applyRound(amount*BaseFactor, math.Floor)
}

// ToBaseCeil rounds a major-unit float up (toward +inf) to ledger units.
func ToBaseCeil(amount float64) int64 {
	return applyRound(amount*BaseFactor, math.Ceil)
}

func applyRound(scaled float64, fn func(float64) float64) int64 {
	return int64(fn(scaled))
}
