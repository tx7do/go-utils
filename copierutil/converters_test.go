package copierutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
)

func TestNewTypeConverter(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fn := func(src interface{}) (interface{}, error) {
		return timeutil.TimeToTimeString(src.(*time.Time)), nil
	}

	converter := NewTypeConverter(srcType, dstType, fn)

	// 验证转换器的类型
	assert.IsType(t, srcType, converter.SrcType)
	assert.IsType(t, dstType, converter.DstType)

	// 验证转换器的功能
	result, err := converter.Fn(&time.Time{})
	assert.NoError(t, err)
	assert.IsType(t, dstType, result)
}

func TestNewTypeConverterPair(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fromFn := func(src interface{}) (interface{}, error) {
		return timeutil.TimeToTimeString(src.(*time.Time)), nil
	}
	toFn := func(src interface{}) (interface{}, error) {
		return timeutil.StringTimeToTime(src.(*string)), nil
	}

	converters := NewTypeConverterPair(srcType, dstType, fromFn, toFn)
	assert.Len(t, converters, 2, "expected 2 converters")

	// 验证第一个转换器
	assert.IsType(t, srcType, converters[0].SrcType)
	assert.IsType(t, dstType, converters[0].DstType)
	result, err := converters[0].Fn(&time.Time{})
	assert.NoError(t, err)
	assert.IsType(t, dstType, result)

	// 验证第二个转换器
	assert.IsType(t, dstType, converters[1].SrcType)
	assert.IsType(t, srcType, converters[1].DstType)
	result, err = converters[1].Fn(trans.Ptr(""))
	assert.NoError(t, err)
	assert.IsType(t, srcType, result)
}

func TestNewGenericTypeConverterPair(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fromFn := timeutil.TimeToTimeString
	toFn := timeutil.StringTimeToTime

	converters := NewGenericTypeConverterPair(srcType, dstType, fromFn, toFn)
	assert.Len(t, converters, 2, "expected 2 converters")

	// 验证第一个转换器
	assert.IsType(t, srcType, converters[0].SrcType)
	assert.IsType(t, dstType, converters[0].DstType)
	result, err := converters[0].Fn(&time.Time{})
	assert.NoError(t, err)
	assert.IsType(t, dstType, result)

	// 验证第二个转换器
	assert.IsType(t, dstType, converters[1].SrcType)
	assert.IsType(t, srcType, converters[1].DstType)
	result, err = converters[1].Fn(trans.Ptr(""))
	assert.NoError(t, err)
	assert.IsType(t, srcType, result)
}

func TestNewErrorHandlingGenericTypeConverterPair(t *testing.T) {
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	fromFn := func(src *time.Time) (*string, error) {
		return timeutil.TimeToTimeString(src), nil
	}
	toFn := func(src *string) (*time.Time, error) {
		return timeutil.StringTimeToTime(src), nil
	}

	converters := NewErrorHandlingGenericTypeConverterPair(srcType, dstType, fromFn, toFn)
	assert.Len(t, converters, 2, "expected 2 converters")

	// 验证第一个转换器
	assert.IsType(t, srcType, converters[0].SrcType)
	assert.IsType(t, dstType, converters[0].DstType)
	result, err := converters[0].Fn(&time.Time{})
	assert.NoError(t, err)
	assert.IsType(t, dstType, result)

	// 验证第二个转换器
	assert.IsType(t, dstType, converters[1].SrcType)
	assert.IsType(t, srcType, converters[1].DstType)
	result, err = converters[1].Fn(trans.Ptr(""))
	assert.NoError(t, err)
	assert.IsType(t, srcType, result)
}

func TestNewTimeStringConverterPair(t *testing.T) {
	converters := NewTimeStringConverterPair()
	assert.Len(t, converters, 2, "expected 2 converters")

	// 验证第一个转换器
	srcType := &time.Time{}
	dstType := trans.Ptr("")
	assert.IsType(t, srcType, converters[0].SrcType)
	assert.IsType(t, dstType, converters[0].DstType)
	result, err := converters[0].Fn(&time.Time{})
	assert.NoError(t, err)
	assert.IsType(t, dstType, result)

	// 验证第二个转换器
	assert.IsType(t, dstType, converters[1].SrcType)
	assert.IsType(t, srcType, converters[1].DstType)
	result, err = converters[1].Fn(trans.Ptr(""))
	assert.NoError(t, err)
	assert.IsType(t, srcType, result)
}

func TestNewTimeTimestamppbConverterPair(t *testing.T) {
	converters := NewTimeTimestamppbConverterPair()
	assert.Len(t, converters, 2, "expected 2 converters")

	// 验证第一个转换器
	srcType := &time.Time{}
	dstType := &timestamppb.Timestamp{}
	assert.IsType(t, srcType, converters[0].SrcType)
	assert.IsType(t, dstType, converters[0].DstType)
	result, err := converters[0].Fn(&time.Time{})
	assert.NoError(t, err)
	assert.IsType(t, dstType, result)

	// 验证第二个转换器
	assert.IsType(t, dstType, converters[1].SrcType)
	assert.IsType(t, srcType, converters[1].DstType)
	result, err = converters[1].Fn(&timestamppb.Timestamp{})
	assert.NoError(t, err)
	assert.IsType(t, srcType, result)
}
