# bank_card —— 银行卡 BIN 查询

[English](README.md) | 简体中文

`bank_card` 提供银行卡号（BIN）的银行归属与卡种查询、以及卡号合法性校验：内置中国主要银行的 BIN 前缀表与一份内嵌 SQLite BIN 数据库，开箱即用、无需外部数据源。

## 特性

- **内置表快速查询**：`GetNameOfBank` 用卡号前 6 位查内置 `BANKBIN`/`BANKNAME` 表，返回"银行·卡种"名称；
- **SQLite 精确查询**：`QueryBankByCardNumber` 依次用卡号前 8/7/6 位查内嵌 SQLite 数据库，返回含银行、卡名、卡种、卡号长度的完整信息；
- **卡号校验**：`IsValidLuhn` / `IsValidBankCardNo` 基于 Luhn 算法校验卡号；
- **数据内嵌**：BIN 数据库通过 `go:embed` 打包进二进制，import 即可用，无需部署外部文件。

## 安装

```bash
go get github.com/tx7do/go-utils/bank_card
```

```go
import "github.com/tx7do/go-utils/bank_card"
```

> **编译要求**：本模块依赖 [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3)，需要 `CGO_ENABLED=1` 与 C 编译器（如 gcc）。

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/bank_card"
)

func main() {
	cardNo := "6228480402564890018"

	// 1) 按前 6 位查内置表，返回"银行·卡种"名称（如"工商银行·牡丹灵通卡"格式）
	name := bank_card.GetNameOfBank(cardNo)
	fmt.Println("bank name:", name)

	// 2) 按前 8/7/6 位查内嵌 SQLite，返回完整卡信息
	bankCard := bank_card.QueryBankByCardNumber(cardNo)
	if bankCard != nil {
		fmt.Println("bank:", bankCard.BankName)                 // 银行名称
		fmt.Println("card:", bankCard.CardName)                  // 银行卡名称
		fmt.Println("type:", bankCard.CardType, bankCard.CardTypeName()) // 卡种代码与名称
		fmt.Println("length:", bankCard.CardLength)              // 该 BIN 对应卡号长度
	}

	// 3) 卡号校验：长度 12-19 位且通过 Luhn 算法
	fmt.Println(bank_card.IsValidBankCardNo("6226095711989751")) // true
	fmt.Println(bank_card.IsValidBankCardNo("6226095711989752")) // false（Luhn 校验失败）
}
```

> 示例卡号取自模块测试用例；`QueryBankByCardNumber` 未命中时返回 `nil`，请判空。

## 配置 / 选项

`NewDatabase(openFile bool)` 支持两种数据源模式：

| 模式 | 参数 | 说明 |
| --- | --- | --- |
| 内嵌模式 | `openFile=false`（默认，包初始化即采用） | 把 `assets.BankCardDatabase`（`go:embed` 的 `bank_card.db`）反序列化进共享内存 SQLite，**只读** |
| 文件模式 | `openFile=true` | 打开相对路径 `assets/bank_card.db`（相对进程工作目录），可写 |

`QueryBankByCardNumber` 使用包级的 `defaultDatabase`（内嵌模式，import 时自动创建）；要维护自定义 BIN 数据时，自行 `NewDatabase(true)` 打开文件模式并调用更新接口。

## API 参考

### 查询与校验

- `func GetNameOfBank(cardNo string) string` —— 取卡号前 6 位在 `BANKBIN` 表中查找（重复 BIN 取最后一次命中），返回 `BANKNAME` 对应项；卡号不足 6 位或未命中返回空串；
- `func QueryBankByCardNumber(cardNo string) *BankCard` —— 依次尝试卡号前 8、7、6 位作为 BIN 查内嵌 SQLite，返回首个命中的完整卡信息；卡号不足 6 位或全部未命中返回 nil；
- `func IsValidLuhn(cardNo string) bool` —— Luhn 算法校验（要求纯数字字符串，空串返回 false）；
- `func IsValidBankCardNo(cardNo string) bool` —— 卡号长度 12-19 位且通过 Luhn 算法。

### 数据库管理

- `func NewDatabase(openFile bool) *Database` —— 创建数据库连接：内嵌内存模式或文件模式；
- `func (d *Database) Close()` —— 关闭连接；
- `func (d *Database) UpdateBankCardTableCardName(bin uint32, cardName string)` —— 更新指定 BIN 的卡名（可用于修正/自定义数据）。

### 数据与模型

- `var BANKBIN []string` —— 内置 BIN 前缀表（约 1400 项）；
- `var BANKNAME []string` —— 与 `BANKBIN` 一一对应的"银行·卡种"名称表；
- `assets.BankCardDatabase []byte` —— `go:embed` 内嵌的 SQLite 数据库字节（约 104 KB）；
- `type Bank struct{ Id uint32; BankCode, BankName string }` —— 银行信息；
- `type BankCard struct{ BIN uint32; BankCode, BankName, CardType, CardName string; CardLength uint32 }` —— 银行卡信息；
- 卡种常量与 `func (b *BankCard) CardTypeName() string`：

| 常量 | 值 | 含义 | `CardTypeName()` 返回 |
| --- | --- | --- | --- |
| `CardTypeDC` | `"DC"` | 储蓄卡 | `"储蓄卡"` |
| `CardTypeCC` | `"CC"` | 信用卡 | `"信用卡"` |
| `CardTypeSCC` | `"SCC"` | 准贷记卡 | `"准贷记卡"` |
| `CardTypePC` | `"PC"` | 预付费卡 | `"预付费卡"` |

未匹配的卡种返回空串。

## 注意事项

- **需要 CGO**：依赖 `mattn/go-sqlite3`，交叉编译或禁用 CGO 的环境（如默认的 `CGO_ENABLED=0` 构建产物）无法使用 `QueryBankByCardNumber`；纯查询场景若不需要 CGO，可仅使用 `GetNameOfBank` / `IsValidBankCardNo`（不触发 SQLite）——注意包级 `defaultDatabase` 在 import 时仍会初始化；
- **数据时效**：内置 BIN 表与 SQLite 数据为静态快照，新发 BIN / 新卡种可能缺失，可用 `NewDatabase(true)` + `UpdateBankCardTableCardName` 自行维护；
- **文件模式路径**：`NewDatabase(true)` 打开的是相对进程工作目录的 `assets/bank_card.db`，部署时需保证该相对路径可达；
- **校验局限**：Luhn 通过不代表卡号真实存在，仅排除录入错误；
- **本模块为独立 Go 模块**（`go.mod` 中 `replace github.com/tx7do/go-utils => ../`），通过 `go get github.com/tx7do/go-utils/bank_card` 引入。

## 参考资料 / 致谢

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
