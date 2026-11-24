package currency

import "sort"

// ISO 4217 currency metadata tables (full active-code list from the
// official standard: numeric code + English name for every active entry,
// display symbols for the commonly used subset). Unknown codes degrade
// gracefully: numeric/name empty, symbol falls back to the code itself.

// Info groups the display metadata of a currency.
type Info struct {
	Code    string // ISO 4217 alphabetic code, e.g. "USD"
	Numeric string // ISO 4217 numeric code, e.g. "840" (empty when unknown)
	Symbol  string // display symbol, e.g. "$" (falls back to the code)
	Name    string // English display name (empty when unknown)
}

// numericTable covers every active ISO 4217 code (funds codes included).
var numericTable = map[string]string{
	"AED": "784",
	"AFN": "971",
	"ALL": "008",
	"AMD": "051",
	"AOA": "973",
	"ARS": "032",
	"AUD": "036",
	"AWG": "533",
	"AZN": "944",
	"BAM": "977",
	"BBD": "052",
	"BDT": "050",
	"BHD": "048",
	"BIF": "108",
	"BMD": "060",
	"BND": "096",
	"BOB": "068",
	"BRL": "986",
	"BSD": "044",
	"BTN": "064",
	"BWP": "072",
	"BYN": "933",
	"BZD": "084",
	"CAD": "124",
	"CDF": "976",
	"CHE": "947",
	"CHF": "756",
	"CHW": "948",
	"CLF": "990",
	"CLP": "152",
	"CNY": "156",
	"COP": "170",
	"COU": "970",
	"CRC": "188",
	"CUP": "192",
	"CVE": "132",
	"CZK": "203",
	"DJF": "262",
	"DKK": "208",
	"DOP": "214",
	"DZD": "012",
	"EGP": "818",
	"ERN": "232",
	"ETB": "230",
	"EUR": "978",
	"FJD": "242",
	"FKP": "238",
	"GBP": "826",
	"GEL": "981",
	"GHS": "936",
	"GIP": "292",
	"GMD": "270",
	"GNF": "324",
	"GTQ": "320",
	"GYD": "328",
	"HKD": "344",
	"HNL": "340",
	"HTG": "332",
	"HUF": "348",
	"IDR": "360",
	"ILS": "376",
	"INR": "356",
	"IQD": "368",
	"IRR": "364",
	"ISK": "352",
	"JMD": "388",
	"JOD": "400",
	"JPY": "392",
	"KES": "404",
	"KGS": "417",
	"KHR": "116",
	"KMF": "174",
	"KPW": "408",
	"KRW": "410",
	"KWD": "414",
	"KYD": "136",
	"KZT": "398",
	"LAK": "418",
	"LBP": "422",
	"LKR": "144",
	"LRD": "430",
	"LSL": "426",
	"LYD": "434",
	"MAD": "504",
	"MDL": "498",
	"MGA": "969",
	"MKD": "807",
	"MMK": "104",
	"MNT": "496",
	"MOP": "446",
	"MRU": "929",
	"MUR": "480",
	"MVR": "462",
	"MWK": "454",
	"MXN": "484",
	"MXV": "979",
	"MYR": "458",
	"MZN": "943",
	"NAD": "516",
	"NGN": "566",
	"NIO": "558",
	"NOK": "578",
	"NPR": "524",
	"NZD": "554",
	"OMR": "512",
	"PAB": "590",
	"PEN": "604",
	"PGK": "598",
	"PHP": "608",
	"PKR": "586",
	"PLN": "985",
	"PYG": "600",
	"QAR": "634",
	"RON": "946",
	"RSD": "941",
	"RUB": "643",
	"RWF": "646",
	"SAR": "682",
	"SBD": "090",
	"SCR": "690",
	"SDG": "938",
	"SEK": "752",
	"SGD": "702",
	"SHP": "654",
	"SLE": "925",
	"SOS": "706",
	"SRD": "968",
	"SSP": "728",
	"STN": "930",
	"SVC": "222",
	"SYP": "760",
	"SZL": "748",
	"THB": "764",
	"TJS": "972",
	"TMT": "934",
	"TND": "788",
	"TOP": "776",
	"TRY": "949",
	"TTD": "780",
	"TWD": "901",
	"TZS": "834",
	"UAH": "980",
	"UGX": "800",
	"USD": "840",
	"UYU": "858",
	"UZS": "860",
	"VED": "926",
	"VES": "928",
	"VND": "704",
	"VUV": "548",
	"WST": "882",
	"XAF": "950",
	"XCD": "951",
	"XCG": "532",
	"XDR": "960",
	"XOF": "952",
	"XPF": "953",
	"YER": "886",
	"ZAR": "710",
	"ZMW": "967",
	"ZWG": "924",
}

