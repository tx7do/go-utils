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
