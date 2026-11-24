# currency — 币种金额换算

`currency` 是币种金额换算的统一入口，覆盖三层表示：

1. **主币单位**——面向人的浮点金额（如 `1.23` USD）；
2. **ISO 最小单位**——整数分（ISO 4217 指数表：USD/CNY → 2 位、JPY/KRW → 0 位、KWD/BHD → 3 位）；
3. **账本单位**——统一固定精度整数记账单位（**4 位小数**，1 主币 = `10000` 账本单位），与各币种 ISO 指数解耦。

> **约定**：所有金额换算必须经本包。内联魔法数字换算（`×100` / `÷100`）是静默舍入 bug 与资金差错之源。

[English](README.md) | 简体中文

## 特性

- **ISO 4217 指数表**：0 位与 3 位币种显式收录，其余默认 2 位主档。
- **安全兜底**：未知/小写/空码归一后按 2 位小数兜底——错兜代价最小。
- **half away from zero**：`ToMinor`/`ToBase` 用 `math.Round` 语义，负数（退款/调整）对称不漂移。
- **统一账本单位**：跨币种聚合只乘汇率，无需查各币种档位。
- **边界桥接**：`MinorToBase` / `BaseToMinor` 打通按最小单位结算的渠道与全精度账本。
- **Money 类型**：币种+整数账本单位金额的类型化封装，币种不匹配报错而非静默混算；精确拆分/等分。
- **格式化与元数据**：千分位分组、符号前缀、ISO 4217 数字码/符号/英文名。
- **零依赖**：纯标准库；仅测试使用 `testify`。

## 安装

```bash
go get github.com/tx7do/go-utils/currency
```

```go
import "github.com/tx7do/go-utils/currency"
```

## 快速开始

```go
// ISO 最小单位（展示格式化 / 渠道结算边界）
currency.ToMinor(0.3, "USD")   // 30
currency.ToMinor(120, "JPY")   // 120（0 位币种）

// 账本单位（统一 4 位小数记账）
currency.ToBase(1.2345)        // 12345
currency.FromBase(12345)       // 1.2345

// Money 类型
price := currency.MoneyFromMinor("USD", 399)
total, _ := price.Mul(5)
parts, _ := total.Split(3)     // 精确拆分，总和守恒

// 汇率（调用方显式提供）
out, _ := currency.ConvertMinor(100, "USD", "JPY", 150) // 150

// 格式化与元数据
currency.FormatMinor(123456, "USD")  // "1,234.56"
currency.FormatSymbolMinor(120, "JPY") // "¥120"
currency.Symbol("USD")               // "$"
currency.NumericCode("EUR")          // "978"
```

## API

| 函数 | 说明 |
|---|---|
| `Normalize / NormalizeOr` | 币种码归一（去空白、转大写；空码回退）。 |
| `Exponent / Factor` | ISO 4217 小数位 / 换算系数（未知兜底 2 位 / 100）。 |
| `ToMinor / ToMajor / ToMajorInt` | 主币 ↔ 最小单位（half away from zero / 向零截断）。 |
| `ToBase / FromBase / FromBaseInt` | 主币 ↔ 账本单位（统一 1e4，无币种参数）。 |
| `MinorToBase / BaseToMinor` | 最小单位 ↔ 账本单位桥接（下发向零截断）。 |
| `Money` + 构造/运算 | 类型化金额：`MoneyFromMinor` 等、`Add/Sub/Mul/Cmp/Allocate/Split`（币种校验）。 |
| `Sum / Min / Max`、`LessThan / GreaterThan` | 聚合与比较糖（币种校验）。 |
| `MulRatio(num, den)` / `Percent(pct)` | 精确整数比例乘法与百分比——账本级零浮点路径。 |
| `RoundToMinor()` | 归整亚分账本尘（half away from zero）。 |
| `ParseMoney(s)` / `ParseMinor(s, code)` / `ParseBaseString(s, code)` | 解析金额串（`"1,234.56 USD"`、符号前缀形态）；亚分输入显式拒绝，绝不静默舍入。 |
| `ConvertMinorExact(minor, from, to, num, den)` | `ConvertMinor` 的整数分数汇率版——零浮点。 |
| `CodeFromSymbol(symbol)` | `"$"` → `"USD"`（歧义符号优先 USD，其余字典序）。 |
| `FormatMinor(base, code)` | 展示渲染：千分位 + 指数位小数。 |
| `FormatBase(base, code)` | 账本单位渲染，非零亚分精度保留。 |
| `FormatSymbolMinor(minor, code)` | 带符号前缀渲染（`"$1,234.56"`、`"¥120"`）。 |
| `ConvertMinor / ConvertBase` | 显式汇率跨币种换算（拒绝非正/非有限汇率）。 |
| `ToMinorFloor/Ceil`、`ToBaseFloor/Ceil` | 方向性舍入变体（费率/结算边界）。 |
| `LookupInfo/Symbol/NumericCode/DisplayName/IsKnown` | ISO 4217 元数据：约 30 币种数字码/符号/英文名。 |
| `DefaultCode` | 未指定币种时的参考币种常量（`"USD"`）。 |
| `BaseDigits / BaseFactor` | 账本精度常量：`4` / `10000`。 |

## 精度说明

- **舍入**：向上换算（`ToMinor`、`ToBase`）half away from zero——负数对称；向下换算（`ToMajorInt`、`FromBaseInt`、`BaseToMinor`）向零截断，与整数除法一致。
- **浮点伪差**：`0.29×100` 裸浮点会得 `28.999999...`；`ToMinor` 经 Round 吸收。整数域换算无损往返（见 `TestBaseRoundTrip`）。
- **`BaseToMinor` 截断**：5 账本单位 = 0.5 fils（KWD）→ 截为 0。只在边界下发整数倍最小单位；账本内部永远保留全精度。
- **UTF 无关**：本包不处理字符串编码，只处理数值与显示串。

## 汇率说明

汇率为调用方显式传入的 float64——本包不抓取不缓存。换算结果视为结算估算；账本级换算应在整数域先缩放汇率再应用。

## 扩展

未知码默认 2 位指数。新收录的 ISO 4217 币种加入 `zeroDecimal` / `threeDecimal` 表即可，无需改其他代码。
