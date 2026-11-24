package sfs2x

import (
	"bytes"
	"compress/zlib"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	inner := New()
	inner.Put("resourceId", TUTFString, "22048")
	inner.Put("sessionToken", TUTFString, "session-abc-1")

	outer := New()
	outer.Put("c", TByte, byte(0))
	outer.Put("a", TShort, int16(0))
	outer.Put("p", TSFSObject, inner)

	encoded := outer.Encode()

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.GetByte("c") != 0 {
		t.Errorf("c = %d, want 0", decoded.GetByte("c"))
	}
	if decoded.GetShort("a") != 0 {
		t.Errorf("a = %d, want 0", decoded.GetShort("a"))
	}
	p := decoded.GetSFSObject("p")
	if p == nil {
		t.Fatal("p is nil")
	}
	if p.GetUTFString("resourceId") != "22048" {
		t.Errorf("resourceId = %q", p.GetUTFString("resourceId"))
	}
}

func TestEncodeTypes(t *testing.T) {
	o := New()
	o.Put("b", TBool, true)
	o.Put("by", TByte, byte(42))
	o.Put("s", TShort, int16(1000))
	o.Put("i", TInt, int32(100000))
	o.Put("l", TLong, int64(9999999999))
	o.Put("d", TDouble, 3.14159265358979)
	o.Put("str", TUTFString, "hello world")
	o.Put("nil", TNull, nil)

	data := o.Encode()
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !decoded.Get("b").(bool) {
		t.Error("b mismatch")
	}
	if decoded.Get("by").(byte) != 42 {
		t.Error("by mismatch")
	}
	if decoded.Get("i").(int32) != 100000 {
		t.Error("i mismatch")
	}
	if decoded.Get("str").(string) != "hello world" {
		t.Error("str mismatch")
	}
}

func TestNestedSFSObject(t *testing.T) {
	inner := New()
	inner.Put("key", TUTFString, "value")

	mid := New()
	mid.Put("inner", TSFSObject, inner)

	outer := New()
	outer.Put("mid", TSFSObject, mid)

	data := outer.Encode()
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	midD := decoded.GetSFSObject("mid")
	if midD == nil {
		t.Fatal("mid nil")
	}
	innerD := midD.GetSFSObject("inner")
	if innerD == nil {
		t.Fatal("inner nil")
	}
	if innerD.GetUTFString("key") != "value" {
		t.Error("key mismatch")
	}
}

func TestFrameRoundtrip(t *testing.T) {
	o := New()
	o.Put("c", TByte, byte(0))
	o.Put("a", TShort, int16(0))
	inner := New()
	inner.Put("tk", TUTFString, "token123")
	inner.Put("ms", TInt, 1000000)
	o.Put("p", TSFSObject, inner)

	frame := EncodeFrame(o)
	decoded, err := DecodeFrame(frame)
	if err != nil {
		t.Fatalf("frame decode: %v", err)
	}
	if decoded.GetByte("c") != 0 {
		t.Error("c mismatch")
	}
	p := decoded.GetSFSObject("p")
	if p == nil {
		t.Fatal("p nil")
	}
	if p.GetUTFString("tk") != "token123" {
		t.Error("tk mismatch")
	}
	if p.GetInt("ms") != 1000000 {
		t.Error("ms mismatch")
	}
}

func TestDecodeFrame_LongLengthForm(t *testing.T) {
	payload := New().Encode()

	// 手工拼装 uint32 长度形态（header bit 0x08）——DecodeFrame 必须兼容
	var buf bytes.Buffer
	buf.WriteByte(Signature | LongLenMask)
	n := len(payload)
	buf.WriteByte(byte(n >> 24))
	buf.WriteByte(byte(n >> 16))
	buf.WriteByte(byte(n >> 8))
	buf.WriteByte(byte(n))
	buf.Write(payload)

	decoded, err := DecodeFrame(buf.Bytes())
	if err != nil {
		t.Fatalf("long-form frame decode: %v", err)
	}
	if decoded == nil {
		t.Fatal("decoded nil")
	}
}

