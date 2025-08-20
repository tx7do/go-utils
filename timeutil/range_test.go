package timeutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentTimeRangeDateString(t *testing.T) {
	fmt.Println(GetTodayRangeDateString())
	fmt.Println(GetCurrentMonthRangeDateString())
	fmt.Println(GetCurrentYearRangeDateString())

	fmt.Println(GetYesterdayRangeDateString())
	fmt.Println(GetLastMonthRangeDateString())
	fmt.Println(GetLastYearRangeDateString())
}

func TestGetCurrentRangeTime(t *testing.T) {
	fmt.Println(GetTodayRangeTime())
	fmt.Println(GetCurrentMonthRangeTime())
	fmt.Println(GetCurrentYearRangeTime())

	fmt.Println(GetYesterdayRangeTime())
	fmt.Println(GetLastMonthRangeTime())
	fmt.Println(GetLastYearRangeTime())
}

func TestGetCurrentTimeRangeTimeString(t *testing.T) {
	fmt.Println(GetTodayRangeTimeString())
	fmt.Println(GetCurrentMonthRangeTimeString())
	fmt.Println(GetCurrentYearRangeTimeString())

	fmt.Println(GetYesterdayRangeTimeString())
	fmt.Println(GetLastMonthRangeTimeString())
	fmt.Println(GetLastYearRangeTimeString())
}

func TestRangeStringDateToTime(t *testing.T) {
	// 测试用例 1: 正常日期范围
	startDate := "2023-10-01"
	endDate := "2023-10-02"
	startTime, endTime := RangeStringDateToTime(startDate, endDate)

	assert.Equal(t, time.Date(2023, 10, 1, 0, 0, 0, 0, GetDefaultTimeLocation()), startTime)
	assert.Equal(t, time.Date(2023, 10, 2, 23, 59, 59, 0, GetDefaultTimeLocation()), endTime)

	// 测试用例 2: 只有开始日期
	startDate = "2023-10-01"
	endDate = ""
	startTime, endTime = RangeStringDateToTime(startDate, endDate)

	assert.Equal(t, time.Date(2023, 10, 1, 0, 0, 0, 0, GetDefaultTimeLocation()), startTime)
	assert.Equal(t, time.Date(2023, 10, 1, 23, 59, 59, 0, GetDefaultTimeLocation()), endTime)

	// 测试用例 3: 只有结束日期
	startDate = ""
	endDate = "2023-10-02"
	startTime, endTime = RangeStringDateToTime(startDate, endDate)

	assert.Equal(t, time.Time{}, startTime)
	assert.Equal(t, time.Date(2023, 10, 2, 23, 59, 59, 0, GetDefaultTimeLocation()), endTime)

	// 测试用例 4: 无效日期
	startDate = "invalid-date"
	endDate = "invalid-date"
	startTime, endTime = RangeStringDateToTime(startDate, endDate)

	assert.Equal(t, time.Time{}, startTime)
	assert.Equal(t, time.Time{}, endTime)

	// 测试用例 5: 起始时间和结束时间相同
	startDate = "2023-10-01"
	endDate = "2023-10-01"
	startTime, endTime = RangeStringDateToTime(startDate, endDate)

	assert.Equal(t, time.Date(2023, 10, 1, 0, 0, 0, 0, GetDefaultTimeLocation()), startTime)
	assert.Equal(t, time.Date(2023, 10, 1, 23, 59, 59, 0, GetDefaultTimeLocation()), endTime)
}
