package copierutil

import (
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
)

var TimeToStringConverter = copier.TypeConverter{
	SrcType: &time.Time{},  // 源类型
	DstType: trans.Ptr(""), // 目标类型
	Fn: func(src any) (any, error) {
		return timeutil.TimeToTimeString(src.(*time.Time)), nil
	},
}

var StringToTimeConverter = copier.TypeConverter{
	SrcType: trans.Ptr(""),
	DstType: &time.Time{},
	Fn: func(src any) (any, error) {
		return timeutil.StringTimeToTime(src.(*string)), nil
	},
}

var TimeToTimestamppbConverter = copier.TypeConverter{
	SrcType: &time.Time{},
	DstType: &timestamppb.Timestamp{},
	Fn: func(src any) (any, error) {
		return timeutil.TimeToTimestamppb(src.(*time.Time)), nil
	},
}

var TimestamppbToTimeConverter = copier.TypeConverter{
	SrcType: &timestamppb.Timestamp{},
	DstType: &time.Time{},
	Fn: func(src any) (any, error) {
		return timeutil.TimestamppbToTime(src.(*timestamppb.Timestamp)), nil
	},
}

func TimeToString(tm *time.Time) *string {
	return timeutil.TimeToString(tm, timeutil.ISO8601)
}

func NewTimeStringConverterPair() []copier.TypeConverter {
	srcType := &time.Time{}
	dstType := trans.Ptr("")

	fromFn := TimeToString
	toFn := timeutil.StringTimeToTime

	return NewGenericTypeConverterPair(srcType, dstType, fromFn, toFn)
}

func NewTimeTimestamppbConverterPair() []copier.TypeConverter {
	srcType := &time.Time{}
	dstType := &timestamppb.Timestamp{}

	fromFn := timeutil.TimeToTimestamppb
	toFn := timeutil.TimestamppbToTime

	return NewGenericTypeConverterPair(srcType, dstType, fromFn, toFn)
}

var StringToTimestamppbConverter = copier.TypeConverter{
	SrcType: trans.Ptr(""),
	DstType: &timestamppb.Timestamp{},
	Fn: func(src any) (any, error) {
		return timeutil.StringToTimestamppb(src.(*string)), nil
	},
}

var TimestamppbToStringConverter = copier.TypeConverter{
	SrcType: &timestamppb.Timestamp{},
	DstType: trans.Ptr(""),
	Fn: func(src any) (any, error) {
		return timeutil.TimestamppbToString(src.(*timestamppb.Timestamp)), nil
	},
}

func NewStringTimestamppbConverterPair() []copier.TypeConverter {
	srcType := trans.Ptr("")
	dstType := &timestamppb.Timestamp{}

	fromFn := timeutil.StringToTimestamppb
	toFn := timeutil.TimestamppbToString

	return NewGenericTypeConverterPair(srcType, dstType, fromFn, toFn)
}

func NewDurationpbNumberConverterPair[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](timePrecision time.Duration) []copier.TypeConverter {
	srcType := &durationpb.Duration{}
	dstType := trans.Ptr(T(0))

	fromFn := func(src *durationpb.Duration) *T {
		return timeutil.DurationpbToNumber[T](src, timePrecision)
	}
	toFn := func(src *T) *durationpb.Duration {
		return timeutil.NumberToDurationpb(src, timePrecision)
	}

	return NewGenericTypeConverterPair(srcType, dstType, fromFn, toFn)
}

func NewTypeConverter(srcType, dstType any, fn func(src any) (any, error)) copier.TypeConverter {
	return copier.TypeConverter{
		SrcType: srcType,
		DstType: dstType,
		Fn:      fn,
	}
}

func NewTypeConverterPair(srcType, dstType any, fromFn, toFn func(src any) (any, error)) []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn:      fromFn,
		},
		{
			SrcType: dstType,
			DstType: srcType,
			Fn:      toFn,
		},
	}
}

func NewGenericTypeConverterPair[A any, B any](srcType A, dstType B, fromFn func(src A) B, toFn func(src B) A) []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn: func(src any) (any, error) {
				return fromFn(src.(A)), nil
			},
		},
		{
			SrcType: dstType,
			DstType: srcType,
			Fn: func(src any) (any, error) {
				return toFn(src.(B)), nil
			},
		},
	}
}

func NewErrorHandlingGenericTypeConverterPair[A any, B any](srcType A, dstType B, fromFn func(src A) (B, error), toFn func(src B) (A, error)) []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn: func(src any) (any, error) {
				return fromFn(src.(A))
			},
		},
		{
			SrcType: dstType,
			DstType: srcType,
			Fn: func(src any) (any, error) {
				return toFn(src.(B))
			},
		},
	}
}
