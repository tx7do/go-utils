package util

import (
	"time"

	"github.com/tx7do/kratos-utils/trans"
)

var DefaultTimeLocation *time.Location

func RefreshDefaultTimeLocation(name string) {
	DefaultTimeLocation, _ = time.LoadLocation(name)
}

// UnixMilliToStringPtr 毫秒时间戳 -> 字符串
func UnixMilliToStringPtr(tm *int64) *string {
	if tm == nil {
		return nil
	}
	str := time.UnixMilli(*tm).Format(TimeLayout)
	return &str
}

// StringToUnixMilliInt64Ptr 字符串 -> 毫秒时间戳
func StringToUnixMilliInt64Ptr(tm *string) *int64 {
	theTime := StringTimeToTime(tm)
	if theTime == nil {
		return nil
	}
	unixTime := theTime.UnixMilli()
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

	if DefaultTimeLocation == nil {
		RefreshDefaultTimeLocation(DefaultTimeLocationName)
	}

	var err error
	var theTime time.Time

	theTime, err = time.ParseInLocation(TimeLayout, *str, DefaultTimeLocation)
	if err == nil {
		return &theTime
	}

	theTime, err = time.ParseInLocation(DateLayout, *str, DefaultTimeLocation)
	if err == nil {
		return &theTime
	}

	theTime, err = time.ParseInLocation(ClockLayout, *str, DefaultTimeLocation)
	if err == nil {
		return &theTime
	}

	return nil
}

// TimeToTimeString 时间 -> 时间字符串
func TimeToTimeString(tm *time.Time) *string {
	if tm == nil {
		return nil
	}
	return trans.String(tm.Format(TimeLayout))
}

// StringDateToTime 字符串 -> 时间
func StringDateToTime(str *string) *time.Time {
	if str == nil {
		return nil
	}
	if len(*str) == 0 {
		return nil
	}

	if DefaultTimeLocation == nil {
		RefreshDefaultTimeLocation(DefaultTimeLocationName)
	}

	var err error
	var theTime time.Time

	theTime, err = time.ParseInLocation(TimeLayout, *str, DefaultTimeLocation)
	if err == nil {
		return &theTime
	}

	theTime, err = time.ParseInLocation(DateLayout, *str, DefaultTimeLocation)
	if err == nil {
		return &theTime
	}

	theTime, err = time.ParseInLocation(ClockLayout, *str, DefaultTimeLocation)
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
	return trans.String(tm.Format(DateLayout))
}