func TestDecodeFrame_Errors(t *testing.T) {
	// 过短
	if _, err := DecodeFrame([]byte{Signature, 0}); err == nil {
		t.Error("short frame must fail")
	}
	// 载荷截断
	bad := []byte{Signature, 0x00, 0x10, byte(TSFSObject), 0x00}
	if _, err := DecodeFrame(bad); err == nil {
		t.Error("truncated payload must fail")
	}
	// 非法 marker
	if _, err := Decode([]byte{0x99, 0x00, 0x00}); err == nil {
		t.Error("invalid marker must fail")
	}
}

func TestDeterministicEncoding(t *testing.T) {
	// 相同内容（不同插入序）必须产出逐字节相同的编码——key 排序保证确定性
	a := New()
	a.Put("zz", TUTFString, "1")
	a.Put("aa", TUTFString, "2")

	b := New()
	b.Put("aa", TUTFString, "2")
	b.Put("zz", TUTFString, "1")

	if !bytes.Equal(a.Encode(), b.Encode()) {
		t.Error("encoding must be deterministic regardless of insertion order")
	}
	if len(a.GetKeys()) != 2 || a.GetKeys()[0] != "zz" {
		t.Errorf("GetKeys must preserve insertion order, got %v", a.GetKeys())
	}
}

// TestDecodeTruncatedNeverPanics walks every truncation of a rich object —
// malformed input must produce an error, never a panic (a library decoding
// untrusted bytes cannot afford slice-out-of-range crashes).
func TestDecodeTruncatedNeverPanics(t *testing.T) {
	inner := New()
	inner.Put("tok", TUTFString, "abcdefgh")
	inner.Put("blob", TByteArray, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	o := New()
	o.Put("b", TBool, true)
	o.Put("s", TShort, int16(-2))
	o.Put("i", TInt, int32(-3))
	o.Put("l", TLong, int64(-4))
	o.Put("f", TFloat, float32(1.5))
	o.Put("d", TDouble, 2.5)
	o.Put("arr", TSFSArray, []interface{}{int32(1), "x", true})
	o.Put("p", TSFSObject, inner)
	full := o.Encode()

	for cut := 0; cut <= len(full); cut++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Decode panicked on truncation %d/%d: %v", cut, len(full), r)
				}
			}()
			_, _ = Decode(full[:cut]) // error expected (or full decode at cut==len)
		}()
	}

	// 单字节翻转也不得 panic
	for i := 0; i < len(full); i++ {
		mutated := append([]byte(nil), full...)
		mutated[i] ^= 0xFF
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Decode panicked on mutation @%d: %v", i, r)
				}
			}()
			_, _ = Decode(mutated)
		}()
	}
}

// TestEncodeFrameAutoBigSize 载荷超过 uint16 上限时必须自动切换 BIG_SIZE
// 长形态（0x08 + uint32）——静默截断会让对端按错误长度读帧。
func TestEncodeFrameAutoBigSize(t *testing.T) {
	o := New()
	o.Put("big", TByteArray, make([]byte, 70000)) // payload > 0xFFFF

	frame := EncodeFrame(o)
	if frame[0]&LongLenMask == 0 {
		t.Fatalf("payload %d must switch to BIG_SIZE form, header %02x", len(frame), frame[0])
	}

	decoded, err := DecodeFrame(frame)
	if err != nil {
		t.Fatalf("big frame decode: %v", err)
	}
	blob, ok := decoded.Get("big").([]byte)
	if !ok || len(blob) != 70000 {
		t.Fatalf("big payload roundtrip failed: got %d bytes", len(blob))
	}
}