// nameTable covers every active ISO 4217 code.
var nameTable = map[string]string{
	"AED": "United Arab Emirates dirham",
	"AFN": "Afghan afghani",
	"ALL": "Albanian lek",
	"AMD": "Armenian dram",
	"AOA": "Angolan kwanza",
	"ARS": "Argentine peso",
	"AUD": "Australian dollar",
	"AWG": "Aruban florin",
	"AZN": "Azerbaijani manat",
	"BAM": "Bosnia and Herzegovina convertible mark",
	"BBD": "Barbados dollar",
	"BDT": "Bangladeshi taka",
	"BHD": "Bahraini dinar",
	"BIF": "Burundian franc",
	"BMD": "Bermudian dollar",
	"BND": "Brunei dollar",
	"BOB": "Boliviano",
	"BRL": "Brazilian real",
	"BSD": "Bahamian dollar",
	"BTN": "Bhutanese ngultrum",
	"BWP": "Botswana pula",
	"BYN": "Belarusian ruble",
	"BZD": "Belize dollar",
	"CAD": "Canadian dollar",
	"CDF": "Congolese franc",
	"CHE": "WIR euro",
	"CHF": "Swiss franc",
	"CHW": "WIR franc",
	"CLF": "Unidad de Fomento",
	"CLP": "Chilean peso",
	"CNY": "Renminbi",
	"COP": "Colombian peso",
	"COU": "Unidad de Valor Real",
	"CRC": "Costa Rican colon",
	"CUP": "Cuban peso",
	"CVE": "Cape Verdean escudo",
	"CZK": "Czech koruna",
	"DJF": "Djiboutian franc",
	"DKK": "Danish krone",
	"DOP": "Dominican peso",
	"DZD": "Algerian dinar",
	"EGP": "Egyptian pound",
	"ERN": "Eritrean nakfa",
	"ETB": "Ethiopian birr",
	"EUR": "Euro",
	"FJD": "Fiji dollar",
	"FKP": "Falkland Islands pound",
	"GBP": "Pound sterling",
	"GEL": "Georgian lari",
	"GHS": "Ghanaian cedi",
	"GIP": "Gibraltar pound",
	"GMD": "Gambian dalasi",
	"GNF": "Guinean franc",
	"GTQ": "Guatemalan quetzal",
	"GYD": "Guyanese dollar",
	"HKD": "Hong Kong dollar",
	"HNL": "Honduran lempira",
	"HTG": "Haitian gourde",
	"HUF": "Hungarian forint",
	"IDR": "Indonesian rupiah",
	"ILS": "Israeli new shekel",
	"INR": "Indian rupee",
	"IQD": "Iraqi dinar",
	"IRR": "Iranian rial",
	"ISK": "Icelandic krona",
	"JMD": "Jamaican dollar",
	"JOD": "Jordanian dinar",
	"JPY": "Japanese yen",
	"KES": "Kenyan shilling",
	"KGS": "Kyrgyzstani som",
	"KHR": "Cambodian riel",
	"KMF": "Comoro franc",
	"KPW": "North Korean won",
	"KRW": "South Korean won",
	"KWD": "Kuwaiti dinar",
	"KYD": "Cayman Islands dollar",
	"KZT": "Kazakhstani tenge",
	"LAK": "Lao kip",
	"LBP": "Lebanese pound",
	"LKR": "Sri Lankan rupee",
	"LRD": "Liberian dollar",
	"LSL": "Lesotho loti",
	"LYD": "Libyan dinar",
	"MAD": "Moroccan dirham",
	"MDL": "Moldovan leu",
	"MGA": "Malagasy ariary",
	"MKD": "Macedonian denar",
	"MMK": "Myanmar kyat",
	"MNT": "Mongolian togrog",
	"MOP": "Macanese pataca",
	"MRU": "Mauritanian ouguiya",
	"MUR": "Mauritian rupee",
	"MVR": "Maldivian rufiyaa",
	"MWK": "Malawian kwacha",
	"MXN": "Mexican peso",
	"MXV": "Mexican Unidad de Inversion",
	"MYR": "Malaysian ringgit",
	"MZN": "Mozambican metical",
	"NAD": "Namibian dollar",
	"NGN": "Nigerian naira",
	"NIO": "Nicaraguan cordoba",
	"NOK": "Norwegian krone",
	"NPR": "Nepalese rupee",
	"NZD": "New Zealand dollar",
	"OMR": "Omani rial",
	"PAB": "Panamanian balboa",
	"PEN": "Peruvian sol",
	"PGK": "Papua New Guinean kina",
	"PHP": "Philippine peso",
	"PKR": "Pakistani rupee",
	"PLN": "Polish zloty",
	"PYG": "Paraguayan guarani",
	"QAR": "Qatari riyal",
	"RON": "Romanian leu",
	"RSD": "Serbian dinar",
	"RUB": "Russian ruble",
	"RWF": "Rwandan franc",
	"SAR": "Saudi riyal",
	"SBD": "Solomon Islands dollar",
	"SCR": "Seychelles rupee",
	"SDG": "Sudanese pound",
	"SEK": "Swedish krona",
	"SGD": "Singapore dollar",
	"SHP": "Saint Helena pound",
	"SLE": "Sierra Leonean leone",
	"SOS": "Somali shilling",
	"SRD": "Surinamese dollar",
	"SSP": "South Sudanese pound",
	"STN": "Sao Tome and Principe dobra",
	"SVC": "Salvadoran colon",
	"SYP": "Syrian pound",
	"SZL": "Swazi lilangeni",
	"THB": "Thai baht",
	"TJS": "Tajikistani somoni",
	"TMT": "Turkmenistani manat",
	"TND": "Tunisian dinar",
	"TOP": "Tongan paanga",
	"TRY": "Turkish lira",
	"TTD": "Trinidad and Tobago dollar",
	"TWD": "New Taiwan dollar",
	"TZS": "Tanzanian shilling",
	"UAH": "Ukrainian hryvnia",
	"UGX": "Ugandan shilling",
	"USD": "United States dollar",
	"UYU": "Uruguayan peso",
	"UZS": "Uzbekistani sum",
	"VED": "Venezuelan digital bolivar",
	"VES": "Venezuelan sovereign bolivar",
	"VND": "Vietnamese dong",
	"VUV": "Vanuatu vatu",
	"WST": "Samoan tala",
	"XAF": "CFA franc BEAC",
	"XCD": "East Caribbean dollar",
	"XCG": "Caribbean guilder",
	"XDR": "Special drawing rights",
	"XOF": "CFA franc BCEAO",
	"XPF": "CFP franc",
	"YER": "Yemeni rial",
	"ZAR": "South African rand",
	"ZMW": "Zambian kwacha",
	"ZWG": "Zimbabwe Gold",
}

