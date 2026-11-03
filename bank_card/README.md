# bank_card — Bank Card BIN Lookup

English | [简体中文](README.zh-CN.md)

`bank_card` resolves the issuing bank and card type from a bank card number (BIN) and validates card numbers: it embeds a BIN prefix table for major Chinese banks plus an embedded SQLite BIN database, working out of the box with no external data source.

## Features

- **Built-in table lookup**: `GetNameOfBank` uses the first 6 digits of the card number to look up the built-in `BANKBIN`/`BANKNAME` tables and returns a "Bank·CardName" string;
- **Precise SQLite lookup**: `QueryBankByCardNumber` queries the embedded SQLite database with the first 8/7/6 digits in turn, returning full info including bank, card name, card type and card number length;
- **Card number validation**: `IsValidLuhn` / `IsValidBankCardNo` validate card numbers with the Luhn algorithm;
- **Embedded data**: the BIN database is packed into the binary via `go:embed` — import and go, no external files to deploy.

## Installation

```bash
go get github.com/tx7do/go-utils/bank_card
```

```go
import "github.com/tx7do/go-utils/bank_card"
```

> **Build requirement**: this module depends on [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3), which requires `CGO_ENABLED=1` and a C compiler (e.g. gcc).

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/bank_card"
)

func main() {
	cardNo := "6228480402564890018"

	// 1) Look up the built-in table by the first 6 digits; returns a
	//    "Bank·CardName" string (format like "工商银行·牡丹灵通卡")
	name := bank_card.GetNameOfBank(cardNo)
	fmt.Println("bank name:", name)

	// 2) Query the embedded SQLite by the first 8/7/6 digits for full card info
	bankCard := bank_card.QueryBankByCardNumber(cardNo)
	if bankCard != nil {
		fmt.Println("bank:", bankCard.BankName)                          // bank name
		fmt.Println("card:", bankCard.CardName)                          // card name
		fmt.Println("type:", bankCard.CardType, bankCard.CardTypeName()) // card type code and name
		fmt.Println("length:", bankCard.CardLength)                      // card number length for this BIN
	}

	// 3) Validation: 12-19 digits and a passing Luhn check
	fmt.Println(bank_card.IsValidBankCardNo("6226095711989751")) // true
	fmt.Println(bank_card.IsValidBankCardNo("6226095711989752")) // false (Luhn check fails)
}
```

> The example card numbers come from the module's test cases; `QueryBankByCardNumber` returns `nil` on a miss — check for nil.

## Configuration / Options

`NewDatabase(openFile bool)` supports two data source modes:

| Mode | Argument | Description |
| --- | --- | --- |
| Embedded mode | `openFile=false` (default; what package init uses) | Deserializes `assets.BankCardDatabase` (the `go:embed`-ed `bank_card.db`) into a shared in-memory SQLite database, **read-only** |
| File mode | `openFile=true` | Opens the relative path `assets/bank_card.db` (relative to the process working directory); writable |

`QueryBankByCardNumber` uses the package-level `defaultDatabase` (embedded mode, created at import time); to maintain custom BIN data, open file mode yourself with `NewDatabase(true)` and use the update API.

## API Reference

### Lookup and validation

- `func GetNameOfBank(cardNo string) string` — looks up the first 6 digits in the `BANKBIN` table (the last match wins for duplicate BINs) and returns the corresponding `BANKNAME` entry; returns an empty string for card numbers shorter than 6 digits or on a miss;
- `func QueryBankByCardNumber(cardNo string) *BankCard` — tries the first 8, 7 and 6 digits as the BIN against the embedded SQLite, returning the first full match; returns nil for card numbers shorter than 6 digits or when nothing matches;
- `func IsValidLuhn(cardNo string) bool` — Luhn algorithm check (requires an all-digit string; an empty string returns false);
- `func IsValidBankCardNo(cardNo string) bool` — card number is 12-19 digits and passes the Luhn check.

### Database management

- `func NewDatabase(openFile bool) *Database` — creates a database connection: embedded in-memory mode or file mode;
- `func (d *Database) Close()` — closes the connection;
- `func (d *Database) UpdateBankCardTableCardName(bin uint32, cardName string)` — updates the card name of a given BIN (for fixing/customizing data).

### Data and models

- `var BANKBIN []string` — built-in BIN prefix table (about 1400 entries);
- `var BANKNAME []string` — "Bank·CardName" strings aligned one-to-one with `BANKBIN`;
- `assets.BankCardDatabase []byte` — the embedded SQLite database bytes (about 104 KB);
- `type Bank struct{ Id uint32; BankCode, BankName string }` — bank info;
- `type BankCard struct{ BIN uint32; BankCode, BankName, CardType, CardName string; CardLength uint32 }` — card info;
- Card type constants and `func (b *BankCard) CardTypeName() string`:

| Constant | Value | Meaning | `CardTypeName()` returns |
| --- | --- | --- | --- |
| `CardTypeDC` | `"DC"` | Debit card | `"储蓄卡"` |
| `CardTypeCC` | `"CC"` | Credit card | `"信用卡"` |
| `CardTypeSCC` | `"SCC"` | Semi credit card | `"准贷记卡"` |
| `CardTypePC` | `"PC"` | Prepaid card | `"预付费卡"` |

Unmatched card types return an empty string.

## Notes

- **CGO is required**: the dependency on `mattn/go-sqlite3` means cross-compilation or CGO-disabled builds (e.g. `CGO_ENABLED=0` artifacts) cannot use `QueryBankByCardNumber`; for pure validation without CGO you can use only `GetNameOfBank` / `IsValidBankCardNo` (they do not touch SQLite) — note that the package-level `defaultDatabase` is still initialized at import time;
- **Data freshness**: the built-in BIN table and SQLite data are a static snapshot; newly issued BINs / card types may be missing — maintain them yourself with `NewDatabase(true)` + `UpdateBankCardTableCardName`;
- **File mode path**: `NewDatabase(true)` opens `assets/bank_card.db` relative to the process working directory; make sure that path resolves when deployed;
- **Validation limits**: passing Luhn does not mean the card number actually exists — it only rules out entry mistakes;
- **Standalone Go module** (`replace github.com/tx7do/go-utils => ../` in its `go.mod`); install via `go get github.com/tx7do/go-utils/bank_card`.

## References / Acknowledgements

- [China UnionPay Bank Card BIN Checker](https://github.com/hexindai/bcbc)
- [CommonUtilLibrary](https://github.com/AbrahamCaiJin/CommonUtilLibrary/blob/master/CommonUtil/src/main/java/com/jingewenku/abrahamcaijin/commonutil/BankCheck.java)
- [Luhn algorithm - wikipedia](https://en.wikipedia.org/wiki/Luhn_algorithm)
- [Luhn algorithm - geeksforgeeks](https://www.geeksforgeeks.org/luhn-algorithm/)
- [Luhn's algorithm to validate credit / debit card Numbers](https://medium.com/@akshaymohite/luhns-algorithm-to-validate-credit-debit-card-numbers-1952e6c7a9d0)
- [干货丨银行卡号编码规则及其应用](https://www.woshipm.com/pd/371041.html)
- [bankInfo](https://github.com/giraffe-lib/bankInfo/blob/main/src/map.js)
- [BankCards](https://github.com/geekgao/BankCards/blob/master/bankcode.py)
- [bcbc](https://github.com/hexindai/bcbc)
- [banks-db](https://github.com/ramoona/banks-db)
- [SwiftCodes](https://github.com/PeterNotenboom/SwiftCodes)
- [BanksDataWorldWide](https://github.com/abdalrhman-alajlouni/BanksDataWorldWide)
- [bankcard](https://github.com/caijf/bankcard)
