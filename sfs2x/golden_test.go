package sfs2x

// Golden-vector interop tests: every .bin file in testdata/ was produced by
// the ZewSFS Python reference implementation (testdata/golden_gen.py builds
// them from a standalone copy of its core codec). The test asserts, per case:
//
//  1. our Decode parses the ZewSFS bytes and every key's type code and value
//     match the manifest expectations (wire-level interop: their encoder ↔
//     our decoder);
//  2. re-encoding the decoded object produces byte-identical output
//     (our encoder ↔ their bytes: golden cases insert keys pre-sorted, so
//     the deterministic ordering aligns with ZewSFS insertion order).
//
// Regenerate with: python testdata/golden_gen.py <zewcore-parent-dir>

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type goldenEntry struct {
	Key   string          `json:"key"`
	Type  int             `json:"type"`
	Value json.RawMessage `json:"value"`
}

type goldenCase struct {
	Name    string        `json:"name"`
	Entries []goldenEntry `json:"entries"`
}

type goldenManifest struct {
	Cases []goldenCase `json:"cases"`
}

func TestGoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "golden.json"))
	if err != nil {
		t.Fatalf("read golden.json: %v (run testdata/golden_gen.py first)", err)
	}
	var manifest goldenManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse golden.json: %v", err)
	}
	if len(manifest.Cases) == 0 {
		t.Fatal("empty golden manifest")
	}

	for _, tc := range manifest.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			bin, err := os.ReadFile(filepath.Join("testdata", tc.Name+".bin"))
			if err != nil {
				t.Fatalf("read %s.bin: %v", tc.Name, err)
			}

			decoded, err := Decode(bin)
			if err != nil {
				t.Fatalf("Decode(%s.bin): %v", tc.Name, err)
			}

			// 重建对象（按 manifest 类型码），供字节级比对
			rebuilt := New()
			for i := range tc.Entries {
				e := &tc.Entries[i]
				requireKeyAbsent(t, rebuilt, e.Key)
				rebuilt.Put(e.Key, Type(e.Type), buildGoValue(t, e.Type, e.Value))
			}

			for i := range tc.Entries {
				e := &tc.Entries[i]
				if got := decoded.GetType(e.Key); int(got) != e.Type {
					t.Fatalf("GetType(%q) = %d, want %d", e.Key, got, e.Type)
				}
				assertValue(t, e.Type, e.Value, decoded.Get(e.Key), fmt.Sprintf("key %q", e.Key))
			}

			// 字节级往返：重编码必须与 ZewSFS 原字节一致
			if !bytesEqual(rebuilt.Encode(), bin) {
				t.Fatalf("re-encode mismatch: got %d bytes, want %d (byte-faithful encode required)",
					len(rebuilt.Encode()), len(bin))
			}
		})
	}
}

// ---- helpers ----

