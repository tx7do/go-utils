package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func parseDate(str string) time.Time {
	t, _ := time.Parse(DateLayout, str)
	return t
}

func toSecond(str string) int64 {
	t, _ := time.Parse(DateLayout, str)
	return t.Unix()
}

func TestDifferenceDays(t *testing.T) {
	assert.Equal(t, StringDifferenceDays("2017-09-01", "2017-09-01"), 0)
	assert.Equal(t, StringDifferenceDays("2017-09-01", "2017-09-02"), 1)
	assert.Equal(t, StringDifferenceDays("2017-09-01", "2017-09-03"), 2)
	assert.Equal(t, StringDifferenceDays("2017-09-01", "2017-09-04"), 3)

	assert.Equal(t, StringDifferenceDays("2017-09-01", "2018-03-11"), 191)

	assert.True(t, (StringDifferenceDays("2017-09-01", "2017-09-01")) == 0)
	assert.True(t, (StringDifferenceDays("2017-09-01", "2017-09-02"))%1 == 0)
	assert.True(t, (StringDifferenceDays("2017-09-01", "2017-09-03"))%2 == 0)
}

func TestTimeDifferenceDays(t *testing.T) {
	assert.Equal(t, TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-01")), 0)
	assert.Equal(t, TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-02")), 1)
	assert.Equal(t, TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-03")), 2)
	assert.Equal(t, TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-04")), 3)

	assert.Equal(t, TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2018-03-11")), 191)

	assert.True(t, (TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-01"))) == 0)
	assert.True(t, (TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-02")))%1 == 0)
	assert.True(t, (TimeDifferenceDays(parseDate("2017-09-01"), parseDate("2017-09-03")))%2 == 0)
}

func TestSecondsDifferenceDays(t *testing.T) {
	assert.Equal(t, SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-01")), 0)
	assert.Equal(t, SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-02")), 1)
	assert.Equal(t, SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-03")), 2)
	assert.Equal(t, SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-04")), 3)

	assert.Equal(t, SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2018-03-11")), 191)

	assert.True(t, (SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-01"))) == 0)
	assert.True(t, (SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-02")))%1 == 0)
	assert.True(t, (SecondsDifferenceDays(toSecond("2017-09-01"), toSecond("2017-09-03")))%2 == 0)
}
