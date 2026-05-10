package rand

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	cryptoRand "crypto/rand"
	mathRand "math/rand/v2"

	"github.com/google/uuid"
	utilMath "github.com/tx7do/go-utils/math"
)

// RandType 随机数生成器类型枚举
type RandType string

const (
	PCGRandType     RandType = "PCG"     // PCG 随机数生成器
	ChaCha8RandType RandType = "ChaCha8" // ChaCha8 随机数生成器
	ZipfRandType    RandType = "Zipf"    // Zipf 随机数生成器
	SHA256RandType  RandType = "SHA256"  // 基于 SHA-256 哈希的随机数生成器（测试用，非高性能）
)

type Randomizer struct {
	rnd  *mathRand.Rand
	zipf *mathRand.Zipf

	secureMode bool // 加密安全模式（使用 crypto/rand）
}

var defaultRandomizer = New()

func Default() *Randomizer {
	return defaultRandomizer
}

func New(opts ...Option) *Randomizer {
	// 加载默认配置
	cfg := defaultConfig()
	// 应用用户传入的选项
	for _, opt := range opts {
		opt(&cfg)
	}

	// 固定种子优先
	if cfg.seed != 0 {
		r := newWithFixedSeed(cfg)
		r.secureMode = cfg.secureMode
		return r
	}

	// 自动种子模式
	r := newWithAutoSeed(cfg)
	r.secureMode = cfg.secureMode
	return r
}

func NewRandomizer(randType RandType, seedType SeedType) *Randomizer {
	return New(WithRandType(randType), WithSeedType(seedType))
}

// NewRandomizerWithSeed 兼容旧 API
func NewRandomizerWithSeed(randType RandType, seed uint64) *Randomizer {
	return New(WithRandType(randType), WithFixedSeed(seed))
}

// NewZipfRandomizer 兼容旧 API
func NewZipfRandomizer(seedType SeedType, s, q float64, v uint64) *Randomizer {
	return New(
		WithRandType(ZipfRandType),
		WithSeedType(seedType),
		WithZipfParams(s, q, v),
	)
}

// newWithAutoSeed 自动生成种子
func newWithAutoSeed(cfg config) *Randomizer {
	seeder := NewSeeder(cfg.seedType)
	seed := uint64(seeder.Seed())

	switch cfg.randType {
	case ChaCha8RandType:
		var seedArr [32]byte
		binary.LittleEndian.PutUint64(seedArr[0:8], seed)
		_, _ = cryptoRand.Read(seedArr[8:32])
		return &Randomizer{rnd: mathRand.New(mathRand.NewChaCha8(seedArr))}

	case SHA256RandType:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, seed)
		return &Randomizer{rnd: mathRand.New(NewSha256Source(buf))}

	case ZipfRandType:
		src := mathRand.NewPCG(seed, seed^0x55AA)
		r := mathRand.New(src)
		z := mathRand.NewZipf(r, cfg.zipfS, cfg.zipfQ, cfg.zipfV)
		return &Randomizer{rnd: r, zipf: z}

	default: // PCG
		src := mathRand.NewPCG(seed, seed^0x55AA55AA55AA55AA)
		return &Randomizer{rnd: mathRand.New(src)}
	}
}

// newWithFixedSeed 使用固定种子
func newWithFixedSeed(cfg config) *Randomizer {
	seed := cfg.seed

	switch cfg.randType {
	case ChaCha8RandType:
		var seedArr [32]byte
		binary.LittleEndian.PutUint64(seedArr[0:8], seed)
		_, _ = cryptoRand.Read(seedArr[8:32])
		return &Randomizer{rnd: mathRand.New(mathRand.NewChaCha8(seedArr))}

	case SHA256RandType:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, seed)
		return &Randomizer{rnd: mathRand.New(NewSha256Source(buf))}

	case ZipfRandType:
		src := mathRand.NewPCG(seed, seed^0x55AA)
		r := mathRand.New(src)
		z := mathRand.NewZipf(r, cfg.zipfS, cfg.zipfQ, cfg.zipfV)
		return &Randomizer{rnd: r, zipf: z}

	default: // PCG
		src := mathRand.NewPCG(seed, seed^0x55AA55AA55AA55AA)
		return &Randomizer{rnd: mathRand.New(src)}
	}
}

