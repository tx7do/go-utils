// Package sfs2x implements the SFS2X binary protocol — the object
// serialization and message framing used by SmartFoxServer 2X compatible
// services.
//
// Wire format:
//
//	frame = [header 1B][len][payload]
//	  header: bit5 (0x20) = zlib-compressed payload (threshold 1024),
//	          bit3 (0x08) = length is uint32, otherwise uint16
//	object = [0x12][u16 count][entries...]
//	  entry = [u16 keyLen][key utf8][1B type][value]
//	nested SFS_OBJECT / SFS_ARRAY values carry no separate marker —
//	their type byte IS the marker (the decoder seeks back one byte)
//
// Type codes:
//
//	NULL=0 BOOL=1 BYTE=2 SHORT=3 INT=4 LONG=5 FLOAT=6 DOUBLE=7 UTF_STRING=8
//	BOOL_ARRAY=9 BYTE_ARRAY=10 SHORT_ARRAY=11 INT_ARRAY=12 LONG_ARRAY=13
//	FLOAT_ARRAY=14 DOUBLE_ARRAY=15 UTF_STRING_ARRAY=16 SFS_ARRAY=17 SFS_OBJECT=18
//
// Scope note: EncodeFrame always emits the short (uint16) length form.
// DecodeFrame accepts both length forms; zlib-compressed frames (header bit
// 0x20 set) are not decompressed by this package — callers must handle that
// variant themselves if peers emit it.
package sfs2x

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
)

// Type is an SFS data type code.
type Type byte

// SFS data type codes and frame header flags.
const (
	TNull           Type = 0
	TBool           Type = 1
	TByte           Type = 2
	TShort          Type = 3
	TInt            Type = 4
	TLong           Type = 5
	TFloat          Type = 6
	TDouble         Type = 7
	TUTFString      Type = 8
	TBoolArray      Type = 9
	TByteArray      Type = 10
	TShortArray     Type = 11
	TIntArray       Type = 12
	TLongArray      Type = 13
	TFloatArray     Type = 14
	TDoubleArray    Type = 15
	TUTFStringArray Type = 16
	TSFSArray       Type = 17
	TSFSObject      Type = 18
	CompressionMask      = 0x20
	LongLenMask          = 0x08
	Signature            = 0x80
)

// Object is an ordered key → typed-value mapping.
type Object struct {
	entries map[string]entry
	keys    []string // insertion order
}

type entry struct {
	typ Type
	val interface{}
}

// New creates an empty Object.
func New() *Object {
	return &Object{entries: make(map[string]entry)}
}

