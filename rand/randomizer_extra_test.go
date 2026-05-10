package rand

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ─── 构造器 ───────────────────────────────────────────────────────────────────

func TestNew_DefaultConfig(t *testing.T) {
	r := New()
	assert.NotNil(t, r)
	v := r.IntN(100)
	assert.GreaterOrEqual(t, v, 0)
	assert.Less(t, v, 100)
}

func TestNew_WithFixedSeed_Reproducible(t *testing.T) {
	r1 := New(WithFixedSeed(42))
	r2 := New(WithFixedSeed(42))
	results1 := make([]int64, 10)
	results2 := make([]int64, 10)
	for i := range results1 {
		results1[i] = r1.Int64N(1000)
		results2[i] = r2.Int64N(1000)
	}
	assert.Equal(t, results1, results2, "固定种子应产生相同序列")
}

func TestNew_WithRandType_AllTypes(t *testing.T) {
	types := []RandType{PCGRandType, ChaCha8RandType, SHA256RandType, ZipfRandType}
	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			r := New(WithRandType(typ), WithFixedSeed(1234))
			assert.NotNil(t, r)
			_ = r.Uint64()
		})
	}
}

func TestNew_WithZipfParams(t *testing.T) {
	r := New(WithRandType(ZipfRandType), WithZipfParams(1.5, 2.0, 50), WithFixedSeed(99))
	assert.NotNil(t, r)
	for i := 0; i < 100; i++ {
		v := r.ZipfUint64()
		assert.LessOrEqual(t, v, uint64(50))
	}
}

func TestNew_WithSecureMode(t *testing.T) {
	r := New(WithSecureMode())
	assert.NotNil(t, r)
}

func TestNewRandomizerWithSeed_Reproducible(t *testing.T) {
	r1 := NewRandomizerWithSeed(PCGRandType, 9999)
	r2 := NewRandomizerWithSeed(PCGRandType, 9999)
	assert.Equal(t, r1.Int64(), r2.Int64())
}

func TestDefault_NotNil(t *testing.T) {
	r := Default()
	assert.NotNil(t, r)
	_ = r.IntN(10)
}

// ─── SecureInt64 ─────────────────────────────────────────────────────────────

func TestRandomizer_SecureInt64_NormalMode(t *testing.T) {
	r := New() // secureMode=false
	v := r.SecureInt64()
	_ = v // 仅验证不 panic
}

func TestRandomizer_SecureInt64_SecureMode(t *testing.T) {
	r := New(WithSecureMode())
	// 连续取 10 个值应有差异（极小概率碰撞）
	vals := make(map[int64]struct{}, 10)
	for i := 0; i < 10; i++ {
		vals[r.SecureInt64()] = struct{}{}
	}
	assert.Greater(t, len(vals), 1)
}

// ─── 字符串系列 ───────────────────────────────────────────────────────────────

func TestRandomizer_StringWithCharset(t *testing.T) {
	r := newTestRandomizer()
	charset := "abc"
	result := r.StringWithCharset(20, charset)
	assert.Len(t, result, 20)
	for _, ch := range result {
		assert.Contains(t, charset, string(ch))
	}
}

func TestRandomizer_Hex(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct{ length int }{{0}, {8}, {32}}
	for _, tt := range tests {
		result := r.Hex(tt.length)
		assert.Len(t, result, tt.length)
		for _, ch := range result {
			assert.Contains(t, "0123456789abcdef", string(ch))
		}
	}
}