// TestDecodeFrame_CompressedTransparent 压缩帧经 DecodeFrame 透明解压——
// 不再是拒绝语义（压缩/加密已实现）。
func TestDecodeFrame_CompressedTransparent(t *testing.T) {
	inner := New()
	inner.Put("blob", TByteArray, bytes.Repeat([]byte{7, 8, 9}, 512))
	o := New()
	o.Put("p", TSFSObject, inner)

	frame, err := EncodeFrameWithOptions(o, FrameOptions{CompressThreshold: 64})
	if err != nil {
		t.Fatalf("encode compressed: %v", err)
	}
	if frame[0]&CompressionMask == 0 {
		t.Fatalf("payload %d over threshold must set 0x20, header %02x", len(frame), frame[0])
	}

	decoded, err := DecodeFrame(frame) // 无选项也应透明解压
	if err != nil {
		t.Fatalf("compressed frame decode: %v", err)
	}
	if blob := decoded.GetSFSObject("p").Get("blob").([]byte); len(blob) != 1536 {
		t.Fatalf("blob len = %d, want 1536", len(blob))
	}

	// 阈值以内不压缩
	small := New()
	small.Put("s", TUTFString, "hi")
	smallFrame, err := EncodeFrameWithOptions(small, FrameOptions{CompressThreshold: 1024})
	if err != nil {
		t.Fatalf("encode small: %v", err)
	}
	if smallFrame[0]&CompressionMask != 0 {
		t.Fatal("small payload must not be compressed")
	}
}

// TestDecodeFrame_RejectsFlags 加密帧无钥匙必须显式报错；非二进制帧也必须拒绝。
func TestDecodeFrame_RejectsFlags(t *testing.T) {
	payload := New().Encode()

	build := func(hdr byte) []byte {
		out := []byte{hdr, byte(len(payload) >> 8), byte(len(payload))}
		return append(out, payload...)
	}

	for name, tc := range map[string]struct {
		hdr   byte
		match string
	}{
		"encrypted":  {Signature | 0x40, "no decryption key"},
		"non-binary": {CompressionMask, "not a binary"},
	} {
		_, err := DecodeFrame(build(tc.hdr))
		if err == nil {
			t.Fatalf("%s frame must be rejected", name)
		}
		if !bytes.Contains([]byte(err.Error()), []byte(tc.match)) {
			t.Fatalf("%s frame error %q must mention %q", name, err.Error(), tc.match)
		}
	}
}

