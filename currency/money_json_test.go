package currency

import (
	"encoding/json"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// 溢出检查
// ============================================================

func TestMoneyOverflowDetection(t *testing.T) {
	huge := MoneyFromBase("USD", math.MaxInt64/2+1)

	_, err := huge.Add(huge)
	assert.ErrorIs(t, err, ErrOverflow, "正溢出必须报错而非回绕")

	minV := MoneyFromBase("USD", math.MinInt64)
	_, err = minV.Sub(MoneyFromBase("USD", 1))
	assert.ErrorIs(t, err, ErrOverflow, "负向下溢必须报错")

	_, err = MoneyFromBase("USD", math.MaxInt64/2).Mul(4)
	assert.ErrorIs(t, err, ErrOverflow)

	// 正常运算不受影响
	ok, err := MoneyFromBase("USD", 100).Add(MoneyFromBase("USD", 200))
	require.NoError(t, err)
	assert.Equal(t, int64(300), ok.Base())
}

// ============================================================
// JSON 序列化
// ============================================================

func TestMoneyJSON(t *testing.T) {
	m := MoneyFromMinor("USD", 123456)
	raw, err := m.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"1,234.56 USD"`, string(raw))

	var back Money
	require.NoError(t, back.UnmarshalJSON(raw))
	assert.Equal(t, m.Base(), back.Base())
	assert.Equal(t, m.Code(), back.Code())

	// 结构体内嵌：标准 encoding/json 直接工作
	type doc struct {
		Amount Money  `json:"amount"`
		Bonus  *Money `json:"bonus,omitempty"`
		Zero   Money  `json:"zero"`
	}
	rawDoc, err := json.Marshal(doc{Amount: m, Zero: MoneyFromMinor("JPY", 120)})
	require.NoError(t, err)
	assert.JSONEq(t, `{"amount":"1,234.56 USD","zero":"120 JPY"}`, string(rawDoc))

	var round doc
	require.NoError(t, json.Unmarshal(rawDoc, &round))
	assert.Equal(t, m.Base(), round.Amount.Base())
	assert.Equal(t, int64(1200000), round.Zero.Base())

	// null → 零值
	var nullBack Money
	require.NoError(t, nullBack.UnmarshalJSON([]byte("null")))
	assert.True(t, nullBack.IsZero())

	// 非字符串类型拒绝
	require.Error(t, (new(Money)).UnmarshalJSON([]byte("123")))
}

// ============================================================
// fuzz 种子形态的属性检查（确定性快速版）
// ============================================================

func TestMoneyFormatParseProperty(t *testing.T) {
	// 随机 base × 随机已知/未知币种：Format → Parse 必须还原同一 base
	codes := []string{"USD", "JPY", "KWD", "XYZ", "EUR", ""}
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 2000; i++ {
		base := int64(rng.Int63()) - int64(1<<62)
		code := codes[rng.Intn(len(codes))]
		m := MoneyFromBase(code, base)
		parsed, err := ParseMoney(m.String())
		require.NoError(t, err, "seed %d str %q", i, m.String())
		require.Equal(t, base, parsed.Base(), "seed %d", i)
		require.Equal(t, m.Code(), parsed.Code(), "seed %d", i)
	}
}

// ============================================================
// Benchmark
// ============================================================

func BenchmarkParseMoney(b *testing.B) {
	s := MoneyFromMinor("USD", 123456).String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseMoney(s); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMoneyAllocate(b *testing.B) {
	m := MoneyFromBase("USD", 12345678)
	ratios := []int64{5, 3, 2}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := m.Allocate(ratios); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatSymbolMinor(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FormatSymbolMinor(123456, "USD")
	}
}