func TestRandomizer_LetterString(t *testing.T) {
	r := newTestRandomizer()
	result := r.LetterString(50)
	assert.Len(t, result, 50)
	for _, ch := range result {
		assert.True(t, (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'))
	}
}

// ─── TimeBetween ─────────────────────────────────────────────────────────────

func TestRandomizer_TimeBetween(t *testing.T) {
	r := newTestRandomizer()
	now := time.Now()
	future := now.Add(24 * time.Hour)

	for i := 0; i < 100; i++ {
		v := r.TimeBetween(now, future)
		assert.False(t, v.Before(now))
		assert.False(t, v.After(future))
	}
}

func TestRandomizer_TimeBetween_ReversedReturnsMax(t *testing.T) {
	r := newTestRandomizer()
	now := time.Now()
	past := now.Add(-time.Hour)
	v := r.TimeBetween(now, past) // min > max
	assert.Equal(t, past, v)
}

// ─── Bool 系列 ────────────────────────────────────────────────────────────────

func TestRandomizer_Bool_Distribution(t *testing.T) {
	r := newTestRandomizer()
	trueCount := 0
	for i := 0; i < 2000; i++ {
		if r.Bool() {
			trueCount++
		}
	}
	// 期望 ~50%，5σ 内不应跑偏超过 20%
	assert.Greater(t, trueCount, 600)
	assert.Less(t, trueCount, 1400)
}

func TestRandomizer_WeightedBool(t *testing.T) {
	r := newTestRandomizer()

	// 100% true
	for i := 0; i < 100; i++ {
		assert.True(t, r.WeightedBool(1, 0))
	}
	// 100% false
	for i := 0; i < 100; i++ {
		assert.False(t, r.WeightedBool(0, 1))
	}
	// total=0 → false
	assert.False(t, r.WeightedBool(0, 0))

	// 混合：true 应命中
	hitTrue := 0
	for i := 0; i < 1000; i++ {
		if r.WeightedBool(7, 3) {
			hitTrue++
		}
	}
	assert.Greater(t, hitTrue, 400)
}

// ─── Pick 系列 ────────────────────────────────────────────────────────────────

func TestRandomizer_Pick(t *testing.T) {
	r := newTestRandomizer()
	assert.Nil(t, r.Pick([]any{}))
	assert.Nil(t, r.Pick(nil))

	items := []any{1, "hello", 3.14}
	for i := 0; i < 100; i++ {
		v := r.Pick(items)
		assert.Contains(t, items, v)
	}
}

func TestRandomizer_PickString(t *testing.T) {
	r := newTestRandomizer()
	assert.Equal(t, "", r.PickString(nil))
	assert.Equal(t, "", r.PickString([]string{}))

	list := []string{"a", "b", "c"}
	for i := 0; i < 100; i++ {
		v := r.PickString(list)
		assert.Contains(t, list, v)
	}
}

func TestRandomizer_PickInt(t *testing.T) {
	r := newTestRandomizer()
	assert.Equal(t, 0, r.PickInt(nil))
	assert.Equal(t, 0, r.PickInt([]int{}))

	list := []int{10, 20, 30}
	for i := 0; i < 100; i++ {
		v := r.PickInt(list)
		assert.Contains(t, list, v)
	}
}

// ─── UUID ─────────────────────────────────────────────────────────────────────

func TestRandomizer_UUID(t *testing.T) {
	r := newTestRandomizer()
	id := r.UUID()
	// UUID v4 格式: 8-4-4-4-12
	parts := strings.Split(id, "-")
	assert.Len(t, parts, 5)
	assert.Len(t, parts[0], 8)
	assert.Len(t, parts[1], 4)
	assert.Len(t, parts[2], 4)
	assert.Len(t, parts[3], 4)
	assert.Len(t, parts[4], 12)

	id2 := r.UUID()
	assert.NotEqual(t, id, id2)
}

// ─── 伪数据生成器 ──────────────────────────────────────────────────────────────

func TestRandomizer_PhoneNumber(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 20; i++ {
		p := r.PhoneNumber()
		assert.Len(t, p, 11)
		// 前3位为合法号段（数字）
		for _, ch := range p[:3] {
			assert.True(t, ch >= '0' && ch <= '9')
		}
	}
}

func TestRandomizer_IDCard(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 20; i++ {
		id := r.IDCard()
		assert.Len(t, id, 18)
	}
}

func TestRandomizer_Email(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 20; i++ {
		email := r.Email()
		assert.Contains(t, email, "@")
		assert.Contains(t, email, ".")
	}
}

// ─── 颜色 ─────────────────────────────────────────────────────────────────────

func TestRandomizer_RandomColorHex(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 20; i++ {
		hex := r.RandomColorHex()
		assert.Len(t, hex, 7)
		assert.Equal(t, '#', rune(hex[0]))
	}
}

func TestRandomizer_RandomColorRGB(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 100; i++ {
		rr, gg, bb := r.RandomColorRGB()
		assert.GreaterOrEqual(t, rr, 0)
		assert.Less(t, rr, 256)
		assert.GreaterOrEqual(t, gg, 0)
		assert.Less(t, gg, 256)
		assert.GreaterOrEqual(t, bb, 0)
		assert.Less(t, bb, 256)
	}
}

// ─── 路径 / 网络 ──────────────────────────────────────────────────────────────

func TestRandomizer_RandomFilePath(t *testing.T) {
	r := newTestRandomizer()
	linux := r.RandomFilePath(false)
	assert.True(t, strings.HasPrefix(linux, "/"))

	win := r.RandomFilePath(true)
	assert.True(t, strings.HasPrefix(win, "C:\\"))
}

func TestRandomizer_RandomIPv4(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 20; i++ {
		ip := r.RandomIPv4()
		parts := strings.Split(ip, ".")
		assert.Len(t, parts, 4)
	}
}

func TestRandomizer_RandomIPv6(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 20; i++ {
		ip := r.RandomIPv6()
		parts := strings.Split(ip, ":")
		assert.Len(t, parts, 8)
	}
}

func TestRandomizer_RandomLatLng(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 100; i++ {
		lat, lng := r.RandomLatLng()
		assert.GreaterOrEqual(t, lat, -90.0)
		assert.LessOrEqual(t, lat, 90.0)
		assert.GreaterOrEqual(t, lng, -180.0)
		assert.LessOrEqual(t, lng, 180.0)
	}
}

// ─── 概率 / 抖动 ──────────────────────────────────────────────────────────────

func TestRandomizer_ProbabilityHit(t *testing.T) {
	r := newTestRandomizer()

	// 边界
	assert.False(t, r.ProbabilityHit(0))
	assert.True(t, r.ProbabilityHit(100))
	assert.False(t, r.ProbabilityHit(-5))

	// 50% 概率：期望命中 300~700 次
	hits := 0
	for i := 0; i < 1000; i++ {
		if r.ProbabilityHit(50) {
			hits++
		}
	}
	assert.Greater(t, hits, 300)
	assert.Less(t, hits, 700)
}

func TestRandomizer_JitterDuration(t *testing.T) {
	r := newTestRandomizer()
	base := time.Second

	// 无抖动
	assert.Equal(t, base, r.JitterDuration(base, 0))
	assert.Equal(t, base, r.JitterDuration(base, -10))

	// 20% 抖动：结果应在 [0.8s, 1.2s]
	for i := 0; i < 200; i++ {
		v := r.JitterDuration(base, 20)
		assert.GreaterOrEqual(t, v, time.Duration(float64(base)*0.79))
		assert.LessOrEqual(t, v, time.Duration(float64(base)*1.21))
	}
}

func TestRandomizer_JitterInt(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.JitterInt(100, 20)
		assert.GreaterOrEqual(t, v, 79)
		assert.LessOrEqual(t, v, 121)
	}
}

// ─── 游戏骰子 ─────────────────────────────────────────────────────────────────

func TestRandomizer_Dice(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		fn   func() int
		min  int
		max  int
	}{
		{"D6", r.D6, 1, 6},
		{"D10", r.D10, 1, 10},
		{"D20", r.D20, 1, 20},
		{"D100", r.D100, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 500; i++ {
				v := tt.fn()
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomizer_Roll(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 500; i++ {
		v := r.Roll(5, 15)
		assert.GreaterOrEqual(t, v, 5)
		assert.LessOrEqual(t, v, 15)
	}
}

func TestRandomizer_CritHit(t *testing.T) {
	r := newTestRandomizer()
	assert.False(t, r.CritHit(0))
	assert.True(t, r.CritHit(100))
}

func TestRandomizer_RandomSign(t *testing.T) {
	r := newTestRandomizer()
	got := make(map[int]bool)
	for i := 0; i < 200; i++ {
		s := r.RandomSign()
		assert.True(t, s == 1 || s == -1)
		got[s] = true
	}
	assert.True(t, got[1] && got[-1], "应同时出现 +1 和 -1")
}

// ─── FloatOffset / IntOffset ──────────────────────────────────────────────────

func TestRandomizer_FloatOffset(t *testing.T) {
	r := newTestRandomizer()
	base := 100.0
	for i := 0; i < 300; i++ {
		v := r.FloatOffset(base, 10)
		assert.GreaterOrEqual(t, v, 89.0)
		assert.LessOrEqual(t, v, 111.0)
	}
}

func TestRandomizer_IntOffset(t *testing.T) {
	r := newTestRandomizer()
	base := 100
	for i := 0; i < 300; i++ {
		v := r.IntOffset(base, 10)
		assert.GreaterOrEqual(t, v, 89)
		assert.LessOrEqual(t, v, 111)
	}
}

// ─── 角度 / 几何 ──────────────────────────────────────────────────────────────

func TestRandomizer_RandomAngle(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomAngle()
		assert.GreaterOrEqual(t, v, 0.0)
		assert.LessOrEqual(t, v, 360.0)
	}
}

func TestRandomizer_RandomAngleRad(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomAngleRad()
		assert.GreaterOrEqual(t, v, 0.0)
		assert.LessOrEqual(t, v, 2*math.Pi)
	}
}

func TestRandomizer_CircleRandomPoint(t *testing.T) {
	r := newTestRandomizer()
	cx, cy, radius := 5.0, 5.0, 3.0
	for i := 0; i < 200; i++ {
		x, y := r.CircleRandomPoint(cx, cy, radius)
		dist := math.Sqrt((x-cx)*(x-cx) + (y-cy)*(y-cy))
		assert.LessOrEqual(t, dist, radius+1e-9)
	}
}

func TestRandomizer_RectRandomPoint(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		x, y := r.RectRandomPoint(1.0, 1.0, 5.0, 5.0)
		assert.GreaterOrEqual(t, x, 1.0)
		assert.LessOrEqual(t, x, 5.0)
		assert.GreaterOrEqual(t, y, 1.0)
		assert.LessOrEqual(t, y, 5.0)
	}
}

// ─── 属性分配 ─────────────────────────────────────────────────────────────────

func TestRandomizer_AttrAssign(t *testing.T) {
	r := newTestRandomizer()

	// 空属性返回 nil
	assert.Nil(t, r.AttrAssign(100, 0))
	assert.Nil(t, r.AttrAssign(100, -1))

	for i := 0; i < 100; i++ {
		attrs := r.AttrAssign(100, 5)
		assert.Len(t, attrs, 5)
		sum := 0
		for _, v := range attrs {
			assert.GreaterOrEqual(t, v, 0)
			sum += v
		}
		assert.Equal(t, 100, sum, "总点数应等于 totalPoint")
	}
}

// ─── 保底概率 ─────────────────────────────────────────────────────────────────

func TestRandomizer_GuaranteedProb(t *testing.T) {
	r := newTestRandomizer()

	// failTimes >= limit 必中
	assert.True(t, r.GuaranteedProb(50, 10, 10))
	assert.True(t, r.GuaranteedProb(50, 20, 10))

	// rate=100 必中
	assert.True(t, r.GuaranteedProb(100, 0, 10))
}

func TestRandomizer_RandomRateMul(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomRateMul(0.5, 2.0)
		assert.GreaterOrEqual(t, v, 0.5)
		assert.LessOrEqual(t, v, 2.0)
	}
}

