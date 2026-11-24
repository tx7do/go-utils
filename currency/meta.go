package currency

// Currency metadata: ISO 4217 numeric code, display symbol and English
// name for the commonly used currencies. Unknown codes degrade gracefully:
// numeric code empty, symbol falls back to the code itself.

// Info groups the display metadata of a currency.
type Info struct {
	Code    string // ISO 4217 alphabetic code, e.g. "USD"
	Numeric string // ISO 4217 numeric code, e.g. "840" (empty when unknown)
	Symbol  string // display symbol, e.g. "$" (falls back to the code)
	Name    string // English display name (empty when unknown)
}

// metaTable covers the frequently used currencies. Keep it alphabetical by
// code; adding an entry never changes conversion behavior (only display).
var metaTable = map[string]Info{
	"AUD": {Code: "AUD", Numeric: "036", Symbol: "A$", Name: "Australian Dollar"},
	"BHD": {Code: "BHD", Numeric: "048", Symbol: "BHD", Name: "Bahraini Dinar"},
	"BRL": {Code: "BRL", Numeric: "986", Symbol: "R$", Name: "Brazilian Real"},
	"CAD": {Code: "CAD", Numeric: "124", Symbol: "C$", Name: "Canadian Dollar"},
	"CHF": {Code: "CHF", Numeric: "756", Symbol: "CHF", Name: "Swiss Franc"},
	"CLP": {Code: "CLP", Numeric: "152", Symbol: "$", Name: "Chilean Peso"},
	"CNY": {Code: "CNY", Numeric: "156", Symbol: "¥", Name: "Chinese Yuan"},
	"DKK": {Code: "DKK", Numeric: "208", Symbol: "kr", Name: "Danish Krone"},
	"EUR": {Code: "EUR", Numeric: "978", Symbol: "€", Name: "Euro"},
	"GBP": {Code: "GBP", Numeric: "826", Symbol: "£", Name: "British Pound"},
	"HKD": {Code: "HKD", Numeric: "344", Symbol: "HK$", Name: "Hong Kong Dollar"},
	"IDR": {Code: "IDR", Numeric: "360", Symbol: "Rp", Name: "Indonesian Rupiah"},
	"INR": {Code: "INR", Numeric: "356", Symbol: "₹", Name: "Indian Rupee"},
	"ISK": {Code: "ISK", Numeric: "352", Symbol: "kr", Name: "Icelandic Krona"},
	"JPY": {Code: "JPY", Numeric: "392", Symbol: "¥", Name: "Japanese Yen"},
	"KRW": {Code: "KRW", Numeric: "410", Symbol: "₩", Name: "South Korean Won"},
	"KWD": {Code: "KWD", Numeric: "414", Symbol: "KWD", Name: "Kuwaiti Dinar"},
	"MXN": {Code: "MXN", Numeric: "484", Symbol: "$", Name: "Mexican Peso"},
	"MYR": {Code: "MYR", Numeric: "458", Symbol: "RM", Name: "Malaysian Ringgit"},
	"NOK": {Code: "NOK", Numeric: "578", Symbol: "kr", Name: "Norwegian Krone"},
	"NZD": {Code: "NZD", Numeric: "554", Symbol: "NZ$", Name: "New Zealand Dollar"},
	"OMR": {Code: "OMR", Numeric: "512", Symbol: "OMR", Name: "Omani Rial"},
	"PHP": {Code: "PHP", Numeric: "608", Symbol: "₱", Name: "Philippine Peso"},
	"PLN": {Code: "PLN", Numeric: "985", Symbol: "zł", Name: "Polish Zloty"},
	"RUB": {Code: "RUB", Numeric: "643", Symbol: "₽", Name: "Russian Ruble"},
	"SEK": {Code: "SEK", Numeric: "752", Symbol: "kr", Name: "Swedish Krona"},
	"SGD": {Code: "SGD", Numeric: "702", Symbol: "S$", Name: "Singapore Dollar"},
	"THB": {Code: "THB", Numeric: "764", Symbol: "฿", Name: "Thai Baht"},
	"TRY": {Code: "TRY", Numeric: "949", Symbol: "₺", Name: "Turkish Lira"},
	"TWD": {Code: "TWD", Numeric: "901", Symbol: "NT$", Name: "New Taiwan Dollar"},
	"USD": {Code: "USD", Numeric: "840", Symbol: "$", Name: "US Dollar"},
	"VND": {Code: "VND", Numeric: "704", Symbol: "₫", Name: "Vietnamese Dong"},
}

// LookupInfo returns the display metadata for a currency. Unknown codes
// return an entry with an empty numeric/name and the code itself as symbol.
func LookupInfo(code string) Info {
	info, ok := metaTable[Normalize(code)]
	if !ok {
		return Info{Code: Normalize(code), Symbol: Normalize(code)}
	}
	return info
}

// Symbol returns the display symbol of a currency (unknown codes fall back
// to the code itself).
func Symbol(code string) string {
	return LookupInfo(code).Symbol
}

// NumericCode returns the ISO 4217 numeric code (empty when unknown).
func NumericCode(code string) string {
	return LookupInfo(code).Numeric
}

// DisplayName returns the English currency name (empty when unknown).
func DisplayName(code string) string {
	return LookupInfo(code).Name
}

// IsKnown reports whether the code carries metadata in this package. Being
// unknown does NOT make a currency invalid — conversions still work with
// the 2-decimal fallback.
func IsKnown(code string) bool {
	_, ok := metaTable[Normalize(code)]
	return ok
}