func (r *Randomizer) Float32() float32 {
	return r.rnd.Float32()
}

func (r *Randomizer) Float64() float64 {
	return r.rnd.Float64()
}

func (r *Randomizer) Int() int {
	return r.rnd.Int()
}

func (r *Randomizer) Int32() int32 {
	return r.rnd.Int32()
}

func (r *Randomizer) Int64() int64 {
	return r.rnd.Int64()
}

func (r *Randomizer) Uint32() uint32 {
	return r.rnd.Uint32()
}

func (r *Randomizer) Uint64() uint64 {
	if r.zipf != nil {
		return r.zipf.Uint64()
	}

	return r.rnd.Uint64()
}

func (r *Randomizer) IntN(n int) int {
	return r.rnd.IntN(n)
}

func (r *Randomizer) Int32N(n int32) int32 {
	return r.rnd.Int32N(n)
}

func (r *Randomizer) Int64N(n int64) int64 {
	return r.rnd.Int64N(n)
}

func (r *Randomizer) UintN(n uint) uint {
	return r.rnd.UintN(n)
}

func (r *Randomizer) Uint32N(n uint32) uint32 {
	return r.rnd.Uint32N(n)
}

func (r *Randomizer) Uint64N(n uint64) uint64 {
	return r.rnd.Uint64N(n)
}

// RangeInt 根据区间产生随机数
func (r *Randomizer) RangeInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + r.rnd.IntN(max-min+1)
}

// RangeInt32 根据区间产生随机数
func (r *Randomizer) RangeInt32(min, max int32) int32 {
	if min >= max {
		return max
	}
	return min + r.Int32N(max-min+1)
}

// RangeInt64 根据区间产生随机数
func (r *Randomizer) RangeInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return min + r.Int64N(max-min+1)
}

// RangeUint 根据区间产生随机数
func (r *Randomizer) RangeUint(min, max uint) uint {
	if min >= max {
		return max
	}
	return min + uint(r.IntN(int(max-min+1)))
}

// RangeUint32 根据区间产生随机数
func (r *Randomizer) RangeUint32(min, max uint32) uint32 {
	if min >= max {
		return max
	}
	return min + uint32(r.Int32N(int32(max-min+1)))
}

// RangeUint64 根据区间产生随机数
func (r *Randomizer) RangeUint64(min, max uint64) uint64 {
	if min >= max {
		return max
	}
	return min + uint64(r.Int64N(int64(max-min+1)))
}

// RangeFloat32 根据区间产生随机数
func (r *Randomizer) RangeFloat32(min, max float32) float32 {
	if min >= max {
		return max
	}
	return min + r.Float32()*(max-min)
}

// RangeFloat64 根据区间产生随机数
func (r *Randomizer) RangeFloat64(min, max float64) float64 {
	if min >= max {
		return max
	}
	return min + r.Float64()*(max-min)
}

// SecureInt64 安全随机（crypto/rand）int64（密码级）
func (r *Randomizer) SecureInt64() int64 {
	if !r.secureMode {
		return r.Int64()
	}
	var b [8]byte
	_, _ = cryptoRand.Read(b[:])
	return int64(binary.LittleEndian.Uint64(b[:]))
}

// RandomString 随机字符串，包含大小写字母和数字
func (r *Randomizer) RandomString(l int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[r.rnd.IntN(len(charset))]
	}
	return string(b)
}

// StringWithCharset 自定义字符集随机字符串
func (r *Randomizer) StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.IntN(len(charset))]
	}
	return string(b)
}

// Hex 随机十六进制字符串（常用于ID、密钥）
func (r *Randomizer) Hex(length int) string {
	const hexChars = "0123456789abcdef"
	return r.StringWithCharset(length, hexChars)
}