// ─── RandomPickUnique 系列 ────────────────────────────────────────────────────

func TestRandomizer_RandomPickUniqueStr(t *testing.T) {
	r := newTestRandomizer()

	assert.Equal(t, []string{}, r.RandomPickUniqueStr(nil, 3))
	assert.Equal(t, []string{}, r.RandomPickUniqueStr([]string{}, 3))
	assert.Equal(t, []string{}, r.RandomPickUniqueStr([]string{"a"}, 0))

	list := []string{"a", "b", "c", "d", "e"}

	// count > len => 全部返回
	result := r.RandomPickUniqueStr(list, 10)
	assert.Len(t, result, 5)
	assert.ElementsMatch(t, list, result)

	// count < len => 返回指定数量，无重复
	result = r.RandomPickUniqueStr(list, 3)
	assert.Len(t, result, 3)
	seen := make(map[string]struct{})
	for _, v := range result {
		assert.Contains(t, list, v)
		seen[v] = struct{}{}
	}
	assert.Len(t, seen, 3)

	// 原切片不被修改
	assert.ElementsMatch(t, []string{"a", "b", "c", "d", "e"}, list)
}

func TestRandomizer_RandomPickUniqueInt(t *testing.T) {
	r := newTestRandomizer()

	assert.Equal(t, []int{}, r.RandomPickUniqueInt(nil, 3))
	assert.Equal(t, []int{}, r.RandomPickUniqueInt([]int{}, 3))

	list := []int{10, 20, 30, 40, 50}
	result := r.RandomPickUniqueInt(list, 3)
	assert.Len(t, result, 3)

	seen := make(map[int]struct{})
	for _, v := range result {
		assert.Contains(t, list, v)
		seen[v] = struct{}{}
	}
	assert.Len(t, seen, 3)
}

