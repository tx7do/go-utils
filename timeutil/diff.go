package util

import (
	"math"
	"time"
)

// DayDifferenceHours 两天之间相差了多少小时
func DayDifferenceHours(startDate, endDate string) float64 {
	startTime, _ := time.Parse(DateLayout, startDate)
	endTime, _ := time.Parse(DateLayout, endDate)

	duration := endTime.Sub(startTime)

	return duration.Hours()
}

// StringDifferenceDays 两天之间相差了多少天
func StringDifferenceDays(startDate, endDate string) int {
	hours := DayDifferenceHours(startDate, endDate)
	if hours == 0 {
		return 0
	}
	return int(math.Ceil(hours / 24))
}

func DayTimeDifferenceHours(startDate, endDate time.Time) float64 {
	startTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.Local)

	duration := endTime.Sub(startTime)

	return duration.Hours()
}

// TimeDifferenceDays 两天之间相差了多少天
func TimeDifferenceDays(startDate, endDate time.Time) int {
	hours := DayTimeDifferenceHours(startDate, endDate)
	if hours == 0 {
		return 0
	}
	return int(math.Ceil(hours / 24))
}

func DaySecondsDifferenceHours(startSecond, endSecond int64) float64 {
	startTime := time.Unix(startSecond, 0)
	endTime := time.Unix(endSecond, 0)

	duration := endTime.Sub(startTime)

	return duration.Hours()
}

// SecondsDifferenceDays 两天之间相差了多少天
func SecondsDifferenceDays(startSecond, endSecond int64) int {
	hours := DaySecondsDifferenceHours(startSecond, endSecond)
	if hours == 0 {
		return 0
	}
	return int(math.Ceil(hours / 24))
}