// LetterString 随机大小写字母
func (r *Randomizer) LetterString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return r.StringWithCharset(length, letters)
}

// TimeBetween 随机时间（在两个时间之间）
func (r *Randomizer) TimeBetween(min, max time.Time) time.Time {
	if min.After(max) {
		return max
	}
	delta := max.Unix() - min.Unix()
	sec := r.Int64N(delta)
	return min.Add(time.Second * time.Duration(sec))
}

// WeightedChoice 根据权重随机，返回对应选项的索引，O(n)
func (r *Randomizer) WeightedChoice(weightArray []int) int {
	n := len(weightArray)
	if n == 0 {
		return -1
	}

	if n == 1 {
		if weightArray[0] > 0 {
			return 0
		}
		return -1
	}

	var total int64
	for _, w := range weightArray {
		if w > 0 {
			total += int64(w)
		}
	}

	if total <= 0 {
		return -1
	}

	rv := r.rnd.Int64N(total)
	var cursor int64
	for i, v := range weightArray {
		if v <= 0 {
			continue
		}
		cursor += int64(v)
		if rv < cursor {
			return i
		}
	}

	return n - 1
}

// NonWeightedChoice 根据权重随机，返回对应选项的索引，O(n)，但会将负权重视为0，并且在总权重为0时返回0
func (r *Randomizer) NonWeightedChoice(weightArray []int) int {
	if len(weightArray) == 0 {
		return -1
	}

	// 复制避免修改调用方传入的切片
	weights := make([]int, len(weightArray))
	copy(weights, weightArray)

	for i, weight := range weights {
		if weight < 0 {
			weights[i] = 0
		}
	}

	total := utilMath.SumInt(weights)
	if total <= 0 {
		return 0
	}

	rv := r.Int64N(total)
	for i, v := range weights {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weights) - 1
}

// WeightedBool 权重布尔
// trueWeight / falseWeight 权重
func (r *Randomizer) WeightedBool(trueWeight, falseWeight int) bool {
	total := trueWeight + falseWeight
	if total <= 0 {
		return false
	}
	rv := r.IntN(total)
	return rv < trueWeight
}

// Shuffle 洗牌算法，随机打乱元素顺序
func (r *Randomizer) Shuffle(n int, swap func(i, j int)) {
	r.rnd.Shuffle(n, swap)
}

// Normal 生成符合正态分布的随机数，参数 mean 是均值，stdDev 是标准差
func (r *Randomizer) Normal(mean, stdDev float64) float64 {
	return mean + r.rnd.NormFloat64()*stdDev
}

// Exp 生成符合指数分布的随机数，参数 lambda 是速率参数（lambda > 0）
func (r *Randomizer) Exp(lambda float64) float64 {
	return r.rnd.ExpFloat64() / lambda
}

// Perm 生成一个长度为 n 的随机排列，返回一个包含 0 到 n-1 的切片，顺序被随机打乱
func (r *Randomizer) Perm(n int) []int {
	return r.rnd.Perm(n)
}

// ZipfUint64 生成符合 Zipf 分布的随机数，如果未初始化 Zipf 生成器，则退回到普通随机数生成
func (r *Randomizer) ZipfUint64() uint64 {
	if r.zipf == nil {
		return r.rnd.Uint64()
	}
	return r.zipf.Uint64()
}

// Bool 随机布尔值 true/false
func (r *Randomizer) Bool() bool {
	return r.rnd.Int64()%2 == 0
}

// Pick 随机从切片里选一个元素
func (r *Randomizer) Pick(slice []any) any {
	if len(slice) == 0 {
		return nil
	}
	return slice[r.IntN(len(slice))]
}

// PickString 随机字符串切片元素
func (r *Randomizer) PickString(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[r.IntN(len(list))]
}

// PickInt 随机 int 切片元素
func (r *Randomizer) PickInt(list []int) int {
	if len(list) == 0 {
		return 0
	}
	return list[r.IntN(len(list))]
}