func TestRandomizer_RandomPickUnique(t *testing.T) {
	r := newTestRandomizer()

	assert.Equal(t, []any{}, r.RandomPickUnique(nil, 3))
	assert.Equal(t, []any{}, r.RandomPickUnique([]any{}, 3))

	list := []any{1, 2, 3, 4, 5}
	result := r.RandomPickUnique(list, 3)
	assert.Len(t, result, 3)

	seen := make(map[any]struct{})
	for _, v := range result {
		assert.Contains(t, list, v)
		seen[v] = struct{}{}
	}
	assert.Len(t, seen, 3)
}

// ─── options 功能验证 ─────────────────────────────────────────────────────────

func TestOption_WithSeedType_MapHash(t *testing.T) {
	r := New(WithSeedType(MapHashSeed))
	assert.NotNil(t, r)
	_ = r.IntN(10)
}

func TestOption_WithSeedType_FixedSeed_DefaultZero(t *testing.T) {
	// FixedSeed 且 seed=0 时，cfg.seed==0 走 autoSeed 路径（因为 withFixedSeed 判断 seed!=0）
	r := New(WithSeedType(FixedSeed))
	assert.NotNil(t, r)
}

// ─── 基础原始方法 ─────────────────────────────────────────────────────────────

func TestRandomizer_PrimitiveTypes(t *testing.T) {
	r := newTestRandomizer()

	// Int / Int32 / Uint32
	_ = r.Int()
	_ = r.Int32()
	_ = r.Uint32()

	// UintN / Uint32N / Uint64N
	v := r.UintN(100)
	assert.Less(t, v, uint(100))
	v32 := r.Uint32N(100)
	assert.Less(t, v32, uint32(100))
	v64 := r.Uint64N(100)
	assert.Less(t, v64, uint64(100))
}

