package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/kofoworola/godate"
)

// conversion from https://www.php.net/manual/en/function.date.php to https://golang.org/src/time/format.go
var TimeFormatMap = map[string]string{
	// Year
	// "L":"0", // not supported in go. 1 if it is a leap year, 0 otherwise.  Whether it's a leap year
	// "o":"2006", // not supported in go. Examples: 1999 or 2003  ISO-8601 week-numbering year. This has the same value as Y, except that if the ISO week number (W) belongs to the previous or next year, that year is used instead. (added in PHP 5.1.0)
	"y": "06",   // Examples: 99 or 03  A two digit representation of a year
	"Y": "2006", // Examples: 1999 or 2003  A full numeric representation of a year, 4 digits
	// Month
	"n": "1",  // 1 through 12  Numeric representation of a month, without leading zeros
	"m": "01", // 01 through 12 Numeric representation of a month, with leading zeros
	// Day
	"j": "2",  // 1 to 31 Day of the month without leading zeros
	"d": "02", // 01 to 31 Day of the month, 2 digits with leading zeros
	// "N":"1", // not supported in go. 1 (for Monday) through 7 (for Sunday) ISO-8601 numeric representation of the day of the week (added in PHP 5.1.0)
	// "S":"nd", // not supported in go. st, nd, rd or th. Works well with j English ordinal suffix for the day of the month, 2 characters
	// "w":"1", // not supported in go. 0 (for Sunday) through 6 (for Saturday) Numeric representation of the day of the week
	// "z":"1", // not supported in go. 0 through 365 The day of the year (starting from 0)
	// "t":"31", // not supported in go. 28 through 31 Number of days in the given month
	// Week
	// "W":"01", // not supported in go. Example: 42 (the 42nd week in the year) ISO-8601 week number of year, weeks starting on Monday
	// Time
	// "B":"961", // not supported in go. 000 through 999 Swatch Internet time
	"g": "3", // 1 through 12  12-hour format of an hour without leading zeros
	// "G":"15", // not supported in go. 0 through 23  24-hour format of an hour without leading zeros
	"h": "03",      // 01 through 12 12-hour format of an hour with leading zeros
	"H": "15",      // 00 through 23 24-hour format of an hour with leading zeros
	"i": "04",      // 00 to 59  Minutes with leading zeros
	"s": "05",      // 00 through 59 Seconds with leading zeros
	"u": ".000000", // Example: 654321 Microseconds (added in PHP 5.2.2). Note that date() will always generate 000000 since it takes an integer parameter, whereas DateTime::format() does support microseconds if DateTime was created with microseconds.
	"v": ".000",    // Example: 654  Milliseconds (added in PHP 7.0.0). Same note applies as for u.
	// Timezone
	"e": "MST",    // Examples: UTC, GMT, Atlantic/Azores Timezone identifier (added in PHP 5.1.0)
	"O": "-0700",  // Example: +0200  Difference to Greenwich time (GMT) in hours
	"P": "-07:00", // Example: +02:00 Difference to Greenwich time (GMT) with colon between hours and minutes (added in PHP 5.1.3)
	// "T":"MST", // Examples: EST, MDT ...  Timezone abbreviation
	// "I":"0", // not supported in go. (capital i) //  1 if Daylight Saving Time, 0 otherwise. Whether or not the date is in daylight saving time
	// "Z":"-25200", // not supported in go. -43200 through 50400  Timezone offset in seconds. The offset for timezones west of UTC is always negative, and for those east of UTC is always positive.
	// Full  Date/Time
	"U":            "1136239445",                      // See also time() Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)
	"c":            "2006-01-02T15:04:05-07:00",       // 2004-02-12T15:19:21+00:00 ISO 8601 date (added in PHP 5)
	"r":            "Mon, 02 Jan 2006 15:04:05 -0700", // Example: Thu, 21 Dec 2000 16:01:07 +0200 > RFC 2822 formatted date
	"RFC3339":      "2006-01-02T15:04:05Z07:00",
	"RFC3339Micro": "2006-01-02T15:04:05.999Z07:00",
	"RFC3339Milli": "2006-01-02T15:04:05.999999Z07:00",
	"RFC3339Nano":  "2006-01-02T15:04:05.999999999Z07:00",
	"a":            "pm",      // am or pm  Lowercase Ante meridiem and Post meridiem
	"A":            "PM",      // AM or PM  Uppercase Ante meridiem and Post meridiem
	"M":            "Jan",     // Jan through Dec A short textual representation of a month, three letters
	"F":            "January", // January through December  A full textual representation of a month, such as January or March
	"D":            "Mon",     // Mon through Sun A textual representation of a day, three letters
	"l":            "Monday",  //(lowercase 'L') //  Sunday through Saturday A full textual representation of the day of the week
	// Summary
	// Year : 06 | 2006
	// Month : 01 | 1 | Jan | January
	// Day : 02 | 2 | _2   (width two, right justified)
	// Weekday : Mon | Monday
	// Hours : 03 | 3 | 15
	// Minutes : 04 | 4
	// Seconds : 05 | 5
	// ms : μs | ns | .000 | .000000 | .000000000
	// ms : μs | ns | .999 | .999999 | .999999999   (trailing zeros removed)
	// am/pm : PM | pm
	// Timezone : MST
	// Offset : 0700 | -07 | -07:00 | Z0700 | Z07:00
}

