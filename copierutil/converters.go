package copierutil

import (
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
)

var TimeToStringConverter = copier.TypeConverter{
	SrcType: &time.Time{},  // 源类型
	DstType: trans.Ptr(""), // 目标类型
	Fn: func(src interface{}) (interface{}, error) {
		return timeutil.TimeToTimeString(src.(*time.Time)), nil
	},
}

var StringToTimeConverter = copier.TypeConverter{
	SrcType: trans.Ptr(""),
	DstType: &time.Time{},
	Fn: func(src interface{}) (interface{}, error) {
		return timeutil.StringTimeToTime(src.(*string)), nil
	},
}

var TimeToTimestamppbConverter = copier.TypeConverter{
	SrcType: &time.Time{},
	DstType: &timestamppb.Timestamp{},
	Fn: func(src interface{}) (interface{}, error) {
		return timeutil.TimeToTimestamppb(src.(*time.Time)), nil
	},
}

var TimestamppbToTimeConverter = copier.TypeConverter{
	SrcType: &timestamppb.Timestamp{},
	DstType: &time.Time{},
	Fn: func(src interface{}) (interface{}, error) {
		return timeutil.TimestamppbToTime(src.(*timestamppb.Timestamp)), nil
	},
}

func MakeTypeConverter(srcType, dstType interface{}, fn func(src interface{}) (interface{}, error)) copier.TypeConverter {
	return copier.TypeConverter{
		SrcType: srcType,
		DstType: dstType,
		Fn:      fn,
	}
}

func MakeTypeConverterPair(srcType, dstType interface{}, fromFn, toFn func(src interface{}) (interface{}, error)) []copier.TypeConverter {
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

func MakeGenericTypeConverterPair[A interface{}, B interface{}](srcType A, dstType B, fromFn func(src A) B, toFn func(src B) A) []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn: func(src interface{}) (interface{}, error) {
				return fromFn(src.(A)), nil
			},
		},
		{
			SrcType: dstType,
			DstType: srcType,
			Fn: func(src interface{}) (interface{}, error) {
				return toFn(src.(B)), nil
			},
		},
	}
}

func MakeErrorHandlingTypeConverterPair[A interface{}, B interface{}](srcType A, dstType B, fromFn func(src A) (B, error), toFn func(src B) (A, error)) []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn: func(src interface{}) (interface{}, error) {
				return fromFn(src.(A))
			},
		},
		{
			SrcType: dstType,
			DstType: srcType,
			Fn: func(src interface{}) (interface{}, error) {
				return toFn(src.(B))
			},
		},
	}
}