// ─── RangeInt32 / RangeInt64 / RangeUint 正常区间路径 ────────────────────────

func TestRandomizer_RangeInt32_NormalRange(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 500; i++ {
		v := r.RangeInt32(-10, 10)
		assert.GreaterOrEqual(t, v, int32(-10))
		assert.LessOrEqual(t, v, int32(10))
	}
}

func TestRandomizer_RangeInt64_NormalRange(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 500; i++ {
		v := r.RangeInt64(-100, 100)
		assert.GreaterOrEqual(t, v, int64(-100))
		assert.LessOrEqual(t, v, int64(100))
	}
}

func TestRandomizer_RangeUint_NormalRange(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 500; i++ {
		v := r.RangeUint(1, 50)
		assert.GreaterOrEqual(t, v, uint(1))
		assert.LessOrEqual(t, v, uint(50))
	}
}

// ─── NonWeightedChoice 正常路径 ───────────────────────────────────────────────

func TestRandomizer_NonWeightedChoice_NormalRange(t *testing.T) {
	r := newTestRandomizer()
	weights := []int{1, 2, 3, 4, 5}
	counts := make([]int, len(weights))
	for i := 0; i < 1000; i++ {
		idx := r.NonWeightedChoice(weights)
		assert.GreaterOrEqual(t, idx, 0)
		assert.Less(t, idx, len(weights))
		counts[idx]++
	}
	// 权重最大的最后一个应被选中更多
	assert.Greater(t, counts[4], counts[0])
}

// ─── RandomPickUnique 边界覆盖 ────────────────────────────────────────────────

func TestRandomizer_RandomPickUniqueInt_CountGTLen(t *testing.T) {
	r := newTestRandomizer()
	list := []int{1, 2, 3}
	result := r.RandomPickUniqueInt(list, 10)
	assert.Len(t, result, 3)
	assert.ElementsMatch(t, list, result)
}

func TestRandomizer_RandomPickUnique_CountGTLen(t *testing.T) {
	r := newTestRandomizer()
	list := []any{1, "two", 3.0}
	result := r.RandomPickUnique(list, 10)
	assert.Len(t, result, 3)
	assert.ElementsMatch(t, list, result)
}

// ─── ProbPermilleHit ─────────────────────────────────────────────────────────