// RandomPickUniqueStr 从字符串切片随机选 count 个不重复元素
func (r *Randomizer) RandomPickUniqueStr(list []string, count int) []string {
	if count <= 0 || len(list) == 0 {
		return []string{}
	}
	if count >= len(list) {
		count = len(list)
	}

	// 复制切片避免修改原数组
	tmp := make([]string, len(list))
	copy(tmp, list)

	// 洗牌
	r.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	return tmp[:count]
}

// RandomPickUniqueInt 从 int 切片随机选 count 个不重复元素
func (r *Randomizer) RandomPickUniqueInt(list []int, count int) []int {
	if count <= 0 || len(list) == 0 {
		return []int{}
	}
	if count >= len(list) {
		count = len(list)
	}

	tmp := make([]int, len(list))
	copy(tmp, list)

	r.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	return tmp[:count]
}

// RandomPickUnique 通用任意类型（返回 any）
func (r *Randomizer) RandomPickUnique(list []any, count int) []any {
	if count <= 0 || len(list) == 0 {
		return []any{}
	}
	if count >= len(list) {
		count = len(list)
	}

	tmp := make([]any, len(list))
	copy(tmp, list)

	r.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	return tmp[:count]
}

func (r *Randomizer) UUID() string {
	return uuid.New().String()
}

// PhoneNumber 生成随机手机号（中国大陆常见号段）
func (r *Randomizer) PhoneNumber() string {
	// 国内常见号段
	prefixes := []string{
		"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"147", "150", "151", "152", "153", "155", "156", "157", "158", "159",
		"170", "176", "177", "178", "180", "181", "182", "183", "184", "185", "186", "187", "188", "189",
		"191", "192", "193", "195", "196", "197", "198", "199",
	}
	p := prefixes[r.IntN(len(prefixes))]
	suffix := r.RandomString(8)
	return p + suffix
}

// IDCard 生成随机身份证号码（简化版，非真实算法）
func (r *Randomizer) IDCard() string {
	// 简化：地区6位 + 生日8位 + 顺序3位 + 校验位
	area := fmt.Sprintf("%06d", r.IntN(999999))
	year := r.RangeInt(1950, 2005)
	month := r.RangeInt(1, 12)
	day := r.RangeInt(1, 28)
	birth := fmt.Sprintf("%04d%02d%02d", year, month, day)
	seq := fmt.Sprintf("%03d", r.IntN(999))
	last := r.StringWithCharset(1, "0123456789X")
	return area + birth + seq + last
}

// Email 生成随机邮箱地址
func (r *Randomizer) Email() string {
	domains := []string{"gmail.com", "qq.com", "163.com", "outlook.com", "icloud.com"}
	name := r.LetterString(r.RangeInt(6, 12))
	domain := domains[r.IntN(len(domains))]
	return name + "@" + domain
}

// RandomColorHex 随机 #RRGGBB
func (r *Randomizer) RandomColorHex() string {
	rr := r.IntN(256)
	gg := r.IntN(256)
	bb := r.IntN(256)
	return fmt.Sprintf("#%02X%02X%02X", rr, gg, bb)
}

// RandomColorRGB 随机 r,g,b 0-255
func (r *Randomizer) RandomColorRGB() (int, int, int) {
	return r.IntN(256), r.IntN(256), r.IntN(256)
}

// RandomFilePath Linux/Windows 随机路径
func (r *Randomizer) RandomFilePath(isWindows bool) string {
	names := []string{"data", "log", "temp", "cache", "config", "backup", "upload", "download"}
	ext := []string{"txt", "json", "csv", "jpg", "png", "bin", "log", "xml"}
	dir := names[r.IntN(len(names))]
	file := r.LetterString(8)
	e := ext[r.IntN(len(ext))]

	if isWindows {
		return fmt.Sprintf("C:\\%s\\%s.%s", dir, file, e)
	}
	return fmt.Sprintf("/%s/%s.%s", dir, file, e)
}

