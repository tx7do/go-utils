"""One-off generator: rewrites meta.go with the full active ISO 4217 list."""
import io
import re

DATA = open("iso4217_active.txt", encoding="utf-8").read()

def clean(name):
    return re.sub(r"\s*\([^)]*\)", "", name).strip()

rows = []
for line in DATA.strip().splitlines():
    code, numeric, name, _minor = line.split("|")
    rows.append((code, numeric, clean(name)))

SYMBOLS = {
    "AUD": "A$", "BRL": "R$", "CAD": "C$", "CHF": "CHF", "CLP": "$",
    "CNY": "¥", "DKK": "kr", "EUR": "€", "GBP": "£", "HKD": "HK$",
    "IDR": "Rp", "INR": "₹", "ISK": "kr", "JPY": "¥", "KRW": "₩",
    "KWD": "KWD", "MXN": "$", "MYR": "RM", "NOK": "kr", "NZD": "NZ$",
    "OMR": "OMR", "PHP": "₱", "PLN": "zł", "RUB": "₽", "SEK": "kr",
    "SGD": "S$", "THB": "฿", "TRY": "₺", "TWD": "NT$", "USD": "$",
    "VND": "₫",
}

numeric_lines = []
name_lines = []
symbol_lines = []
for code, numeric, name in rows:
    numeric_lines.append(f'\t\t"{code}": "{numeric}",')
    name_lines.append(f'\t\t"{code}": "{name}",')
    if code in SYMBOLS:
        symbol_lines.append(f'\t\t"{code}": "{SYMBOLS[code]}",')

out = f'''package currency

// ISO 4217 currency metadata tables (full active-code list from the
// official standard: numeric code + English name for every active entry,
// display symbols for the commonly used subset). Unknown codes degrade
// gracefully: numeric/name empty, symbol falls back to the code itself.

// Info groups the display metadata of a currency.
type Info struct {{
\tCode    string // ISO 4217 alphabetic code, e.g. "USD"
\tNumeric string // ISO 4217 numeric code, e.g. "840" (empty when unknown)
\tSymbol  string // display symbol, e.g. "$" (falls back to the code)
\tName    string // English display name (empty when unknown)
}}

// numericTable covers every active ISO 4217 code (funds codes included).
var numericTable = map[string]string{{
{chr(10).join(numeric_lines)}
}}

// nameTable covers every active ISO 4217 code.
var nameTable = map[string]string{{
{chr(10).join(name_lines)}
}}

// symbolTable carries display symbols for the commonly used subset.
var symbolTable = map[string]string{{
{chr(10).join(symbol_lines)}
}}

// LookupInfo returns the display metadata for a currency. Unknown codes
// return an entry with an empty numeric/name and the code itself as symbol.
func LookupInfo(code string) Info {{
\tcode = Normalize(code)
\treturn Info{{
\t\tCode:    code,
\t\tNumeric: numericTable[code],
\t\tSymbol:  Symbol(code),
\t\tName:    nameTable[code],
\t}}
}}

// Symbol returns the display symbol of a currency (unknown codes fall back
// to the code itself).
func Symbol(code string) string {{
\tcode = Normalize(code)
\tif sym, ok := symbolTable[code]; ok {{
\t\treturn sym
\t}}
\treturn code
}}

// NumericCode returns the ISO 4217 numeric code (empty when unknown).
func NumericCode(code string) string {{
\treturn numericTable[Normalize(code)]
}}

// CodeFromNumeric resolves an ISO 4217 numeric code to its alphabetic
// code ("840" → "USD"). Unknown numerics return "".
func CodeFromNumeric(numeric string) string {{
\tfor code, n := range numericTable {{
\t\tif n == numeric {{
\t\t\treturn code
\t\t}}
\t}}
\treturn ""
}}

// DisplayName returns the English currency name (empty when unknown).
func DisplayName(code string) string {{
\treturn nameTable[Normalize(code)]
}}

// IsKnown reports whether the code carries metadata in this package. Being
// unknown does NOT make a currency invalid — conversions still work with
// the 2-decimal fallback.
func IsKnown(code string) bool {{
\t_, ok := numericTable[Normalize(code)]
\treturn ok
}}
'''
io.open("meta.go", "w", encoding="utf-8", newline="").write(out)
print("meta.go generated:", len(rows), "entries")