func TestRandomizer_ProbPermilleHit(t *testing.T) {
	r := newTestRandomizer()

	assert.False(t, r.ProbPermilleHit(0))
	assert.False(t, r.ProbPermilleHit(-1))
	assert.True(t, r.ProbPermilleHit(1000))
	assert.True(t, r.ProbPermilleHit(1001))

	hits := 0
	for i := 0; i < 2000; i++ {
		if r.ProbPermilleHit(500) {
			hits++
		}
	}
	assert.Greater(t, hits, 600)
	assert.Less(t, hits, 1400)
}

// ─── 游戏/领域辅助方法 ────────────────────────────────────────────────────────

func TestRandomizer_RandomGrade(t *testing.T) {
	r := newTestRandomizer()
	weights := []int{5, 3, 2}
	for i := 0; i < 200; i++ {
		idx := r.RandomGrade(weights)
		assert.GreaterOrEqual(t, idx, 0)
		assert.Less(t, idx, len(weights))
	}
}

func TestRandomizer_RandomTwoChoice(t *testing.T) {
	r := newTestRandomizer()

	// total=0 → false
	assert.False(t, r.RandomTwoChoice(0, 0))

	// w1 全权重 → 全 true
	for i := 0; i < 100; i++ {
		assert.True(t, r.RandomTwoChoice(10, 0))
	}
	// w2 全权重 → 全 false
	for i := 0; i < 100; i++ {
		assert.False(t, r.RandomTwoChoice(0, 10))
	}

	// 混合
	trueCount := 0
	for i := 0; i < 1000; i++ {
		if r.RandomTwoChoice(7, 3) {
			trueCount++
		}
	}
	assert.Greater(t, trueCount, 400)
}

func TestRandomizer_RandomAttrValue(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomAttrValue(100, 0.8, 1.2)
		assert.GreaterOrEqual(t, v, 79)
		assert.LessOrEqual(t, v, 121)
	}
}

func TestRandomizer_RandomInterval(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomInterval(10, 3)
		assert.GreaterOrEqual(t, v, 7)
		assert.LessOrEqual(t, v, 13)
	}
}

func TestRandomizer_RandomSleepMs(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomSleepMs(100, 500)
		assert.GreaterOrEqual(t, v, 100)
		assert.LessOrEqual(t, v, 500)
	}
}

func TestRandomizer_RandomDir4(t *testing.T) {
	r := newTestRandomizer()
	dirs := make(map[int]bool)
	for i := 0; i < 500; i++ {
		d := r.RandomDir4()
		assert.GreaterOrEqual(t, d, 0)
		assert.Less(t, d, 4)
		dirs[d] = true
	}
	assert.Len(t, dirs, 4, "4个方向应全部出现")
}

func TestRandomizer_RandomDir8(t *testing.T) {
	r := newTestRandomizer()
	dirs := make(map[int]bool)
	for i := 0; i < 1000; i++ {
		d := r.RandomDir8()
		assert.GreaterOrEqual(t, d, 0)
		assert.Less(t, d, 8)
		dirs[d] = true
	}
	assert.Len(t, dirs, 8, "8个方向应全部出现")
}

func TestRandomizer_RandomBoolRatio(t *testing.T) {
	r := newTestRandomizer()
	assert.False(t, r.RandomBoolRatio(0))
	assert.True(t, r.RandomBoolRatio(100))

	hits := 0
	for i := 0; i < 1000; i++ {
		if r.RandomBoolRatio(50) {
			hits++
		}
	}
	assert.Greater(t, hits, 300)
	assert.Less(t, hits, 700)
}

func TestRandomizer_RandomIndex(t *testing.T) {
	r := newTestRandomizer()
	assert.Equal(t, 0, r.RandomIndex(0))
	assert.Equal(t, 0, r.RandomIndex(-1))

	for i := 0; i < 500; i++ {
		idx := r.RandomIndex(10)
		assert.GreaterOrEqual(t, idx, 0)
		assert.Less(t, idx, 10)
	}
}

func TestRandomizer_LuckyValue(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.LuckyValue()
		assert.GreaterOrEqual(t, v, 1)
		assert.LessOrEqual(t, v, 100)
	}
}