// RandomIPv4 随机 IPv4
func (r *Randomizer) RandomIPv4() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		r.RangeInt(1, 254),
		r.RangeInt(0, 255),
		r.RangeInt(0, 255),
		r.RangeInt(1, 254),
	)
}

// RandomIPv6 简易随机 IPv6
func (r *Randomizer) RandomIPv6() string {
	parts := make([]string, 8)
	for i := 0; i < 8; i++ {
		parts[i] = r.Hex(4)
	}
	return strings.Join(parts, ":")
}

// RandomLatLng 随机经纬度 纬度[-90,90] 经度[-180,180]
func (r *Randomizer) RandomLatLng() (float64, float64) {
	lat := r.RangeFloat64(-90.0, 90.0)
	lng := r.RangeFloat64(-180.0, 180.0)
	return math.Round(lat*10000) / 10000, math.Round(lng*10000) / 10000
}

// ProbabilityHit 百分比概率是否命中，p 传 0~100
func (r *Randomizer) ProbabilityHit(p float64) bool {
	if p <= 0 {
		return false
	}
	if p >= 100 {
		return true
	}
	return r.RangeFloat64(0, 100) < p
}

// ProbPermilleHit 随机是否触发（支持千分比，精度更高）
// permille 千分比 0~1000
func (r *Randomizer) ProbPermilleHit(permille int) bool {
	if permille <= 0 {
		return false
	}
	if permille >= 1000 {
		return true
	}
	return r.IntN(1000) < permille
}

// JitterDuration 在 base 基础上 ±jitterPercent 抖动
// 例：base=1s, jitterPercent=20 => 0.8s ~ 1.2s
func (r *Randomizer) JitterDuration(base time.Duration, jitterPercent float64) time.Duration {
	if jitterPercent <= 0 {
		return base
	}
	ratio := r.RangeFloat64(1-jitterPercent/100, 1+jitterPercent/100)
	return time.Duration(float64(base) * ratio)
}

// JitterInt 整数抖动，在 base 上下浮动 percent%
func (r *Randomizer) JitterInt(base int, percent float64) int {
	ratio := r.RangeFloat64(1-percent/100, 1+percent/100)
	return int(math.Round(float64(base) * ratio))
}

// Roll 游戏roll点 [min, max] 闭区间整数
func (r *Randomizer) Roll(min, max int) int {
	return r.RangeInt(min, max)
}

// D6 掷6面骰子 1~6
func (r *Randomizer) D6() int {
	return r.Roll(1, 6)
}

// D10 掷10面骰子 1~10
func (r *Randomizer) D10() int {
	return r.Roll(1, 10)
}

// D20 掷20面骰子 1~20
func (r *Randomizer) D20() int {
	return r.Roll(1, 20)
}

// D100 百分骰 1~100
func (r *Randomizer) D100() int {
	return r.Roll(1, 100)
}

// CritHit 暴击判定 rate 0~100
func (r *Randomizer) CritHit(rate float64) bool {
	return r.ProbabilityHit(rate)
}

// RandomSign 随机正负 +1 / -1
func (r *Randomizer) RandomSign() int {
	if r.Bool() {
		return 1
	}
	return -1
}

// FloatOffset 数值随机浮动
// base 基础值, percent 浮动百分比
func (r *Randomizer) FloatOffset(base float64, percent float64) float64 {
	offset := base * percent / 100
	return r.RangeFloat64(base-offset, base+offset)
}

// IntOffset 整数随机浮动
func (r *Randomizer) IntOffset(base int, percent float64) int {
	val := r.FloatOffset(float64(base), percent)
	return int(math.Round(val))
}

// RandomAngle 随机角度 0~360 度
func (r *Randomizer) RandomAngle() float64 {
	return r.RangeFloat64(0, 360)
}

// RandomAngleRad 随机弧度 0~2π
func (r *Randomizer) RandomAngleRad() float64 {
	return r.RangeFloat64(0, 2*math.Pi)
}

