package timeutil

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestUnixMilliToStringPtr(t *testing.T) {
	now := time.Now().UnixMilli()
	str := UnixMilliToStringPtr(&now)
	fmt.Println(now)
	fmt.Println(*str)

	fmt.Println(*UnixMilliToStringPtr(trans.Int64(1677135885288)))
	fmt.Println(*UnixMilliToStringPtr(trans.Int64(1677647430853)))
	fmt.Println(*UnixMilliToStringPtr(trans.Int64(1677647946234)))
	fmt.Println(*UnixMilliToStringPtr(trans.Int64(1678245350773)))

	fmt.Println("START: ", *StringToUnixMilliInt64Ptr(trans.Ptr("2023-03-09 00:00:00")))
	fmt.Println("END: ", *StringToUnixMilliInt64Ptr(trans.Ptr("2023-03-09 23:59:59")))

	fmt.Println(StringTimeToTime(trans.Ptr("2023-03-01 00:00:00")))
	fmt.Println(*StringDateToTime(trans.Ptr("2023-03-01")))

	fmt.Println(StringTimeToTime(trans.Ptr("2023-03-08 00:00:00")).UnixMilli())
	fmt.Println(StringDateToTime(trans.Ptr("2023-03-07")).UnixMilli())

	// 测试有效输入
	now = time.Now().UnixMilli()
	result := UnixMilliToStringPtr(&now)
	assert.NotNil(t, result)
	expected := time.UnixMilli(now).Format(TimeLayout)
	assert.Equal(t, expected, *result)

	// 测试空输入
	result = UnixMilliToStringPtr(nil)
	assert.Nil(t, result)
}

func TestStringToUnixMilliInt64Ptr(t *testing.T) {
	// 测试有效输入
	input := "2023-03-09 00:00:00"
	expected := time.Date(2023, 3, 9, 0, 0, 0, 0, GetDefaultTimeLocation()).UnixMilli()
	result := StringToUnixMilliInt64Ptr(&input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, *result)

	// 测试无效输入
	invalidInput := "invalid-date"
	result = StringToUnixMilliInt64Ptr(&invalidInput)
	assert.Nil(t, result)

	// 测试空字符串输入
	emptyInput := ""
	result = StringToUnixMilliInt64Ptr(&emptyInput)
	assert.Nil(t, result)

	// 测试空指针输入
	result = StringToUnixMilliInt64Ptr(nil)
	assert.Nil(t, result)
}

func TestStringTimeToTime(t *testing.T) {
	// 测试有效时间字符串输入
	input := "2023-03-09 12:34:56"
	expected := time.Date(2023, 3, 9, 12, 34, 56, 0, GetDefaultTimeLocation())
	result := StringTimeToTime(&input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, *result)

	// 测试有效日期字符串输入
	input = "2023-03-09"
	expected = time.Date(2023, 3, 9, 0, 0, 0, 0, GetDefaultTimeLocation())
	result = StringTimeToTime(&input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, *result)

	// 测试无效时间字符串输入
	invalidInput := "invalid-date"
	result = StringTimeToTime(&invalidInput)
	assert.Nil(t, result)

	// 测试空字符串输入
	emptyInput := ""
	result = StringTimeToTime(&emptyInput)
	assert.Nil(t, result)

	// 测试空指针输入
	result = StringTimeToTime(nil)
	assert.Nil(t, result)
}

func TestTimeToTimeString(t *testing.T) {
	// 测试非空输入
	now := time.Now()
	result := TimeToTimeString(&now)
	assert.NotNil(t, result)
	expected := now.Format(TimeLayout)
	assert.Equal(t, expected, *result)

	// 测试空输入
	result = TimeToTimeString(nil)
	assert.Nil(t, result)
}

func TestStringDateToTime(t *testing.T) {
	// 测试有效日期字符串输入
	input := "2023-03-09"
	expected := time.Date(2023, 3, 9, 0, 0, 0, 0, GetDefaultTimeLocation())
	result := StringDateToTime(&input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, *result)

	// 测试无效日期字符串输入
	invalidInput := "invalid-date"
	result = StringDateToTime(&invalidInput)
	assert.Nil(t, result)

	// 测试空字符串输入
	emptyInput := ""
	result = StringDateToTime(&emptyInput)
	assert.Nil(t, result)

	// 测试空指针输入
	result = StringDateToTime(nil)
	assert.Nil(t, result)
}

func TestTimeToDateString(t *testing.T) {
	fmt.Println(*TimeToTimeString(trans.Time(time.Now())))
	fmt.Println(*TimeToDateString(trans.Time(time.Now())))

	// 测试非空输入
	now := time.Now()
	result := TimeToDateString(&now)
	assert.NotNil(t, result)
	expected := now.Format(DateLayout)
	assert.Equal(t, expected, *result)

	// 测试空输入
	result = TimeToDateString(nil)
	assert.Nil(t, result)
}