// Put writes a typed value.
func (o *Object) Put(key string, typ Type, val interface{}) {
	if _, ok := o.entries[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.entries[key] = entry{typ: typ, val: val}
}

// Get reads the raw value.
func (o *Object) Get(key string) interface{} {
	e, ok := o.entries[key]
	if !ok {
		return nil
	}
	return e.val
}

// GetType reads the entry's type code.
func (o *Object) GetType(key string) Type {
	return o.entries[key].typ
}

// GetKeys returns the keys in insertion order.
func (o *Object) GetKeys() []string {
	return o.keys
}

// GetByte reads a BYTE value.
func (o *Object) GetByte(key string) byte { return o.entries[key].val.(byte) }

// GetShort reads a SHORT value.
func (o *Object) GetShort(key string) int16 { return o.entries[key].val.(int16) }

// GetInt reads an INT value.
func (o *Object) GetInt(key string) int32 { return o.entries[key].val.(int32) }

// GetLong reads a LONG value.
func (o *Object) GetLong(key string) int64 { return o.entries[key].val.(int64) }

// GetUTFString reads a UTF_STRING value (empty string when absent or of a
// different type).
func (o *Object) GetUTFString(key string) string {
	v, ok := o.entries[key].val.(string)
	if !ok {
		return ""
	}
	return v
}

// GetSFSObject reads a nested SFS_OBJECT value (nil when absent or of a
// different type).
func (o *Object) GetSFSObject(key string) *Object {
	v, ok := o.entries[key].val.(*Object)
	if !ok {
		return nil
	}
	return v
}

// GetSFSArray reads a nested SFS_ARRAY value (nil when absent or of a
// different type).
func (o *Object) GetSFSArray(key string) []interface{} {
	v, ok := o.entries[key].val.([]interface{})
	if !ok {
		return nil
	}
	return v
}

// Encode serializes the object including the leading 0x12 marker (the
// top-level / standalone form).
func (o *Object) Encode() []byte {
	return append([]byte{byte(TSFSObject)}, o.entriesBin()...)
}

// entriesBin serializes count+entries (the nested form, without the 0x12
// marker). Keys are sorted for deterministic output.
func (o *Object) entriesBin() []byte {
	var buf bytes.Buffer
	keys := make([]string, len(o.keys))
	copy(keys, o.keys)
	sort.Strings(keys)
	buf.Write(encodeUint16(len(keys)))
	for _, key := range keys {
		e := o.entries[key]
		buf.Write(encodeUint16(len(key)))
		buf.WriteString(key)
		buf.WriteByte(byte(e.typ))
		encodeValue(&buf, e.typ, e.val)
	}
	return buf.Bytes()
}

func encodeValue(buf *bytes.Buffer, t Type, v interface{}) {
	switch t {
	case TNull:
	case TBool:
		if v.(bool) {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
	case TByte:
		buf.WriteByte(v.(byte))
	case TShort:
		buf.Write(encodeInt16(v.(int16)))
	case TInt:
		switch n := v.(type) {
		case int32:
			buf.Write(encodeInt32(n))
		case int:
			buf.Write(encodeInt32(int32(n)))
		}
	case TLong:
		switch n := v.(type) {
		case int64:
			buf.Write(encodeInt64(n))
		case int:
			buf.Write(encodeInt64(int64(n)))
		}
	case TFloat:
		buf.Write(encodeFloat32(float32(v.(float32))))
	case TDouble:
		buf.Write(encodeFloat64(v.(float64)))
	case TUTFString:
		s := v.(string)
		buf.Write(encodeUint16(len(s)))
		buf.WriteString(s)
	case TByteArray:
		b := v.([]byte)
		buf.Write(encodeUint32(len(b)))
		buf.Write(b)
	case TBoolArray:
		a := v.([]bool)
		buf.Write(encodeUint16(len(a)))
		for _, x := range a {
			if x {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		}
	case TShortArray:
		a := v.([]int16)
		writeTypedArray(buf, TShort, len(a), func(i int) interface{} { return a[i] })
	case TIntArray:
		a := v.([]int32)
		writeTypedArray(buf, TInt, len(a), func(i int) interface{} { return a[i] })
	case TLongArray:
		a := v.([]int64)
		writeTypedArray(buf, TLong, len(a), func(i int) interface{} { return a[i] })
	case TFloatArray:
		a := v.([]float32)
		writeTypedArray(buf, TFloat, len(a), func(i int) interface{} { return a[i] })
	case TDoubleArray:
		a := v.([]float64)
		writeTypedArray(buf, TDouble, len(a), func(i int) interface{} { return a[i] })
	case TUTFStringArray:
		a := v.([]string)
		buf.Write(encodeUint16(len(a)))
		for _, s := range a {
			buf.Write(encodeUint16(len(s)))
			buf.WriteString(s)
		}
	case TSFSArray:
		a := v.([]interface{})
		buf.Write(encodeUint16(len(a)))
		for _, item := range a {
			et, ev := inferType(item)
			buf.WriteByte(byte(et))
			encodeValue(buf, et, ev)
		}
	case TSFSObject:
		sub := v.(*Object)
		buf.Write(sub.entriesBin())
	}
}

func writeTypedArray(buf *bytes.Buffer, t Type, n int, get func(int) interface{}) {
	buf.Write(encodeUint16(n))
	for i := 0; i < n; i++ {
		encodeValue(buf, t, get(i))
	}
}

// Decode deserializes an SFSObject from binary data (which must start with
// the 0x12 marker).
func Decode(data []byte) (*Object, error) {
	if len(data) == 0 || data[0] != byte(TSFSObject) {
		return nil, fmt.Errorf("invalid SFSObject marker: %d", data[0])
	}
	p := &pos{off: 1} // skip 0x12
	return decodeEntries(data, p)
}

type pos struct {
	off int
}

// decodeEntries parses entries starting at the count field (the caller has
// already consumed the marker).
func decodeEntries(data []byte, p *pos) (*Object, error) {
	count := int(binary.BigEndian.Uint16(data[p.off:]))
	p.off += 2
	o := New()
	for i := 0; i < count; i++ {
		kl := int(binary.BigEndian.Uint16(data[p.off:]))
		p.off += 2
		key := string(data[p.off : p.off+kl])
		p.off += kl
		val, err := decodeValue(data, p)
		if err != nil {
			return nil, err
		}
		o.Put(key, val.typ, val.val)
	}
	return o, nil
}

type sfsEntry struct {
	typ Type
	val interface{}
}

func decodeValue(data []byte, p *pos) (sfsEntry, error) {
	t := Type(data[p.off])
	p.off++
	switch t {
	case TNull:
		return sfsEntry{t, nil}, nil
	case TBool:
		v := data[p.off]
		p.off++
		return sfsEntry{t, v != 0}, nil
	case TByte:
		v := data[p.off]
		p.off++
		return sfsEntry{t, v}, nil
	case TShort:
		v := int16(binary.BigEndian.Uint16(data[p.off:]))
		p.off += 2
		return sfsEntry{t, v}, nil
	case TInt:
		v := int32(binary.BigEndian.Uint32(data[p.off:]))
		p.off += 4
		return sfsEntry{t, v}, nil
	case TLong:
		v := int64(binary.BigEndian.Uint64(data[p.off:]))
		p.off += 8
		return sfsEntry{t, v}, nil
	case TFloat:
		v := math.Float32frombits(binary.BigEndian.Uint32(data[p.off:]))
		p.off += 4
		return sfsEntry{t, v}, nil
	case TDouble:
		v := math.Float64frombits(binary.BigEndian.Uint64(data[p.off:]))
		p.off += 8
		return sfsEntry{t, v}, nil
	case TUTFString:
		l := int(binary.BigEndian.Uint16(data[p.off:]))
		p.off += 2
		s := string(data[p.off : p.off+l])
		p.off += l
		return sfsEntry{t, s}, nil
	case TByteArray:
		l := int(binary.BigEndian.Uint32(data[p.off:]))
		p.off += 4
		b := make([]byte, l)
		copy(b, data[p.off:p.off+l])
		p.off += l
		return sfsEntry{t, b}, nil
	case TSFSArray:
		p.off-- // seek back: the type byte doubles as the marker
		n := int(binary.BigEndian.Uint16(data[p.off+1:]))
		p.off += 3
		a := make([]interface{}, n)
		for i := 0; i < n; i++ {
			e, err := decodeValue(data, p)
			if err != nil {
				return sfsEntry{}, err
			}
			a[i] = e.val
		}
		return sfsEntry{t, a}, nil
	case TSFSObject:
		p.off-- // seek back: the type byte doubles as the marker
		n := int(binary.BigEndian.Uint16(data[p.off+1:]))
		p.off += 3
		sub := New()
		for i := 0; i < n; i++ {
			kl := int(binary.BigEndian.Uint16(data[p.off:]))
			p.off += 2
			key := string(data[p.off : p.off+kl])
			p.off += kl
			e, err := decodeValue(data, p)
			if err != nil {
				return sfsEntry{}, err
			}
			sub.Put(key, e.typ, e.val)
		}
		return sfsEntry{t, sub}, nil
	}
	return sfsEntry{}, fmt.Errorf("unknown SFS type %d @ %d", t, p.off)
}

// ---- Framing ----

// EncodeFrame wraps an Object into an SFS frame: signature byte (0x80),
// uint16 payload length, payload.
func EncodeFrame(o *Object) []byte {
	payload := o.Encode()
	var buf bytes.Buffer
	buf.WriteByte(Signature)
	buf.Write(encodeUint16(len(payload)))
	buf.Write(payload)
	return buf.Bytes()
}

// DecodeFrame extracts the payload from an SFS frame and parses it into an
// Object. Both uint16 and uint32 (header bit 0x08) length forms are
// accepted; compressed frames are not handled (see package docs).
func DecodeFrame(data []byte) (*Object, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("frame too short: %d", len(data))
	}
	off := 1
	var ln int
	hdr := data[0]
	if hdr&LongLenMask != 0 {
		ln = int(binary.BigEndian.Uint32(data[off:]))
		off += 4
	} else {
		ln = int(binary.BigEndian.Uint16(data[off:]))
		off += 2
	}
	if off+ln > len(data) {
		return nil, fmt.Errorf("frame payload truncated")
	}
	return Decode(data[off : off+ln])
}

// ---- Encoding helpers ----

func encodeUint16(v int) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	return b
}

func encodeUint32(v int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return b
}

func encodeInt16(v int16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	return b
}

func encodeInt32(v int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return b
}

func encodeInt64(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func encodeFloat32(v float32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, math.Float32bits(v))
	return b
}

func encodeFloat64(v float64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
	return b
}

func inferType(v interface{}) (Type, interface{}) {
	switch val := v.(type) {
	case nil:
		return TNull, nil
	case bool:
		return TBool, val
	case byte:
		return TByte, val
	case int16:
		return TShort, val
	case int32:
		return TInt, val
	case int64:
		return TLong, val
	case float32:
		return TFloat, val
	case float64:
		return TDouble, val
	case string:
		return TUTFString, val
	case []byte:
		return TByteArray, val
	case *Object:
		return TSFSObject, val
	case []interface{}:
		return TSFSArray, val
	}
	return TUTFString, fmt.Sprintf("%v", v)
}
