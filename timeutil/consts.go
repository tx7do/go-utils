package timeutil

import "time"

const (
	DateLayout  = "2006-01-02"
	ClockLayout = "15:04:05"
	TimeLayout  = DateLayout + " " + ClockLayout

	DefaultTimeLocationName = "Asia/Shanghai"
)

// More predefined layouts for use in Time.Format and time.Parse.
const (
	DT14                = "20060102150405"
	DT8                 = "20060102"
	DT8MDY              = "01022006"
	DT6                 = "200601"
	MonthDay            = "1/2"
	DIN5008FullDate     = "02.01.2006" // German DIN 5008 standard
	DIN5008Date         = "02.01.06"
	RFC3339FullDate     = time.DateOnly
	RFC3339Milli        = "2006-01-02T15:04:05.999Z07:00"
	RFC3339Dash         = "2006-01-02T15-04-05Z07-00"
	ISO8601             = "2006-01-02T15:04:05Z0700"
	ISO8601TZHour       = "2006-01-02T15:04:05Z07"
	ISO8601NoTZ         = "2006-01-02T15:04:05"
	ISO8601MilliNoTZ    = "2006-01-02T15:04:05.999"
	ISO8601Milli        = "2006-01-02T15:04:05.999Z0700"
	ISO8601CompactZ     = "20060102T150405Z0700"
	ISO8601CompactNoTZ  = "20060102T150405"
	ISO8601YM           = "2006-01"
	ISO9075             = time.DateTime                    // ISO/IEC 9075 used by MySQL, BigQuery, etc.
	ISO9075MicroTZ      = "2006-01-02 15:04:05.999999-07"  // ISO/IEC 9075 used by PostgreSQL
	RFC5322             = "Mon, 2 Jan 2006 15:04:05 -0700" // RFC5322             = "Mon Jan 02 15:04:05 -0700 2006"
	SQLTimestamp        = ISO9075
	SQLTimestampMinutes = "2006-01-02 15:04"
	Ruby                = "2006-01-02 15:04:05 -0700" // Ruby Time.now.to_s
	InsightlyAPIQuery   = "_1/_2/2006 _3:04:05 PM"
	DateMDY             = "1/2/2006" // an underscore results in a space.
	DateMDYSlash        = "01/02/2006"
	DateDMYDash         = "_2-01-2006"     // Jira XML Date format
	DateDMYHM2          = "02:01:06 15:04" // GMT time in format dd:mm:yy hh:mm
	DateYMD             = RFC3339FullDate
	DateTextUS          = "January 2, 2006"
	DateTextUSAbbr3     = "Jan 2, 2006"
	DateTextEU          = "2 January 2006"
	DateTextEUAbbr3     = "2 Jan 2006"
	MonthAbbrYear       = "Jan 2006"
	MonthYear           = "January 2006"
)

const (
	RFC3339Min         = "0000-01-01T00:00:00Z"
	RFC3339Max         = "9999-12-31T23:59:59Z"
	RFC3339Zero        = "0001-01-01T00:00:00Z" // Golang zero value
	RFC3339ZeroUnix    = "1970-01-01T00:00:00Z"
	RFC3339YMDZeroUnix = int64(-62135596800)
)

var ReferenceTimeValue time.Time = time.Date(2006, 1, 2, 15, 4, 5, 999999999, time.FixedZone("MST", -7*60*60))
