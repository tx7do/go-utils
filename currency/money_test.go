package currency

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// 构造与单位回拉
// ============================================================

func TestMoneyConstructors(t *testing.T) {
	m := MoneyFromMinor("USD", 1234)
	assert.Equal(t, int64(123400), m.Base(), "1234 分 = 1234×100 账本单位")
	assert.Equal(t, int64(1234), m.Minor())

	m = MoneyFromBase("USD", 12345678) // 1234.5678
	assert.InDelta(t, 1234.5678, m.Major(), 1e-9)
	assert.Equal(t, int64(123456), m.Minor(), "Minor 向零截断（丢毫厘位）")

	m = MoneyFromMajor("JPY", 120)
	assert.Equal(t, int64(1200000), m.Base(), "1 圆 = 1e4 账本单位")
	assert.Equal(t, int64(120), m.Minor())

	// 币种码归一 + 缺省回退
	assert.Equal(t, "USD", MoneyFromMajor(" usd ", 1).Code())
	assert.Equal(t, "USD", NewMoney("").Code())
}

// ============================================================
// 算术与比较
// ============================================================

func TestMoneyArithmetic(t *testing.T) {
	a := MoneyFromMinor("USD", 1000)
	b := MoneyFromMinor("USD", 250)

	sum, err := a.Add(b)
	require.NoError(t, err)
	assert.Equal(t, int64(1250), sum.Minor())

	diff, err := a.Sub(b)
	require.NoError(t, err)
	assert.Equal(t, int64(750), diff.Minor())

	// 币种不匹配必须拒绝
	_, err = a.Add(MoneyFromMinor("EUR", 1))
	assert.ErrorIs(t, err, ErrCurrencyMismatch)
	_, err = a.Cmp(MoneyFromMinor("EUR", 1))
	assert.ErrorIs(t, err, ErrCurrencyMismatch)

	cmp, err := a.Cmp(b)
	require.NoError(t, err)
	assert.Equal(t, 1, cmp)

	assert.True(t, a.IsPositive())
	assert.True(t, MoneyFromMinor("USD", 0).IsZero())

	neg, err := b.Sub(a)
	require.NoError(t, err)
	assert.True(t, neg.IsNegative())
	abs := neg.Abs()
	assert.Equal(t, int64(750), abs.Minor())
	assert.Equal(t, int64(-750), neg.Minor())
}

func TestMoneyMul(t *testing.T) {
	unit := MoneyFromMinor("USD", 399) // 3.99
	assert.Equal(t, int64(1995), unit.Mul(5).Minor(), "3.99 × 5")
}

// ============================================================
// Allocate / Split（Fowler Money 分配）
// ============================================================

func TestMoneyAllocate(t *testing.T) {
	// 经典用例：100 分拆 3 份等比 → 账本单位 3334/3333/3333（余数进最大
	// 比例档），守恒不变量在账本单位域精确成立
	m := MoneyFromMinor("USD", 100)
	parts, err := m.Split(3)
	require.NoError(t, err)
	got := int64(0)
	for _, p := range parts {
		got += p.Base()
	}
	assert.Equal(t, m.Base(), got, "拆分总和必须守恒（账本单位域）")
	assert.Equal(t, int64(3333), parts[0].Base())
	assert.Equal(t, int64(3334), parts[2].Base(), "累计截断法余数自然落尾位")

	// 按比例 1:2
	parts, err = MoneyFromMinor("USD", 300).Allocate([]int64{1, 2})
	require.NoError(t, err)
	assert.Equal(t, int64(100), parts[0].Minor())
	assert.Equal(t, int64(200), parts[1].Minor())

	// 负数对称分配且守恒
	parts, err = MoneyFromMinor("USD", -100).Split(3)
	require.NoError(t, err)
	got = 0
	for _, p := range parts {
		got += p.Base()
	}
	assert.Equal(t, int64(-10000), got)

	// 账本单位全精度拆分（分下精度不丢）
	parts, err = MoneyFromBase("USD", 10001).Split(2)
	require.NoError(t, err)
	got = 0
	for _, p := range parts {
		got += p.Base()
	}
	assert.Equal(t, int64(10001), got)

	// 非法入参
	_, err = m.Split(0)
	assert.Error(t, err)
	_, err = m.Allocate(nil)
	assert.Error(t, err)
	_, err = m.Allocate([]int64{0, 0})
	assert.Error(t, err)
	_, err = m.Allocate([]int64{1, -1})
	assert.Error(t, err)
}

// ============================================================
// 格式化
// ============================================================