func TestTimestamppbToTime(t *testing.T) {
	// 测试有效输入
	timestamp := timestamppb.Now()
	result := TimestamppbToTime(timestamp)
	assert.NotNil(t, result)
	assert.Equal(t, timestamp.AsTime(), *result)

	// 测试零时间输入
	zeroTimestamp := timestamppb.New(time.Time{})
	result = TimestamppbToTime(zeroTimestamp)
	assert.NotNil(t, result)
	assert.Equal(t, time.Time{}, *result)

	// 测试空输入
	result = TimestamppbToTime(nil)
	assert.Nil(t, result)
}

func TestTimeToTimestamppb(t *testing.T) {
	// 测试非空输入
	now := time.Now()
	result := TimeToTimestamppb(&now)
	assert.NotNil(t, result)
	assert.Equal(t, timestamppb.New(now), result)

	// 测试空输入
	result = TimeToTimestamppb(nil)
	assert.Nil(t, result)
}

func TestDurationpb(t *testing.T) {
	fmt.Println(FloatToDurationpb(trans.Ptr(100.0), time.Nanosecond).String())
	fmt.Println(*DurationpbToFloat(durationpb.New(100*time.Nanosecond), time.Nanosecond))

	fmt.Println(FloatToDurationpb(trans.Ptr(100.0), time.Second).String())
	fmt.Println(*DurationpbToFloat(durationpb.New(100*time.Second), time.Second))

	fmt.Println(FloatToDurationpb(trans.Ptr(100.0), time.Minute).String())
	fmt.Println(*DurationpbToFloat(durationpb.New(100*time.Minute), time.Minute))

	//

	fmt.Println(NumberToDurationpb(trans.Ptr(100.0), time.Nanosecond).String())
	fmt.Println(*DurationpbToNumber[float64](durationpb.New(100*time.Nanosecond), time.Nanosecond))

	fmt.Println(NumberToDurationpb(trans.Ptr(100.0), time.Second).String())
	fmt.Println(*DurationpbToNumber[float64](durationpb.New(100*time.Second), time.Second))

	fmt.Println(NumberToDurationpb(trans.Ptr(100.0), time.Minute).String())
	fmt.Println(*DurationpbToNumber[float64](durationpb.New(100*time.Minute), time.Minute))
}

func TestFloatToDurationpb(t *testing.T) {
	// 测试有效输入
	input := 1.5 // 1.5秒
	timePrecision := time.Second
	expected := durationpb.New(1500 * time.Millisecond)
	result := FloatToDurationpb(&input, timePrecision)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)

	// 测试零输入
	input = 0.0
	expected = durationpb.New(0)
	result = FloatToDurationpb(&input, timePrecision)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)

	// 测试空输入
	result = FloatToDurationpb(nil, timePrecision)
	assert.Nil(t, result)
}

func TestDurationToDurationpb(t *testing.T) {
	// 测试非空输入
	duration := time.Duration(5 * time.Second)
	result := DurationToDurationpb(&duration)
	assert.NotNil(t, result)
	assert.Equal(t, durationpb.New(duration), result)

	// 测试空输入
	result = DurationToDurationpb(nil)
	assert.Nil(t, result)
}

func TestDurationpbToDuration(t *testing.T) {
	// 测试非空输入
	durationpbValue := durationpb.New(5 * time.Second)
	result := DurationpbToDuration(durationpbValue)
	assert.NotNil(t, result)
	assert.Equal(t, 5*time.Second, *result)

	// 测试空输入
	result = DurationpbToDuration(nil)
	assert.Nil(t, result)
}

func TestFloat64ToDurationpb(t *testing.T) {
	// 测试有效输入
	input := 1.5 // 1.5秒
	expected := durationpb.New(1500 * time.Millisecond)
	result := Float64ToDurationpb(input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)

	// 测试零输入
	input = 0.0
	expected = durationpb.New(0)
	result = Float64ToDurationpb(input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)

	// 测试负数输入
	input = -2.5 // -2.5秒
	expected = durationpb.New(-2500 * time.Millisecond)
	result = Float64ToDurationpb(input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)
}

func TestSecondToDurationpb(t *testing.T) {
	// 测试有效输入
	input := 2.5 // 2.5秒
	expected := durationpb.New(2500 * time.Millisecond)
	result := SecondToDurationpb(&input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)

	// 测试零输入
	input = 0.0
	expected = durationpb.New(0)
	result = SecondToDurationpb(&input)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)

	// 测试空输入
	result = SecondToDurationpb(nil)
	assert.Nil(t, result)
}

func TestDurationpbSecond(t *testing.T) {
	// 测试非空输入
	duration := durationpb.New(5 * time.Second)
	result := DurationpbSecond(duration)
	assert.NotNil(t, result)
	assert.Equal(t, 5.0, *result, "应返回正确的秒数")

	// 测试零输入
	duration = durationpb.New(0)
	result = DurationpbSecond(duration)
	assert.NotNil(t, result)
	assert.Equal(t, 0.0, *result, "应返回零秒")

	// 测试空输入
	result = DurationpbSecond(nil)
	assert.Nil(t, result, "空输入应返回nil")
}