func TimeFormatPhpToGo(format string) string {
	goFormat := ""
	for k, v := range TimeFormatMap {
		if goFormat == "" {
			goFormat = strings.ReplaceAll(format, k, v)
		} else {
			goFormat = strings.ReplaceAll(goFormat, k, v)
		}
	}
	return goFormat
}

func GetCurrentTime(format string) string {
	if format == "" {
		format = "Y-m-d H:i:s"
	}
	return time.Now().Format(TimeFormatPhpToGo(format))
}

func ParsingDate(src string, formatFrom string, formatTo string) string {
	date, err := godate.Parse(TimeFormatPhpToGo(formatFrom), src)
	if formatTo == "" {
		formatTo = formatFrom
	}
	if err == nil {
		return date.Format(TimeFormatPhpToGo(formatTo))
	} else {
		return ""
	}
}

func SubDateTime(unit string, src string, n int, format string) string {
	return AddDateTime(unit, src, n*-1, format)
}

func AddDateTime(unit string, src string, n int, format string) string {
	var u godate.Unit
	switch unit {
	case "day":
		u = godate.DAY
	case "hour":
		u = godate.HOUR
	case "minute":
		u = godate.MINUTE
	case "second":
		u = godate.SECOND
	}
	if src == "now" {
		date := godate.Now(time.UTC)
		date = date.Add(n, u)
		return date.Format(TimeFormatPhpToGo(format))
	} else {
		date, err := godate.Parse(TimeFormatPhpToGo(format), src)
		if err == nil {
			date = date.Add(n, u)
		}
		return date.Format(TimeFormatPhpToGo(format))
	}
}

func HaveArrayDate(params map[string][]string) bool {
	var ret bool = false
	for i, _ := range params {
		if strings.Index(i, "data.$") > -1 {
			ret = true
		}
	}
	return ret
}

func SetFilterDate(params map[string][]string, tipe string) {
	if _, ok := params["date"]; ok {
		if !HaveArrayDate(params) {
			if tipe == "range" {
				params["date.$gte"] = params["date"]
				params["date.$lte"] = params["date"]
			} else {
				params["date.$lte"] = params["date"]
			}
			delete(params, "date")
		}
	}
	if _, ok := params["date.$lte"]; !ok {
		if gte, ok := params["date.$gte"]; ok {
			params["date.$lte"] = gte
		} else {
			lte := []string{}
			lte = append(lte, GetCurrentTime("Y-m-d"))
			params["date.$lte"] = lte
		}
	}
	if tipe == "range" {
		if _, ok := params["date.$gte"]; !ok {
			gte := []string{}
			gte = append(gte, ParsingDate(params["date.$lte"][0], "Y-m-d", "Y-m")+"-01")
			params["date.$gte"] = gte
		}
	}
}

func IsEndOfYearDate(date string) bool {
	var b bool = false
	var daymonth string = ParsingDate(date, "Y-m-d", "dm")
	if daymonth == "3112" {
		b = true
	}
	return b
}

func IsStartOfYearDate(date string) bool {
	var b bool = false
	var daymonth string = ParsingDate(date, "Y-m-d", "dm")
	if daymonth == "0101" {
		b = true
	}
	return b
}

func GetEndOfMonth(date string) (string, error) {
	var endmonths []string = []string{"3101", "3103", "3004", "3105", "3006", "3107", "3108", "3009", "3110", "3011", "3112"}
	YearMonth := ParsingDate(date, "Y-m-d", "Y-m")
	Y, err := strconv.Atoi(YearMonth[0:4])
	if err != nil {
		return "", err
	}
	if YearMonth[2:] == "02" {
		if Y%4 == 0 {
			return YearMonth + "-29", nil
		} else {
			return YearMonth + "-28", nil
		}
	} else {
		for _, m := range endmonths {
			if m[2:] == YearMonth[5:7] {
				return YearMonth + "-" + m[0:2], nil
			}
		}
	}
	return "", nil
}

func IsEndOfMonth(date string) bool {
	var daymonth string = ParsingDate(date, "Y-m-d", "dm")
	var b bool = false
	var endmonths []string = []string{"3101", "3103", "3004", "3105", "3006", "3107", "3108", "3009", "3110", "3011", "3112"}
	Y, err := strconv.Atoi(ParsingDate(date, "Y-m-d", "Y"))
	if err != nil {
		return b
	}
	if daymonth[2:] == "02" {
		if Y%4 == 0 && daymonth == "2902" {
			b = true
		} else if daymonth == "2802" {
			b = true
		}
	} else {
		for _, m := range endmonths {
			if m == daymonth {
				b = true
			}
		}
	}
	return b
}

func GetPeriodFromDate(date string, dateStart string) string {
	var period string = "daily"
	if date == "" {
		return period
	}
	if dateStart == "" {
		if IsEndOfYearDate(date) {
			period = "yearly"
		} else if IsEndOfMonth(date) {
			period = "monthly"
		}
	} else {
		if IsStartOfYearDate(dateStart) && IsEndOfYearDate(date) {
			period = "yearly"
		} else if ParsingDate(dateStart, "Y-m-d", "d") == "01" && IsEndOfMonth(date) {
			period = "monthly"
		}
	}
	return period
}
