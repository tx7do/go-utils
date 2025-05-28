package timeutil

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tx7do/go-utils/trans"
)

var defaultTimeLocation *time.Location

func RefreshDefaultTimeLocation(name string) *time.Location {
	if defaultTimeLocation == nil {
		defaultTimeLocation, _ = time.LoadLocation(name)
	}
	return defaultTimeLocation
}

func GetDefaultTimeLocation() *time.Location {
	if defaultTimeLocation == nil {
		RefreshDefaultTimeLocation(DefaultTimeLocationName)
	}
	return defaultTimeLocation
}

// UnixMilliToStringPtr 毫秒时间戳 -> 字符串
func UnixMilliToStringPtr(milli *int64) *string {
	if milli == nil {
		return nil
	}

	tm := time.UnixMilli(*milli)

	str := tm.In(GetDefaultTimeLocation()).Format(TimeLayout)
	return &str
}

// StringToUnixMilliInt64Ptr 字符串 -> 毫秒时间戳
func StringToUnixMilliInt64Ptr(tm *string) *int64 {
	if tm == nil {
		return nil
	}

	theTime := StringTimeToTime(tm)
	if theTime == nil {
		return nil
	}

	unixTime := theTime.UnixMilli()
	return &unixTime
}

// UnixMilliToTimePtr 毫秒时间戳 -> 时间
func UnixMilliToTimePtr(milli *int64) *time.Time {
	if milli == nil {
		return nil
	}

	unixMilli := time.UnixMilli(*milli)
	return &unixMilli
}

// TimeToUnixMilliInt64Ptr 时间 -> 毫秒时间戳
func TimeToUnixMilliInt64Ptr(tm *time.Time) *int64 {
	if tm == nil {
		return nil
	}

	unixTime := tm.UnixMilli()
	return &unixTime
}

// UnixSecondToTimePtr 秒时间戳 -> 时间
func UnixSecondToTimePtr(second *int64) *time.Time {
	if second == nil {
		return nil
	}

	unixMilli := time.Unix(*second, 0)

	return &unixMilli
}

// TimeToUnixSecondInt64Ptr 时间 -> 秒时间戳
func TimeToUnixSecondInt64Ptr(tm *time.Time) *int64 {
	if tm == nil {
		return nil
	}

	unixTime := tm.Unix()
	return &unixTime
}

// StringTimeToTime 时间字符串 -> 时间
func StringTimeToTime(str *string) *time.Time {
	if str == nil {
		return nil
	}
	if len(*str) == 0 {
		return nil
	}

	var err error
	var theTime time.Time

	if theTime, err = time.ParseInLocation(TimeLayout, *str, GetDefaultTimeLocation()); err == nil {
		return &theTime
	}

	if theTime, err = time.ParseInLocation(DateLayout, *str, GetDefaultTimeLocation()); err == nil {
		return &theTime
	}

	if theTime, err = time.ParseInLocation(ClockLayout, *str, GetDefaultTimeLocation()); err == nil {
		return &theTime
	}

	if theTime, err = time.ParseInLocation(ISO9075MicroTZ, *str, GetDefaultTimeLocation()); err == nil {
		return &theTime
	}

	return nil
}

// TimeToTimeString 时间 -> 时间字符串
func TimeToTimeString(tm *time.Time) *string {
	if tm == nil {
		return nil
	}

	return trans.String(tm.In(GetDefaultTimeLocation()).Format(TimeLayout))
}

// StringDateToTime 字符串 -> 时间
func StringDateToTime(str *string) *time.Time {
	if str == nil {
		return nil
	}
	if len(*str) == 0 {
		return nil
	}

	var err error
	var theTime time.Time

	theTime, err = time.ParseInLocation(TimeLayout, *str, GetDefaultTimeLocation())
	if err == nil {
		return &theTime
	}

	theTime, err = time.ParseInLocation(DateLayout, *str, GetDefaultTimeLocation())
	if err == nil {
		return &theTime
	}

	theTime, err = time.ParseInLocation(ClockLayout, *str, GetDefaultTimeLocation())
	if err == nil {
		return &theTime
	}

	return nil
}

// TimeToDateString 时间 -> 日期字符串
func TimeToDateString(tm *time.Time) *string {
	if tm == nil {
		return nil
	}

	return trans.String(tm.In(GetDefaultTimeLocation()).Format(DateLayout))
}

// TimestamppbToTime timestamppb.Timestamp -> time.Time
func TimestamppbToTime(timestamp *timestamppb.Timestamp) *time.Time {
	if timestamp != nil {
		return trans.Ptr(timestamp.AsTime())
	}
	return nil
}

// TimeToTimestamppb time.Time -> timestamppb.Timestamp
func TimeToTimestamppb(tm *time.Time) *timestamppb.Timestamp {
	if tm != nil {
		return timestamppb.New(*tm)
	}
	return nil
}

func FloatToDurationpb(duration *float64, timePrecision time.Duration) *durationpb.Duration {
	if duration == nil {
		return nil
	}
	return durationpb.New(time.Duration(*duration * float64(timePrecision)))
}

func Float64ToDurationpb(d float64) *durationpb.Duration {
	duration := time.Duration(d * float64(time.Second))
	return durationpb.New(duration)
}

func SecondToDurationpb(seconds *float64) *durationpb.Duration {
	return FloatToDurationpb(seconds, time.Second)
}

func DurationpbToFloat(duration *durationpb.Duration, timePrecision time.Duration) *float64 {
	if duration == nil {
		return nil
	}
	seconds := duration.AsDuration().Seconds()
	secondsWithPrecision := seconds / timePrecision.Seconds()
	return &secondsWithPrecision
}

func NumberToDurationpb[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](duration *T, timePrecision time.Duration) *durationpb.Duration {
	if duration == nil {
		return nil
	}
	return durationpb.New(time.Duration(*duration) * timePrecision)
}

func DurationpbToNumber[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](duration *durationpb.Duration, timePrecision time.Duration) *T {
	if duration == nil {
		return nil
	}
	seconds := duration.AsDuration().Seconds()
	secondsWithPrecision := T(seconds / timePrecision.Seconds())
	return &secondsWithPrecision
}

func DurationToDurationpb(duration *time.Duration) *durationpb.Duration {
	if duration == nil {
		return nil
	}
	return durationpb.New(*duration)
}

func DurationpbToDuration(duration *durationpb.Duration) *time.Duration {
	if duration == nil {
		return nil
	}
	d := duration.AsDuration()
	return &d
}

func DurationpbToSecond(duration *durationpb.Duration) *float64 {
	if duration == nil {
		return nil
	}
	seconds := duration.AsDuration().Seconds()
	secondsInt64 := seconds
	return &secondsInt64
}

func StringToDurationpb(in *string) *durationpb.Duration {
	if in == nil {
		return nil
	}

	f, _ := time.ParseDuration(*in)
	return durationpb.New(f)
}

func DurationpbToString(in *durationpb.Duration) *string {
	if in == nil {
		return nil
	}

	return trans.Ptr(in.AsDuration().String())
}