func requireKeyAbsent(t *testing.T, o *Object, key string) {
	t.Helper()
	for _, k := range o.GetKeys() {
		if k == key {
			t.Fatalf("duplicate key %q in manifest (sorted-insertion contract broken)", key)
		}
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// buildGoValue converts a manifest JSON value into the Go representation
// that Put expects for the given type code.
func buildGoValue(t *testing.T, typeCode int, raw json.RawMessage) interface{} {
	t.Helper()
	switch Type(typeCode) {
	case TNull:
		return nil
	case TBool:
		var v bool
		mustJSON(t, raw, &v)
		return v
	case TByte:
		var v int64
		mustJSON(t, raw, &v)
		return byte(v)
	case TShort:
		var v int64
		mustJSON(t, raw, &v)
		return int16(v)
	case TInt:
		var v int64
		mustJSON(t, raw, &v)
		return int32(v)
	case TLong:
		var v int64
		mustJSON(t, raw, &v)
		return v
	case TFloat:
		var v float64
		mustJSON(t, raw, &v)
		return float32(v)
	case TDouble:
		var v float64
		mustJSON(t, raw, &v)
		return v
	case TUTFString, TText:
		var v string
		mustJSON(t, raw, &v)
		return v
	case TByteArray:
		var b64 string
		mustJSON(t, raw, &b64)
		v, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatalf("base64 ByteArray: %v", err)
		}
		return v
	case TBoolArray:
		var list []bool
		mustJSON(t, raw, &list)
		return list
	case TShortArray:
		var list []int64
		mustJSON(t, raw, &list)
		out := make([]int16, len(list))
		for i, v := range list {
			out[i] = int16(v)
		}
		return out
	case TIntArray:
		var list []int64
		mustJSON(t, raw, &list)
		out := make([]int32, len(list))
		for i, v := range list {
			out[i] = int32(v)
		}
		return out
	case TLongArray:
		var list []int64
		mustJSON(t, raw, &list)
		return list
	case TFloatArray:
		var list []float64
		mustJSON(t, raw, &list)
		out := make([]float32, len(list))
		for i, v := range list {
			out[i] = float32(v)
		}
		return out
	case TDoubleArray:
		var list []float64
		mustJSON(t, raw, &list)
		return list
	case TUTFStringArray:
		var list []string
		mustJSON(t, raw, &list)
		return list
	case TSFSArray:
		var elems []struct {
			Type  int             `json:"type"`
			Value json.RawMessage `json:"value"`
		}
		mustJSON(t, raw, &elems)
		out := make([]interface{}, len(elems))
		for i, e := range elems {
			out[i] = buildGoValue(t, e.Type, e.Value)
		}
		return out
	case TSFSObject:
		var wrapper struct {
			Entries []goldenEntry `json:"entries"`
		}
		mustJSON(t, raw, &wrapper)
		entries := wrapper.Entries
		sub := New()
		for i := range entries {
			e := &entries[i]
			sub.Put(e.Key, Type(e.Type), buildGoValue(t, e.Type, e.Value))
		}
		return sub
	default:
		t.Fatalf("unsupported golden type code %d", typeCode)
		return nil
	}
}

// assertValue compares a decoded value against the manifest expectation.
func assertValue(t *testing.T, typeCode int, raw json.RawMessage, got interface{}, label string) {
	t.Helper()
	switch Type(typeCode) {
	case TNull:
		if got != nil {
			t.Fatalf("%s: expected null, got %v", label, got)
		}
	case TBool:
		want := raw[0] == 't'
		if g, _ := got.(bool); g != want {
			t.Fatalf("%s: bool = %v, want %v", label, g, want)
		}
	case TByte:
		assertInt(t, label, got, raw, 8)
	case TShort:
		assertInt(t, label, got, raw, 16)
	case TInt:
		assertInt(t, label, got, raw, 32)
	case TLong:
		assertInt(t, label, got, raw, 64)
	case TFloat:
		var want float64
		mustJSON(t, raw, &want)
		g, _ := got.(float32)
		if math.Abs(float64(g)-want) > 1e-6*float64(1+math.Abs(want)) {
			t.Fatalf("%s: float32 = %v, want %v", label, g, want)
		}
	case TDouble:
		var want float64
		mustJSON(t, raw, &want)
		g, _ := got.(float64)
		if g != want {
			t.Fatalf("%s: float64 = %v, want %v", label, g, want)
		}
	case TUTFString, TText:
		var want string
		mustJSON(t, raw, &want)
		g, _ := got.(string)
		if g != want {
			t.Fatalf("%s: string = %q, want %q", label, g, want)
		}
	case TByteArray:
		var b64 string
		mustJSON(t, raw, &b64)
		want, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatalf("%s: base64: %v", label, err)
		}
		g, _ := got.([]byte)
		if !bytesEqual(g, want) {
			t.Fatalf("%s: ByteArray = %d bytes, want %d", label, len(g), len(want))
		}
	case TBoolArray, TShortArray, TIntArray, TLongArray, TFloatArray, TDoubleArray, TUTFStringArray:
		assertTypedArray(t, Type(typeCode), raw, got, label)
	case TSFSArray:
		var elems []struct {
			Type  int             `json:"type"`
			Value json.RawMessage `json:"value"`
		}
		mustJSON(t, raw, &elems)
		list, _ := got.([]interface{})
		if len(list) != len(elems) {
			t.Fatalf("%s: array len = %d, want %d", label, len(list), len(elems))
		}
		for i, e := range elems {
			assertValue(t, e.Type, e.Value, list[i], fmt.Sprintf("%s[%d]", label, i))
		}
	case TSFSObject:
		var wrapper struct {
			Entries []goldenEntry `json:"entries"`
		}
		mustJSON(t, raw, &wrapper)
		entries := wrapper.Entries
		sub, _ := got.(*Object)
		if sub == nil {
			t.Fatalf("%s: expected nested object, got nil", label)
		}
		for i := range entries {
			e := &entries[i]
			if _, ok := sub.entries[e.Key]; !ok {
				t.Fatalf("%s: missing key %q", label, e.Key)
			}
			assertValue(t, e.Type, e.Value, sub.Get(e.Key), fmt.Sprintf("%s.%s", label, e.Key))
		}
	}
}

