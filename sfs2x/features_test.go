package sfs2x

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// 官方固定 key+IV 加密模式
// ============================================================

func TestFixedIVEncryption_OfficialMode(t *testing.T) {
	ck, err := ParseCryptoKeyResponseBase64(
		base64Of(bytes.Repeat([]byte{0xAB}, 16), bytes.Repeat([]byte{0x11}, 16)))
	require.NoError(t, err)

	o := New()
	o.Put("cmd", TUTFString, "secure op")
	o.Put("n", TInt, int32(-7))

	// 官方模式：固定 key+IV，载荷只有密文（无 IV 前缀）
	frame, err := EncodeFrameWithOptions(o, FrameOptions{EncryptionKey: ck.Key, EncryptionIV: ck.IV})
	require.NoError(t, err)
	if frame[0]&0x40 == 0 {
		t.Fatal("must set 0x40")
	}
	// 自描述模式同一对象：密文长度 = 官方密文 + 16B IV 前缀（IV 在载荷里）
	selfFrame, err := EncodeFrameWithOptions(o, FrameOptions{EncryptionKey: ck.Key})
	require.NoError(t, err)
	// 帧长度差 = IV 前缀（两模式 padding 相同、长度字段宽度相同）
	if len(selfFrame)-len(frame) != ivSize {
		t.Fatalf("self-describing mode should carry +%d IV prefix: self=%d official=%d",
			ivSize, len(selfFrame), len(frame))
	}

	// 官方模式回环
	decoded, err := DecodeFrameWithOptions(frame, FrameOptions{EncryptionKey: ck.Key, EncryptionIV: ck.IV})
	require.NoError(t, err)
	assert.Equal(t, "secure op", decoded.GetUTFString("cmd"))
	assert.Equal(t, int32(-7), decoded.GetInt("n"))

	// 固定 IV 模式确定性（同输入同输出——IV 不随机）
	frame2, err := EncodeFrameWithOptions(o, FrameOptions{EncryptionKey: ck.Key, EncryptionIV: ck.IV})
	require.NoError(t, err)
	assert.True(t, bytesEqual(frame, frame2), "fixed-IV mode is deterministic")

	// 官方模式确定性帧错误钥匙 → 报错
	wrong, err := ParseCryptoKeyResponseBase64(
		base64Of(bytes.Repeat([]byte{0x01}, 16), bytes.Repeat([]byte{0x02}, 16)))
	require.NoError(t, err)
	_, err = DecodeFrameWithOptions(frame, FrameOptions{EncryptionKey: wrong.Key, EncryptionIV: wrong.IV})
	assert.Error(t, err, "wrong fixed key/iv must fail padding check")
}

func TestParseCryptoKeyResponse(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)
	iv := bytes.Repeat([]byte{0x22}, 16)

	ck, err := ParseCryptoKeyResponse(append(append([]byte{}, key...), iv...))
	require.NoError(t, err)
	assert.Equal(t, key, ck.Key)
	assert.Equal(t, iv, ck.IV)

	ck, err = ParseCryptoKeyResponseBase64(base64Of(key, iv))
	require.NoError(t, err)
	assert.Equal(t, key, ck.Key)
	assert.Equal(t, iv, ck.IV)

	_, err = ParseCryptoKeyResponse([]byte("short"))
	assert.Error(t, err, "wrong length must fail")
}

// ============================================================
// EncodeOrdered 插入序字节忠实模式
// ============================================================

func TestEncodeOrdered_ByteFaithfulToInsertionOrder(t *testing.T) {
	// 插入序 z 在前：Encode（排序）与 EncodeOrdered（插入序）必须产出不同字节，
	// 且各自解码回值一致
	o := New()
	o.Put("zz", TUTFString, "first-inserted")
	o.Put("aa", TUTFString, "second-inserted")

	sorted := o.Encode()
	ordered := o.EncodeOrdered()
	assert.False(t, bytesEqual(sorted, ordered), "two orders must differ")

	// 排序版以 "aa" 开头（sorted 编码首个 key 是 aa）
	assert.Contains(t, string(sorted[:16]), "aa")
	// 插入序版以 "zz" 开头
	assert.Contains(t, string(ordered[:16]), "zz")

	for _, enc := range [][]byte{sorted, ordered} {
		decoded, err := Decode(enc)
		require.NoError(t, err)
		assert.Equal(t, "first-inserted", decoded.GetUTFString("zz"))
		assert.Equal(t, "second-inserted", decoded.GetUTFString("aa"))
	}

	// FrameOptions.Ordered 走同一插入序路径
	framePlain := EncodeFrame(o)
	frameOrdered, err := EncodeFrameWithOptions(o, FrameOptions{Ordered: true})
	require.NoError(t, err)
	assert.True(t, bytesEqual(EncodeFrameOf(sorted), framePlain))
	assert.True(t, bytesEqual(EncodeFrameOf(ordered), frameOrdered))
}

// ============================================================
// AsClass（$C/$F 官方 CLASS 约定）
// ============================================================

