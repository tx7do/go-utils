package timeutil

import (
	"fmt"
	"testing"
	"time"

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
}

func TestTimeToDateString(t *testing.T) {
	fmt.Println(*TimeToTimeString(trans.Time(time.Now())))
	fmt.Println(*TimeToDateString(trans.Time(time.Now())))
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