// CircleRandomPoint 圆心范围内随机点
// cx,cy 圆心, radius 半径
func (r *Randomizer) CircleRandomPoint(cx, cy, radius float64) (float64, float64) {
	angle := r.RandomAngleRad()
	rad := r.RangeFloat64(0, radius)
	x := cx + math.Cos(angle)*rad
	y := cy + math.Sin(angle)*rad
	return x, y
}

// RectRandomPoint 矩形内随机点
// x1,y1 左上角 x2,y2 右下角
func (r *Randomizer) RectRandomPoint(x1, y1, x2, y2 float64) (float64, float64) {
	x := r.RangeFloat64(math.Min(x1, x2), math.Max(x1, x2))
	y := r.RangeFloat64(math.Min(y1, y2), math.Max(y1, y2))
	return x, y
}

// AttrAssign 总点数固定，随机分配到多属性
// totalPoint 总点数, attrCount 属性个数
func (r *Randomizer) AttrAssign(totalPoint int, attrCount int) []int {
	if attrCount <= 0 {
		return nil
	}
	points := make([]int, attrCount)
	remain := totalPoint

	for i := 0; i < attrCount-1; i++ {
		alloc := r.Roll(0, remain)
		points[i] = alloc
		remain -= alloc
	}
	points[attrCount-1] = remain

	// 打乱，避免前面偏少后面偏多
	r.Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})
	return points
}

// RandomRateMul 随机倍率
// minMul 最小倍率 maxMul 最大倍率
func (r *Randomizer) RandomRateMul(minMul, maxMul float64) float64 {
	return r.RangeFloat64(minMul, maxMul)
}

// GuaranteedProb 带保底的概率判定
// rate 基础概率, failTimes 连续失败次数, limit 保底次数
func (r *Randomizer) GuaranteedProb(rate float64, failTimes, limit int) bool {
	// 达到保底必中
	if failTimes >= limit {
		return true
	}
	// 每失败一次小幅提升概率
	boost := float64(failTimes) * (100 - rate) / float64(limit)
	return r.ProbabilityHit(rate + boost)
}

// RandomGrade 随机档位区间（多用于装备品质、稀有度）
// 按权重分段，返回档位索引
func (r *Randomizer) RandomGrade(weights []int) int {
	return r.WeightedChoice(weights)
}

// RandomTwoChoice 随机二选一 带权重
func (r *Randomizer) RandomTwoChoice(w1, w2 int) bool {
	total := w1 + w2
	if total <= 0 {
		return false
	}
	return r.IntN(total) < w1
}

// RandomAttrValue 随机浮动数值 保留整数（装备词条浮动）
// base 基础值, minRate 最低倍率, maxRate 最高倍率
func (r *Randomizer) RandomAttrValue(base int, minRate, maxRate float64) int {
	f := r.RangeFloat64(minRate, maxRate)
	return int(float64(base) * f)
}

// RandomInterval 随机生成递增间隔（怪物刷新、技能CD随机打散）
// base 基础间隔, jitter 上下浮动秒数
func (r *Randomizer) RandomInterval(base int, jitter int) int {
	return r.RangeInt(base-jitter, base+jitter)
}

// RandomSleepMs 随机等待毫秒（协程休眠、行为随机延时）
func (r *Randomizer) RandomSleepMs(min, max int) int {
	return r.RangeInt(min, max)
}

// RandomDir4 随机方向 上下左右 4方向
func (r *Randomizer) RandomDir4() int {
	// 0上 1右 2下 3左
	return r.IntN(4)
}

// RandomDir8 随机方向 8方向
func (r *Randomizer) RandomDir8() int {
	// 0~7 八个方向
	return r.IntN(8)
}

// RandomBoolRatio 随机布尔带权重百分比
// ratio 0~100
func (r *Randomizer) RandomBoolRatio(ratio float64) bool {
	return r.ProbabilityHit(ratio)
}