func assertInt(t *testing.T, label string, got interface{}, raw json.RawMessage, bits int) {
	t.Helper()
	var want int64
	mustJSON(t, raw, &want)
	var g int64
	switch bits {
	case 8:
		v, ok := got.(byte)
		if !ok {
			t.Fatalf("%s: want byte, got %T", label, got)
		}
		g = int64(v)
	case 16:
		v, ok := got.(int16)
		if !ok {
			t.Fatalf("%s: want int16, got %T", label, got)
		}
		g = int64(v)
	case 32:
		v, ok := got.(int32)
		if !ok {
			t.Fatalf("%s: want int32, got %T", label, got)
		}
		g = int64(v)
	case 64:
		v, ok := got.(int64)
		if !ok {
			t.Fatalf("%s: want int64, got %T", label, got)
		}
		g = v
	}
	if g != want {
		t.Fatalf("%s: int = %d, want %d", label, g, want)
	}
}

func assertTypedArray(t *testing.T, typ Type, raw json.RawMessage, got interface{}, label string) {
	t.Helper()
	switch typ {
	case TBoolArray:
		var want []bool
		mustJSON(t, raw, &want)
		g, _ := got.([]bool)
		if !boolSlicesEqual(g, want) {
			t.Fatalf("%s: BoolArray = %v, want %v", label, g, want)
		}
	case TShortArray:
		var want []int64
		mustJSON(t, raw, &want)
		g, _ := got.([]int16)
		if len(g) != len(want) {
			t.Fatalf("%s: ShortArray len = %d, want %d", label, len(g), len(want))
		}
		for i := range want {
			if int64(g[i]) != want[i] {
				t.Fatalf("%s: ShortArray[%d] = %d, want %d", label, i, g[i], want[i])
			}
		}
	case TIntArray:
		var want []int64
		mustJSON(t, raw, &want)
		g, _ := got.([]int32)
		if len(g) != len(want) {
			t.Fatalf("%s: IntArray len = %d, want %d", label, len(g), len(want))
		}
		for i := range want {
			if int64(g[i]) != want[i] {
				t.Fatalf("%s: IntArray[%d] = %d, want %d", label, i, g[i], want[i])
			}
		}
	case TLongArray:
		var want []int64
		mustJSON(t, raw, &want)
		g, _ := got.([]int64)
		for i := range want {
			if i < len(g) && g[i] != want[i] {
				t.Fatalf("%s: LongArray[%d] = %d, want %d", label, i, g[i], want[i])
			}
		}
		if len(g) != len(want) {
			t.Fatalf("%s: LongArray len = %d, want %d", label, len(g), len(want))
		}
	case TFloatArray:
		var want []float64
		mustJSON(t, raw, &want)
		g, _ := got.([]float32)
		if len(g) != len(want) {
			t.Fatalf("%s: FloatArray len = %d, want %d", label, len(g), len(want))
		}
		for i := range want {
			if math.Abs(float64(g[i])-want[i]) > 1e-6*float64(1+math.Abs(want[i])) {
				t.Fatalf("%s: FloatArray[%d] = %v, want %v", label, i, g[i], want[i])
			}
		}
	case TDoubleArray:
		var want []float64
		mustJSON(t, raw, &want)
		g, _ := got.([]float64)
		for i := range want {
			if i < len(g) && g[i] != want[i] {
				t.Fatalf("%s: DoubleArray[%d] = %v, want %v", label, i, g[i], want[i])
			}
		}
		if len(g) != len(want) {
			t.Fatalf("%s: DoubleArray len = %d, want %d", label, len(g), len(want))
		}
	case TUTFStringArray:
		var want []string
		mustJSON(t, raw, &want)
		g, _ := got.([]string)
		if len(g) != len(want) {
			t.Fatalf("%s: UtfStringArray len = %d, want %d", label, len(g), len(want))
		}
		for i := range want {
			if g[i] != want[i] {
				t.Fatalf("%s: UtfStringArray[%d] = %q, want %q", label, i, g[i], want[i])
			}
		}
	}
}

func boolSlicesEqual(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mustJSON(t *testing.T, raw json.RawMessage, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("manifest value %s: %v", string(raw), err)
	}
}
