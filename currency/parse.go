package currency

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// symbolToCode maps display symbols back to currency codes. Symbols shared
// by several currencies (e.g. "¥" for CNY/JPY, "$" for USD/MXN/CLP)
// prefer USD when covered, otherwise resolve to the alphabetically first
// code; callers who need a specific currency should pass the explicit code.
var symbolToCode = func() map[string]string {
	candidates := make(map[string][]string)
	for code, info := range metaTable {
		candidates[info.Symbol] = append(candidates[info.Symbol], code)
	}
	m := make(map[string]string)
	for sym, codes := range candidates {
		sort.Strings(codes)
		chosen := codes[0]
		for _, c := range codes { // 歧义符号优先 USD（全球默认习惯）
			if c == "USD" {
				chosen = c
				break
			}
		}
		m[sym] = chosen
	}
	return m
}()

// splitAmountCode splits a money string into the amount token and the
// currency token. Accepted shapes:
//
//	"1234.56 USD" / "1234.56USD"  — explicit code
//	"$1,234.56 USD"               — symbol attached to the amount
//	"$1,234.56"                   — symbol as the currency carrier
//	"¥120"                        — symbol resolves ambiguous codes
//	  (alphabetically first), so prefer explicit codes for shared symbols
func splitAmountCode(s string) (amount, code string, err error) {
	fields := strings.Fields(s)
	switch len(fields) {
	case 1:
		// 可能是 "$1,234.56" 或裸数字（无币种信息，回退 DefaultCode）
		tok := fields[0]
		if sym, ok := stripLeadingSymbol(tok); ok {
			code := CodeFromSymbol(sym)
			if code == "" {
				return "", "", fmt.Errorf("currency: unknown symbol %q", sym)
			}
			return tok[len(sym):], code, nil
		}
		return tok, DefaultCode, nil
	case 2:
		code := Normalize(fields[1])
		if _, known := metaTable[code]; !known {
			// 第二段可能是符号（"$USD" 场景已拆开；这里处理 "12.00 $"）
			if resolved := CodeFromSymbol(fields[1]); resolved != "" {
				return fields[0], resolved, nil
			}
		}
		amount := fields[0]
		if sym, ok := stripLeadingSymbol(amount); ok && Symbol(code) == sym {
			amount = amount[len(sym):]
		}
		return amount, code, nil
	default:
		return "", "", fmt.Errorf("currency: cannot parse money string %q", s)
	}
}

// stripLeadingSymbol reports the longest attached display symbol of an
// amount token, if any ("$1,234" → "$"; "S$12" → "S$"; longest match wins
// so "HK$5" is not misread as "H").
func stripLeadingSymbol(tok string) (string, bool) {
	best := ""
	for sym := range symbolToCode {
		if len(sym) > len(best) && strings.HasPrefix(tok, sym) {
			best = sym
		}
	}
	return best, best != ""
}

// ParseMinor parses a display decimal string into ISO minor units.
// Thousands separators (",") are ignored; the fractional part must not
// exceed the currency's exponent — sub-minor digits are rejected rather
// than silently rounded ("0.001" USD is an error; use ParseBase).
func ParseMinor(s, code string) (int64, error) {
	base, err := ParseBaseString(s, code)
	if err != nil {
		return 0, err
	}
	// 亚分位拒绝：账本单位里残留 ISO 指数以下精度 → 显式报错而非静默截断
	if rem := base % int64(math.Pow10(BaseDigits-Exponent(code))); rem != 0 {
		return 0, fmt.Errorf("currency: %q carries sub-minor digits beyond %s precision", s, code)
	}
	return BaseToMinor(base, code), nil
}

// ParseBaseString parses a display decimal string into ledger units.
// Accepts an optional sign, thousands separators, and up to BaseDigits
// fractional digits ("1234.5678"). Digits beyond ledger precision are
// rejected rather than silently rounded.
func ParseBaseString(s, code string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("currency: empty amount")
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimPrefix(s, "+")
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, fmt.Errorf("currency: no digits in %q", s)
	}

	intPart, fracPart, hasFrac := strings.Cut(s, ".")
	if !allDigits(intPart) || intPart == "" {
		return 0, fmt.Errorf("currency: invalid integer part in %q", s)
	}
	if hasFrac {
		if !allDigits(fracPart) {
			return 0, fmt.Errorf("currency: invalid fractional part in %q", s)
		}
		if len(fracPart) > BaseDigits {
			return 0, fmt.Errorf("currency: %d fractional digits exceed ledger precision (%d)", len(fracPart), BaseDigits)
		}
	}

	var b strings.Builder
	b.WriteString(intPart)
	// 分数右补零到账本精度：0.56 → 5600
	b.WriteString(fracPart + strings.Repeat("0", BaseDigits-len(fracPart)))

	var base int64
	for _, ch := range []byte(b.String()) {
		base = base*10 + int64(ch-'0')
	}
	if base < 0 {
		return 0, fmt.Errorf("currency: amount overflows int64")
	}
	if neg {
		base = -base
	}
	return base, nil
}

// ParseMoney parses a money string — the exact format `Money.String`
// emits, plus symbol-prefixed variants (see splitAmountCode) — into a
// Money value. Currency mismatch is impossible by construction: the code
// travels with the string.
//
//	Money.ParseMoney(MoneyFromMinor("USD", 123456).String()) // roundtrip
func ParseMoney(s string) (Money, error) {
	amountTok, code, err := splitAmountCode(s)
	if err != nil {
		return Money{}, err
	}
	base, err := ParseBaseString(amountTok, code)
	if err != nil {
		return Money{}, err
	}
	return MoneyFromBase(code, base), nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range []byte(s) {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