func TestFormatMinor(t *testing.T) {
	assert.Equal(t, "1,234.56", FormatMinor(123456, "USD"))
	assert.Equal(t, "-1,234.56", FormatMinor(-123456, "USD"))
	assert.Equal(t, "0.29", FormatMinor(29, "USD"))
	assert.Equal(t, "120", FormatMinor(120, "JPY"), "0 位币种无小数点")
	assert.Equal(t, "1,234", FormatMinor(1234, "KRW"))
	assert.Equal(t, "1.234", FormatMinor(1234, "KWD"), "3 位小数")
	assert.Equal(t, "12.00", FormatMinor(1200, "EUR"))
}

func TestFormatBase(t *testing.T) {
	// 指数位精度为下限；毫厘位非零则保留
	assert.Equal(t, "1,234.56", FormatBase(12345600, "USD"))
	assert.Equal(t, "1,234.5678", FormatBase(12345678, "USD"))
	assert.Equal(t, "0.001", FormatBase(10, "USD"), "亚分精度保留（base=10 = 0.001）")
	assert.Equal(t, "120", FormatBase(1200000, "JPY"))
	assert.Equal(t, "-0.50", FormatBase(-5000, "USD"))
	assert.Equal(t, "1.2345", FormatBase(12345, "KWD"))
}

func TestFormatSymbolMinor(t *testing.T) {
	assert.Equal(t, "$1,234.56", FormatSymbolMinor(123456, "USD"))
	assert.Equal(t, "¥120", FormatSymbolMinor(120, "JPY"))
	assert.Equal(t, "€12.00", FormatSymbolMinor(1200, "EUR"))
	assert.Equal(t, "XYZ 12.00", FormatSymbolMinor(1200, "XYZ"), "未知币种以码为符号")
}

func TestMoneyString(t *testing.T) {
	assert.Equal(t, "1,234.56 USD", MoneyFromMinor("USD", 123456).String())
	assert.Equal(t, "120 JPY", MoneyFromMinor("JPY", 120).String())
	assert.Equal(t, "1,234.5678 USD", MoneyFromBase("USD", 12345678).String())
}

// ============================================================
// ISO 元数据
// ============================================================

func TestMeta(t *testing.T) {
	info := LookupInfo("USD")
	assert.Equal(t, "840", info.Numeric)
	assert.Equal(t, "$", info.Symbol)

	assert.Equal(t, "¥", Symbol("JPY"))
	assert.Equal(t, "XYZ", Symbol("XYZ"), "未知币种符号回退为码")
	assert.Equal(t, "978", NumericCode("EUR"))
	assert.Empty(t, NumericCode("XYZ"))
	assert.Equal(t, "Japanese Yen", DisplayName("jpy"))
	assert.True(t, IsKnown("USD"))
	assert.False(t, IsKnown("XYZ"))
}

// ============================================================
// 汇率换算
// ============================================================

func TestConvertMinor(t *testing.T) {
	// USD 100 分 × 150 = 150 JPY（大额零头由各币种指数吸收）
	out, err := ConvertMinor(100, "USD", "JPY", 150)
	require.NoError(t, err)
	assert.Equal(t, int64(150), out)

	// rate=1 且同指数 → 恒等
	out, err = ConvertMinor(1234, "USD", "EUR", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1234), out)

	// 指数差吸收：USD → KWD rate 0.30：1.00 USD = 0.30 KWD = 300 fils
	out, err = ConvertMinor(100, "USD", "KWD", 0.3)
	require.NoError(t, err)
	assert.Equal(t, int64(300), out)

	// 非法汇率
	_, err = ConvertMinor(1, "USD", "EUR", 0)
	assert.Error(t, err)
	_, err = ConvertMinor(1, "USD", "EUR", -1.5)
	assert.Error(t, err)
}

func TestConvertBase(t *testing.T) {
	out, err := ConvertBase(10000, "USD", "JPY", 150) // 1.00 USD
	require.NoError(t, err)
	assert.Equal(t, int64(1500000), out, "150 JPY = 150e4 账本单位")
}

// ============================================================
// 舍入策略
// ============================================================

func TestRoundingVariants(t *testing.T) {
	// 默认 half away from zero
	assert.Equal(t, int64(30), ToMinor(0.3, "USD"))
	// floor / ceil
	assert.Equal(t, int64(29), ToMinorFloor(0.2999, "USD"))
	assert.Equal(t, int64(30), ToMinorCeil(0.2999, "USD"))
	assert.Equal(t, int64(-30), ToMinorFloor(-0.2999, "USD"), "floor 向 -inf")
	assert.Equal(t, int64(-29), ToMinorCeil(-0.2999, "USD"), "ceil 向 +inf")
	assert.Equal(t, int64(1), ToBaseFloor(0.00015))
	assert.Equal(t, int64(2), ToBaseCeil(0.00015))
	assert.Equal(t, int64(2), ToBaseCeil(0.00015))
}