// symbolTable carries display symbols for the commonly used subset.
var symbolTable = map[string]string{
	"AUD": "A$",
	"BRL": "R$",
	"CAD": "C$",
	"CHF": "CHF",
	"CLP": "$",
	"CNY": "¥",
	"DKK": "kr",
	"EUR": "€",
	"GBP": "£",
	"HKD": "HK$",
	"IDR": "Rp",
	"INR": "₹",
	"ISK": "kr",
	"JPY": "¥",
	"KRW": "₩",
	"KWD": "KWD",
	"MXN": "$",
	"MYR": "RM",
	"NOK": "kr",
	"NZD": "NZ$",
	"OMR": "OMR",
	"PHP": "₱",
	"PLN": "zł",
	"RUB": "₽",
	"SEK": "kr",
	"SGD": "S$",
	"THB": "฿",
	"TRY": "₺",
	"TWD": "NT$",
	"USD": "$",
	"VND": "₫",
}

// LookupInfo returns the display metadata for a currency. Unknown codes
// return an entry with an empty numeric/name and the code itself as symbol.
func LookupInfo(code string) Info {
	code = Normalize(code)
	return Info{
		Code:    code,
		Numeric: numericTable[code],
		Symbol:  Symbol(code),
		Name:    nameTable[code],
	}
}

// Symbol returns the display symbol of a currency (unknown codes fall back
// to the code itself).
func Symbol(code string) string {
	code = Normalize(code)
	if sym, ok := symbolTable[code]; ok {
		return sym
	}
	return code
}

// NumericCode returns the ISO 4217 numeric code (empty when unknown).
func NumericCode(code string) string {
	return numericTable[Normalize(code)]
}

// CodeFromSymbol resolves a display symbol to a currency code. Ambiguous
// symbols (shared by several currencies, e.g. "¥" or "$") prefer USD when
// covered, otherwise the alphabetically first code. Unknown symbols
// return "".
func CodeFromSymbol(symbol string) string {
	sym := Normalize(symbol)
	var candidates []string
	for code, s := range symbolTable {
		if s == sym {
			candidates = append(candidates, code)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.Strings(candidates)
	for _, c := range candidates {
		if c == "USD" {
			return c
		}
	}
	return candidates[0]
}

// CodeFromNumeric resolves an ISO 4217 numeric code to its alphabetic
// code ("840" → "USD"). Unknown numerics return "".
func CodeFromNumeric(numeric string) string {
	for code, n := range numericTable {
		if n == numeric {
			return code
		}
	}
	return ""
}

// DisplayName returns the English currency name (empty when unknown).
func DisplayName(code string) string {
	return nameTable[Normalize(code)]
}

// IsKnown reports whether the code carries metadata in this package. Being
// unknown does NOT make a currency invalid — conversions still work with
// the 2-decimal fallback.
func IsKnown(code string) bool {
	_, ok := numericTable[Normalize(code)]
	return ok
}
