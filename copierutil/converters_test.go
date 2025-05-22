package copierutil

import (
	"testing"
	"time"

	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
)

func TestMakeTypeConverter(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fn := func(src interface{}) (interface{}, error) {
		return timeutil.TimeToTimeString(src.(*time.Time)), nil
	}

	converter := MakeTypeConverter(srcType, dstType, fn)

	// 验证转换器的类型
	if converter.SrcType != srcType || converter.DstType != dstType {
		t.Errorf("converter types mismatch")
	}

	// 验证转换器的功能
	result, err := converter.Fn(&time.Time{})
	if err != nil {
		t.Errorf("converter function failed: %v", err)
	}
	if _, ok := result.(*string); !ok {
		t.Errorf("converter result type mismatch")
	}
}

func TestMakeTypeConverterPair(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fromFn := func(src interface{}) (interface{}, error) {
		return timeutil.TimeToTimeString(src.(*time.Time)), nil
	}
	toFn := func(src interface{}) (interface{}, error) {
		return timeutil.StringTimeToTime(src.(*string)), nil
	}

	converters := MakeTypeConverterPair(srcType, dstType, fromFn, toFn)

	if len(converters) != 2 {
		t.Fatalf("expected 2 converters, got %d", len(converters))
	}

	// 验证第一个转换器
	if converters[0].SrcType != srcType || converters[0].DstType != dstType {
		t.Errorf("first converter types mismatch")
	}
	result, err := converters[0].Fn(&time.Time{})
	if err != nil {
		t.Errorf("first converter function failed: %v", err)
	}
	if _, ok := result.(*string); !ok {
		t.Errorf("first converter result type mismatch")
	}

	// 验证第二个转换器
	if converters[1].SrcType != dstType || converters[1].DstType != srcType {
		t.Errorf("second converter types mismatch")
	}
	result, err = converters[1].Fn(trans.Ptr(""))
	if err != nil {
		t.Errorf("second converter function failed: %v", err)
	}
	if _, ok := result.(*time.Time); !ok {
		t.Errorf("second converter result type mismatch")
	}
}

func TestMakeGenericTypeConverterPair(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fromFn := timeutil.TimeToTimeString
	toFn := timeutil.StringTimeToTime

	converters := MakeGenericTypeConverterPair(srcType, dstType, fromFn, toFn)

	if len(converters) != 2 {
		t.Fatalf("expected 2 converters, got %d", len(converters))
	}

	// 验证第一个转换器
	if converters[0].SrcType != srcType || converters[0].DstType != dstType {
		t.Errorf("first converter types mismatch")
	}
	result, err := converters[0].Fn(&time.Time{})
	if err != nil {
		t.Errorf("first converter function failed: %v", err)
	}
	if _, ok := result.(*string); !ok {
		t.Errorf("first converter result type mismatch")
	}

	// 验证第二个转换器
	if converters[1].SrcType != dstType || converters[1].DstType != srcType {
		t.Errorf("second converter types mismatch")
	}
	result, err = converters[1].Fn(trans.Ptr(""))
	if err != nil {
		t.Errorf("second converter function failed: %v", err)
	}
	if _, ok := result.(*time.Time); !ok {
		t.Errorf("second converter result type mismatch")
	}
}

func TestMakeErrorHandlingTypeConverterPair(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fromFn := func(src *time.Time) (*string, error) {
		return timeutil.TimeToTimeString(src), nil
	}
	toFn := func(src *string) (*time.Time, error) {
		return timeutil.StringTimeToTime(src), nil
	}

	converters := MakeErrorHandlingTypeConverterPair(srcType, dstType, fromFn, toFn)

	if len(converters) != 2 {
		t.Fatalf("expected 2 converters, got %d", len(converters))
	}

	// 验证第一个转换器
	if converters[0].SrcType != srcType || converters[0].DstType != dstType {
		t.Errorf("first converter types mismatch")
	}
	result, err := converters[0].Fn(&time.Time{})
	if err != nil {
		t.Errorf("first converter function failed: %v", err)
	}
	if _, ok := result.(*string); !ok {
		t.Errorf("first converter result type mismatch")
	}

	// 验证第二个转换器
	if converters[1].SrcType != dstType || converters[1].DstType != srcType {
		t.Errorf("second converter types mismatch")
	}
	result, err = converters[1].Fn(trans.Ptr(""))
	if err != nil {
		t.Errorf("second converter function failed: %v", err)
	}
	if _, ok := result.(*time.Time); !ok {
		t.Errorf("second converter result type mismatch")
	}
}