func TestRandomizer_CritHitWithLucky(t *testing.T) {
	r := newTestRandomizer()

	// rate 溢出 100 仍应 true
	assert.True(t, r.CritHitWithLucky(50, 100, 1.0)) // 50+100=150>100

	// base 0, lucky 0, bonus 0 → 永不暴击
	assert.False(t, r.CritHitWithLucky(0, 0, 0))

	// 正常命中率分布
	hits := 0
	for i := 0; i < 1000; i++ {
		if r.CritHitWithLucky(30, 20, 0.5) { // rate = 30+10 = 40%
			hits++
		}
	}
	assert.Greater(t, hits, 150)
	assert.Less(t, hits, 650)
}

func TestRandomizer_ShrinkIntSlice(t *testing.T) {
	r := newTestRandomizer()

	// dropCnt=0 → 空
	assert.Equal(t, []int{}, r.ShrinkIntSlice([]int{1, 2, 3}, 0))
	// dropCnt >= len → 空
	assert.Equal(t, []int{}, r.ShrinkIntSlice([]int{1, 2, 3}, 3))
	assert.Equal(t, []int{}, r.ShrinkIntSlice([]int{1, 2, 3}, 5))

	arr := []int{1, 2, 3, 4, 5}
	result := r.ShrinkIntSlice(arr, 2)
	assert.Len(t, result, 3)
	for _, v := range result {
		assert.Contains(t, arr, v)
	}
	// 原数组不变
	assert.Equal(t, []int{1, 2, 3, 4, 5}, arr)
}

func TestRandomizer_RoundFloat(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RoundFloat(1.0, 5.0, 2)
		assert.GreaterOrEqual(t, v, 1.0)
		assert.LessOrEqual(t, v, 5.0)
		// 保留2位小数
		rounded := math.Round(v*100) / 100
		assert.InDelta(t, rounded, v, 1e-9)
	}
}

func TestRandomizer_RandomContinuousTimes(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomContinuousTimes(10)
		assert.GreaterOrEqual(t, v, 1)
		assert.LessOrEqual(t, v, 10)
	}
}

func TestRandomizer_SlippagePrice(t *testing.T) {
	r := newTestRandomizer()
	price := 100.0
	rate := 0.01 // 1%
	for i := 0; i < 200; i++ {
		v := r.SlippagePrice(price, rate)
		assert.GreaterOrEqual(t, v, price*(1-rate)-1e-9)
		assert.LessOrEqual(t, v, price*(1+rate)+1e-9)
	}
}

func TestRandomizer_RandomPositionRate(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomPositionRate()
		assert.GreaterOrEqual(t, v, 0.0)
		assert.LessOrEqual(t, v, 1.0)
	}
}

func TestRandomizer_RandomPositionRateRange(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomPositionRateRange(0.2, 0.8)
		assert.GreaterOrEqual(t, v, 0.2)
		assert.LessOrEqual(t, v, 0.8)
	}
}

func TestRandomizer_RandomVolatility(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.RandomVolatility(0.01, 0.05)
		assert.GreaterOrEqual(t, v, 0.01)
		assert.LessOrEqual(t, v, 0.05)
	}
}

func TestRandomizer_RandomReturn(t *testing.T) {
	r := newTestRandomizer()
	v := r.RandomReturn(0.001, 0.02)
	assert.False(t, math.IsNaN(v))
	assert.False(t, math.IsInf(v, 0))
}

func TestRandomizer_TruncatedNormal(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 200; i++ {
		v := r.TruncatedNormal(0, 1, -2, 2)
		assert.GreaterOrEqual(t, v, -2.0)
		assert.LessOrEqual(t, v, 2.0)
	}
}

func TestRandomizer_LogNormal(t *testing.T) {
	r := newTestRandomizer()
	for i := 0; i < 100; i++ {
		v := r.LogNormal(0, 0.1)
		assert.Greater(t, v, 0.0)
		assert.False(t, math.IsNaN(v))
	}
}

func TestRandomizer_ProbWithCooldown(t *testing.T) {
	r := newTestRandomizer()

	// 冷却未满 → 始终 false
	for i := 0; i < 100; i++ {
		assert.False(t, r.ProbWithCooldown(100, 5, 3))
	}
	// 冷却已满、100% → 始终 true
	assert.True(t, r.ProbWithCooldown(100, 5, 5))
	// 冷却已满、0% → 始终 false
	assert.False(t, r.ProbWithCooldown(0, 5, 5))
}

