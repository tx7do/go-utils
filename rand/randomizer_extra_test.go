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
