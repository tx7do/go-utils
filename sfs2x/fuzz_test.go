package sfs2x

// Fuzz targets: malformed input must produce errors, never panics. Run a
// local session with:
//
//	go test -fuzz FuzzDecode -fuzztime 30s
//	go test -fuzz FuzzDecodeFrame -fuzztime 30s

import (
	"testing"
)

func FuzzDecode(f *testing.F) {
	// 种子：全类型富对象 + 空对象 + 混合数组
	o := New()
	o.Put("b", TBool, true)
	o.Put("s", TShort, int16(-2))
	o.Put("i", TInt, int32(-3))
	o.Put("l", TLong, int64(-4))
	o.Put("f", TFloat, float32(1.5))
	o.Put("d", TDouble, 2.5)
	o.Put("str", TUTFString, "seed")
	o.Put("text", TText, "long seed")
	o.Put("blob", TByteArray, []byte{1, 2, 3})
	o.Put("ba", TBoolArray, []bool{true, false})
	o.Put("ia", TIntArray, []int32{7, -8})
	o.Put("ua", TUTFStringArray, []string{"x", "y"})
	o.Put("arr", TSFSArray, []interface{}{int32(1), "x", true})
	inner := New()
	inner.Put("deep", TUTFString, "v")
	o.Put("obj", TSFSObject, inner)
	f.Add(o.Encode())
	f.Add([]byte{0x12, 0x00, 0x00})
	f.Add([]byte{0x99})
	f.Add([]byte(nil))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = Decode(data) // 只断言不 panic；错误/成功皆可
	})
}

func FuzzDecodeFrame(f *testing.F) {
	o := New()
	o.Put("x", TInt, int32(1))
	f.Add(EncodeFrame(o))
	f.Add([]byte{Signature, 0x00, 0x00})
	f.Add([]byte{Signature | LongLenMask, 0, 0, 0, 0})
	f.Add([]byte(nil))

	key := []byte("0123456789abcdef")
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeFrame(data)
		_, _ = DecodeFrameWithOptions(data, FrameOptions{EncryptionKey: key})
	})
}