// RandomIndex 从数组随机取一个索引
func (r *Randomizer) RandomIndex(size int) int {
	if size <= 0 {
		return 0
	}
	return r.IntN(size)
}

// LuckyValue 随机生成幸运值 1~100
func (r *Randomizer) LuckyValue() int {
	return r.RangeInt(1, 100)
}

// CritHitWithLucky 随机是否暴击 带幸运值加成
// baseRate 基础暴击率, lucky 幸运值(0~100), bonus 每点幸运加成
func (r *Randomizer) CritHitWithLucky(baseRate float64, lucky int, bonus float64) bool {
	rate := baseRate + float64(lucky)*bonus
	if rate > 100 {
		rate = 100
	}
	return r.ProbabilityHit(rate)
}

// ShrinkIntSlice 随机丢弃数组中 N 个元素
func (r *Randomizer) ShrinkIntSlice(arr []int, dropCnt int) []int {
	if dropCnt <= 0 || dropCnt >= len(arr) {
		return []int{}
	}
	tmp := make([]int, len(arr))
	copy(tmp, arr)
	r.Shuffle(len(tmp), func(i, j int) { tmp[i], tmp[j] = tmp[j], tmp[i] })
	return tmp[:len(arr)-dropCnt]
}

// RoundFloat 随机生成小数保留指定位数
func (r *Randomizer) RoundFloat(min, max float64, decimals int) float64 {
	f := r.RangeFloat64(min, max)
	pow := math.Pow10(decimals)
	return math.Round(f*pow) / pow
}

// RandomContinuousTimes 随机真假连续次数（模拟NPC行为连击/连续释放技能）
func (r *Randomizer) RandomContinuousTimes(maxTimes int) int {
	return r.RangeInt(1, maxTimes)
}

// SlippagePrice 价格随机滑点
// price 原价, rate 最大滑点比例(0.001=0.1%)
func (r *Randomizer) SlippagePrice(price float64, rate float64) float64 {
	delta := price * rate
	return r.RangeFloat64(price-delta, price+delta)
}

// RandomPositionRate 随机仓位比例 0~1
func (r *Randomizer) RandomPositionRate() float64 {
	return r.RangeFloat64(0, 1.0)
}

// RandomPositionRateRange 指定区间仓位 [min,max] 0~1
func (r *Randomizer) RandomPositionRateRange(min, max float64) float64 {
	return r.RangeFloat64(min, max)
}

// RandomVolatility 随机波动幅度 [min,max] 百分比
func (r *Randomizer) RandomVolatility(minPct, maxPct float64) float64 {
	return r.RangeFloat64(minPct, maxPct)
}

// RandomReturn 生成符合正态分布的收益率
// mean 日均收益均值, std 标准差
func (r *Randomizer) RandomReturn(mean, std float64) float64 {
	return r.Normal(mean, std)
}

// TruncatedNormal 截断正态分布，限制在 [low, high]
func (r *Randomizer) TruncatedNormal(mean, std, low, high float64) float64 {
	for {
		val := r.Normal(mean, std)
		if val >= low && val <= high {
			return val
		}
	}
}

// LogNormal 对数正态分布
func (r *Randomizer) LogNormal(mean, std float64) float64 {
	return math.Exp(r.Normal(mean, std))
}

// ProbWithCooldown 带冷却的概率触发
// p 基础概率, cd 冷却次数, lastTrigger 上次触发后计数
func (r *Randomizer) ProbWithCooldown(p float64, cd int, lastTrigger int) bool {
	if lastTrigger < cd {
		return false
	}
	return r.ProbabilityHit(p)
}

// ShuffleGrid 打乱网格坐标 (x,y)
func (r *Randomizer) ShuffleGrid(width, height int) [][]struct{ X, Y int } {
	var list []struct{ X, Y int }
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			list = append(list, struct{ X, Y int }{x, y})
		}
	}
	r.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })

	res := make([][]struct{ X, Y int }, height)
	for i := 0; i < height; i++ {
		res[i] = list[i*width : (i+1)*width]
	}
	return res
}