func TestAsClass(t *testing.T) {
	inner := New()
	inner.Put("$C", TUTFString, "com.example.dto.Player")
	fields := []interface{}{int32(7), "name-x"}
	inner.Put("$F", TSFSArray, fields)

	className, gotFields, ok := AsClass(inner)
	require.True(t, ok)
	assert.Equal(t, "com.example.dto.Player", className)
	assert.Equal(t, fields, gotFields)

	// 回环：含 $C/$F 的对象经编解码后仍被识别
	decoded, err := Decode(inner.Encode())
	require.NoError(t, err)
	_, _, ok = AsClass(decoded)
	assert.True(t, ok)

	// 普通对象不误判
	plain := New()
	plain.Put("x", TInt, int32(1))
	_, _, ok = AsClass(plain)
	assert.False(t, ok)
}

// ============================================================
// Message 信封
// ============================================================

func TestMessageRoundtrip(t *testing.T) {
	params := New()
	params.Put("zone", TUTFString, "lobby")
	params.Put("reconnect", TBool, false)

	msg := NewMessage(ControllerSystem, SysActionLogin, params)
	frame, err := msg.EncodeOrderedFrame()
	require.NoError(t, err)

	parsed, err := ParseMessage(frame)
	require.NoError(t, err)
	assert.Equal(t, ControllerSystem, parsed.Controller)
	assert.Equal(t, SysActionLogin, parsed.Action)
	assert.Equal(t, "lobby", parsed.Params.GetUTFString("zone"))
}

func TestMessage_EnvelopeErrors(t *testing.T) {
	o := New()
	o.Put("c", TUTFString, "wrong-type")
	_, err := ParseMessageObject(o)
	assert.Error(t, err, "c 非 BYTE 必须拒绝")
}

// ============================================================
// 握手报文
// ============================================================

func TestHandshakeRequestShape(t *testing.T) {
	msg := NewHandshakeRequest("1.7.3", "GoClient/1.0")
	assert.Equal(t, ControllerSystem, msg.Controller)
	assert.Equal(t, SysActionHandshake, msg.Action)
	assert.Equal(t, "1.7.3", msg.Params.GetUTFString(HandshakeKeyAPI))
	assert.Equal(t, "GoClient/1.0", msg.Params.GetUTFString(HandshakeKeyClientType))
}

func TestParseHandshakeResponse(t *testing.T) {
	resp := New()
	resp.Put(HandshakeKeySessionToken, TUTFString, "sess-token-1")
	resp.Put(HandshakeKeyCompression, TInt, int32(1024))
	resp.Put(HandshakeKeyMaxMessage, TInt, int32(1000000))

	parsed, err := ParseHandshakeResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, "sess-token-1", parsed.SessionToken)
	assert.Equal(t, 1024, parsed.CompressionThreshold)
	assert.Equal(t, 1000000, parsed.MaxMessageSize)

	// 缺 tk 报错
	_, err = ParseHandshakeResponse(New())
	assert.Error(t, err)
}

// ============================================================
// Benchmark
// ============================================================

func benchObject() *Object {
	o := New()
	o.Put("id", TLong, int64(1234567890123))
	o.Put("name", TUTFString, "bench-object")
	o.Put("score", TDouble, 98.76)
	o.Put("flags", TBoolArray, []bool{true, false, true, true})
	inner := New()
	inner.Put("x", TFloat, float32(1.5))
	inner.Put("tag", TUTFString, "inner-tag")
	o.Put("meta", TSFSObject, inner)
	return o
}

func BenchmarkEncode(b *testing.B) {
	o := benchObject()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = o.Encode()
	}
}

func BenchmarkDecode(b *testing.B) {
	data := benchObject().Encode()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode(data)
	}
}

func BenchmarkEncodeFrame(b *testing.B) {
	o := benchObject()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = EncodeFrame(o)
	}
}

func BenchmarkEncodeFrameCompressed(b *testing.B) {
	o := benchObject()
	blob := bytes.Repeat([]byte{1, 2, 3, 4}, 256)
	o.Put("blob", TByteArray, blob)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := EncodeFrameWithOptions(o, FrameOptions{CompressThreshold: 256}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeFrameEncrypted(b *testing.B) {
	o := benchObject()
	key := bytes.Repeat([]byte{9}, 16)
	iv := bytes.Repeat([]byte{7}, 16)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := EncodeFrameWithOptions(o, FrameOptions{EncryptionKey: key, EncryptionIV: iv}); err != nil {
			b.Fatal(err)
		}
	}
}

// ---- 小工具 ----

func base64Of(key, iv []byte) string {
	raw := append(append([]byte{}, key...), iv...)
	return base64.StdEncoding.EncodeToString(raw)
}

func EncodeFrameOf(objectBytes []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(Signature)
	n := len(objectBytes)
	buf.WriteByte(byte(n >> 8))
	buf.WriteByte(byte(n))
	buf.Write(objectBytes)
	return buf.Bytes()
}