// TestTypedArraysRoundtrip 七个标量数组类型全量回环——补上此前解码器缺失的
// 7 个类型（bool/short/int/long/float/double/utf_string array）。
func TestTypedArraysRoundtrip(t *testing.T) {
	o := New()
	o.Put("ba", TBoolArray, []bool{true, false, true})
	o.Put("sa", TShortArray, []int16{-2, -1, 0, 1, 2})
	o.Put("ia", TIntArray, []int32{-100000, 0, 100000})
	o.Put("la", TLongArray, []int64{-9999999999, 0, 9999999999})
	o.Put("fa", TFloatArray, []float32{1.5, -2.25})
	o.Put("da", TDoubleArray, []float64{3.14159265358979, -0.5})
	o.Put("ua", TUTFStringArray, []string{"alpha", "", "omega"})

	decoded, err := Decode(o.Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	ba, _ := decoded.Get("ba").([]bool)
	assert.Equal(t, []bool{true, false, true}, ba)
	sa, _ := decoded.Get("sa").([]int16)
	assert.Equal(t, []int16{-2, -1, 0, 1, 2}, sa)
	ia, _ := decoded.Get("ia").([]int32)
	assert.Equal(t, []int32{-100000, 0, 100000}, ia)
	la, _ := decoded.Get("la").([]int64)
	assert.Equal(t, []int64{-9999999999, 0, 9999999999}, la)
	fa, _ := decoded.Get("fa").([]float32)
	require.Len(t, fa, 2)
	assert.InDelta(t, 1.5, fa[0], 1e-6)
	assert.InDelta(t, -2.25, fa[1], 1e-6)
	da, _ := decoded.Get("da").([]float64)
	assert.Equal(t, []float64{3.14159265358979, -0.5}, da)
	ua, _ := decoded.Get("ua").([]string)
	assert.Equal(t, []string{"alpha", "", "omega"}, ua)

	// 类型码钉住：解码后条目类型必须与写入一致
	for key, want := range map[string]Type{
		"ba": TBoolArray, "sa": TShortArray, "ia": TIntArray, "la": TLongArray,
		"fa": TFloatArray, "da": TDoubleArray, "ua": TUTFStringArray,
	} {
		if got := decoded.GetType(key); got != want {
			t.Fatalf("GetType(%q) = %d, want %d", key, got, want)
		}
	}
}

// TestNegativeNumbersRoundtrip 负数两补数编码必须无损回环（此前只测了不 panic）。
func TestNegativeNumbersRoundtrip(t *testing.T) {
	o := New()
	o.Put("s", TShort, int16(-1234))
	o.Put("i", TInt, int32(-123456))
	o.Put("l", TLong, int64(-1234567890123))

	decoded, err := Decode(o.Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	assert.Equal(t, int16(-1234), decoded.GetShort("s"))
	assert.Equal(t, int32(-123456), decoded.GetInt("i"))
	assert.Equal(t, int64(-1234567890123), decoded.GetLong("l"))
}

// TestEmptyObjectAndStrings 空对象/空串/空数组边界。
func TestEmptyObjectAndStrings(t *testing.T) {
	empty := New()
	decoded, err := Decode(empty.Encode())
	if err != nil {
		t.Fatalf("empty object decode: %v", err)
	}
	if len(decoded.GetKeys()) != 0 {
		t.Fatal("empty object must stay empty")
	}

	o := New()
	o.Put("e", TUTFString, "")
	o.Put("ea", TUTFStringArray, []string{})
	o.Put("ia", TIntArray, []int32{})
	decoded, err = Decode(o.Encode())
	if err != nil {
		t.Fatalf("empty values decode: %v", err)
	}
	if decoded.GetUTFString("e") != "" {
		t.Fatal("empty string mismatch")
	}
	if got, _ := decoded.Get("ea").([]string); len(got) != 0 {
		t.Fatal("empty string array mismatch")
	}
	if got, _ := decoded.Get("ia").([]int32); len(got) != 0 {
		t.Fatal("empty int array mismatch")
	}
}

// TestPutOverwriteKeepsSingleKey 同 key 重复 Put 必须覆盖且 keys 去重。
func TestPutOverwriteKeepsSingleKey(t *testing.T) {
	o := New()
	o.Put("k", TUTFString, "first")
	o.Put("k", TInt, int32(7))
	o.Put("other", TBool, true)

	if len(o.GetKeys()) != 2 {
		t.Fatalf("keys after overwrite = %v, want [k other]", o.GetKeys())
	}
	decoded, err := Decode(o.Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.GetInt("k") != 7 {
		t.Fatal("overwrite must win")
	}
}

// TestFloat32Precision float32 回环位精度。
func TestFloat32Precision(t *testing.T) {
	o := New()
	o.Put("f", TFloat, float32(0.1))
	decoded, err := Decode(o.Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	got, _ := decoded.Get("f").(float32)
	if float64(got) != float64(float32(0.1)) {
		t.Fatalf("float32 roundtrip = %v, want %v", got, float32(0.1))
	}
}

// TestEncryptDecryptFrameRoundtrip 加密帧全链路：EncodeFrameWithOptions →
// DecodeFrameWithOptions。压缩+加密叠加时顺序必须与官方一致（先压后加密）。
func TestEncryptDecryptFrameRoundtrip(t *testing.T) {
	key := []byte("0123456789abcdef") // 16B

	o := New()
	o.Put("str", TUTFString, "payload for encryption test")
	o.Put("n", TInt, int32(-42))

	// 纯加密
	frame, err := EncodeFrameWithOptions(o, FrameOptions{EncryptionKey: key})
	if err != nil {
		t.Fatalf("encode encrypted: %v", err)
	}
	if frame[0]&0x40 == 0 {
		t.Fatal("encrypted frame must set 0x40")
	}

	// 随机 IV：同内容两次编码产出不同字节
	frame2, err := EncodeFrameWithOptions(o, FrameOptions{EncryptionKey: key})
	if err != nil {
		t.Fatalf("encode encrypted again: %v", err)
	}
	if bytes.Equal(frame, frame2) {
		t.Fatal("random IV must make encrypted frames non-deterministic")
	}

	decoded, err := DecodeFrameWithOptions(frame, FrameOptions{EncryptionKey: key})
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decoded.GetUTFString("str") != "payload for encryption test" || decoded.GetInt("n") != -42 {
		t.Fatal("encrypted roundtrip mismatch")
	}

	// 无钥匙的 DecodeFrame 拒绝
	if _, err := DecodeFrame(frame); err == nil {
		t.Fatal("DecodeFrame without key must reject encrypted frame")
	}

	// 错钥匙 → 报错不 panic
	wrong := []byte("fedcba9876543210")
	if _, err := DecodeFrameWithOptions(frame, FrameOptions{EncryptionKey: wrong}); err == nil {
		t.Fatal("wrong key must fail")
	}

	// 压缩+加密叠加（先压后加密）
	big := New()
	big.Put("blob", TByteArray, bytes.Repeat([]byte{1, 2, 3}, 512))
	joint, err := EncodeFrameWithOptions(big, FrameOptions{CompressThreshold: 64, EncryptionKey: key})
	if err != nil {
		t.Fatalf("encode joint: %v", err)
	}
	if joint[0]&CompressionMask == 0 || joint[0]&0x40 == 0 {
		t.Fatalf("joint frame must set both flags, header %02x", joint[0])
	}
	decodedJoint, err := DecodeFrameWithOptions(joint, FrameOptions{EncryptionKey: key})
	if err != nil {
		t.Fatalf("decode joint: %v", err)
	}
	if blob := decodedJoint.Get("blob").([]byte); len(blob) != 1536 {
		t.Fatalf("joint roundtrip blob len = %d", len(blob))
	}
}

// TestEncryptKeyValidation 钥匙长度校验（必须 16B）。
func TestEncryptKeyValidation(t *testing.T) {
	// nil = 不加密；非 nil 但长度不对必须在编码期拒绝
	for _, bad := range [][]byte{{}, []byte("short"), make([]byte, 15), make([]byte, 17), make([]byte, 32)} {
		if _, err := EncodeFrameWithOptions(New(), FrameOptions{EncryptionKey: bad}); err == nil {
			t.Fatalf("key len %d must be rejected on encode", len(bad))
		}
	}

	key := make([]byte, 16)
	frame, err := EncodeFrameWithOptions(New(), FrameOptions{EncryptionKey: key})
	if err != nil {
		t.Fatalf("valid key encode: %v", err)
	}
	for _, bad := range [][]byte{nil, []byte("short")} {
		if _, err := DecodeFrameWithOptions(frame, FrameOptions{EncryptionKey: bad}); err == nil {
			t.Fatalf("key len %d must be rejected on decode", len(bad))
		}
	}
}

// TestCompressionBombCapped 恶意压缩帧的解压输出必须有上限（防解压炸弹）。
func TestCompressionBombCapped(t *testing.T) {
	// 手工构造：超出上限的解压输出。zlib 头 + 压缩的零流
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	zw.Write(make([]byte, 100<<20)) // 100MB 零 → 压得很小
	zw.Close()

	payload := buf.Bytes()
	hdr := []byte{Signature | CompressionMask, byte(len(payload) >> 24), byte(len(payload) >> 16), byte(len(payload) >> 8), byte(len(payload))}
	frame := append(hdr, payload...)

	_, err := DecodeFrame(frame)
	if err == nil {
		t.Fatal("decompression bomb must be capped with an error")
	}
}