func TestRandomizer_ShuffleGrid(t *testing.T) {
	r := newTestRandomizer()
	width, height := 4, 3
	grid := r.ShuffleGrid(width, height)
	assert.Len(t, grid, height)

	seen := make(map[[2]int]bool)
	for _, row := range grid {
		assert.Len(t, row, width)
		for _, cell := range row {
			key := [2]int{cell.X, cell.Y}
			assert.False(t, seen[key], "格子(%d,%d)不应重复", cell.X, cell.Y)
			seen[key] = true
			assert.GreaterOrEqual(t, cell.X, 0)
			assert.Less(t, cell.X, width)
			assert.GreaterOrEqual(t, cell.Y, 0)
			assert.Less(t, cell.Y, height)
		}
	}
	assert.Len(t, seen, width*height)
}

// ─── AliasTable Float32 / Float64 变体 ───────────────────────────────────────

func TestRandomizer_NewAliasTableFloat32(t *testing.T) {
	r := newTestRandomizer()

	// 空权重 → nil
	assert.Nil(t, r.NewAliasTableFloat32(nil))
	assert.Nil(t, r.NewAliasTableFloat32([]float32{}))
	// 全无效权重 → nil
	assert.Nil(t, r.NewAliasTableFloat32([]float32{0, -1, 0.00000001}))

	// 正常权重
	weights := []float32{5.0, 3.0, 2.0}
	at := r.NewAliasTableFloat32(weights)
	assert.NotNil(t, at)
	assert.Len(t, at.prob, 3)
	assert.Len(t, at.alias, 3)
	assert.Len(t, at.origIdx, 3)

	// 选择结果应在原始索引范围内
	counts := make(map[int]int)
	for i := 0; i < 1000; i++ {
		idx := r.AliasChoice(at)
		assert.GreaterOrEqual(t, idx, 0)
		assert.Less(t, idx, len(weights))
		counts[idx]++
	}
	// 权重最大的索引 0 应被选中最多
	assert.Greater(t, counts[0], counts[2])
}

func TestRandomizer_NewAliasTableFloat32_WithInvalidEntries(t *testing.T) {
	r := newTestRandomizer()
	// 部分无效，仍应构建（保留原始索引映射）
	weights := []float32{0, 1.5, 0, 0.5}
	at := r.NewAliasTableFloat32(weights)
	assert.NotNil(t, at)
	assert.Len(t, at.prob, 2)
	assert.Len(t, at.origIdx, 2)
	// 有效索引应为 1 和 3
	assert.ElementsMatch(t, []int{1, 3}, at.origIdx)

	for i := 0; i < 200; i++ {
		idx := r.AliasChoice(at)
		assert.Contains(t, []int{1, 3}, idx)
	}
}

func TestRandomizer_NewAliasTableFloat64(t *testing.T) {
	r := newTestRandomizer()

	// 空权重 → nil
	assert.Nil(t, r.NewAliasTableFloat64(nil))
	assert.Nil(t, r.NewAliasTableFloat64([]float64{}))
	// 全无效权重 → nil
	assert.Nil(t, r.NewAliasTableFloat64([]float64{0, -1, 0.00000001}))

	// 正常权重
	weights := []float64{5.0, 3.0, 2.0}
	at := r.NewAliasTableFloat64(weights)
	assert.NotNil(t, at)
	assert.Len(t, at.prob, 3)
	assert.Len(t, at.alias, 3)
	assert.Len(t, at.origIdx, 3)

	counts := make(map[int]int)
	for i := 0; i < 1000; i++ {
		idx := r.AliasChoice(at)
		assert.GreaterOrEqual(t, idx, 0)
		assert.Less(t, idx, len(weights))
		counts[idx]++
	}
	assert.Greater(t, counts[0], counts[2])
}

func TestRandomizer_NewAliasTableFloat64_WithInvalidEntries(t *testing.T) {
	r := newTestRandomizer()
	weights := []float64{0, 2.0, 0, 1.0}
	at := r.NewAliasTableFloat64(weights)
	assert.NotNil(t, at)
	assert.Len(t, at.prob, 2)
	assert.ElementsMatch(t, []int{1, 3}, at.origIdx)

	for i := 0; i < 200; i++ {
		idx := r.AliasChoice(at)
		assert.Contains(t, []int{1, 3}, idx)
	}
}
