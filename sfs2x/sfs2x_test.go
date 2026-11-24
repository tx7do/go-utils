package sfs2x

import (
	"bytes"
	"testing"
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

// TestDecodeFrame_RejectsFlags 压缩/加密/非二进制帧必须显式报错，
// 而不是把未知载荷硬解成误导性的 marker 错误。
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
		"compressed": {Signature | CompressionMask, "compressed"},
		"encrypted":  {Signature | 0x40, "encrypted"},
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
