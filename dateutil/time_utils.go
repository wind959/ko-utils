package dateutil

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var timeFormat map[string]string

func init() {
	timeFormat = map[string]string{
		"yyyy-mm-dd hh:mm:ss": "2006-01-02 15:04:05",
		"yyyy-mm-dd hh:mm":    "2006-01-02 15:04",
		"yyyy-mm-dd hh":       "2006-01-02 15",
		"yyyy-mm-dd":          "2006-01-02",
		"yyyy-mm":             "2006-01",
		"mm-dd":               "01-02",
		"dd-mm-yy hh:mm:ss":   "02-01-06 15:04:05",
		"yyyy/mm/dd hh:mm:ss": "2006/01/02 15:04:05",
		"yyyy/mm/dd hh:mm":    "2006/01/02 15:04",
		"yyyy/mm/dd hh":       "2006/01/02 15",
		"yyyy/mm/dd":          "2006/01/02",
		"yyyy/mm":             "2006/01",
		"mm/dd":               "01/02",
		"dd/mm/yy hh:mm:ss":   "02/01/06 15:04:05",
		"yyyymmdd":            "20060102",
		"mmddyy":              "010206",
		"yyyy":                "2006",
		"yy":                  "06",
		"mm":                  "01",
		"hh:mm:ss":            "15:04:05",
		"hh:mm":               "15:04",
		"mm:ss":               "04:05",
	}
}

// AddMinute 将日期加/减分钟数
func AddMinute(t time.Time, minutes int64) time.Time {
	return t.Add(time.Minute * time.Duration(minutes))
}

// AddHour 将日期加/减小时数
func AddHour(t time.Time, hours int64) time.Time {
	return t.Add(time.Hour * time.Duration(hours))
}

// AddDay 将日期加/减天数
func AddDay(t time.Time, days int64) time.Time {
	return t.Add(24 * time.Hour * time.Duration(days))
}

// AddWeek 将日期加/减星期数
func AddWeek(t time.Time, weeks int64) time.Time {
	return t.Add(7 * 24 * time.Hour * time.Duration(weeks))
}

// AddMonth 将日期加/减月数
func AddMonth(t time.Time, months int64) time.Time {
	return t.AddDate(0, int(months), 0)
}

// AddYear 将日期加/减年数
func AddYear(t time.Time, year int64) time.Time {
	return t.AddDate(int(year), 0, 0)
}

// AddDaySafe 增加/减少指定的天数，并确保日期是有效日期
func AddDaySafe(t time.Time, days int) time.Time {
	t = t.AddDate(0, 0, days)
	year, month, day := t.Date()

	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()

	if day > lastDayOfMonth {
		t = time.Date(year, month, lastDayOfMonth, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	}

	return t
}

// AddMonthSafe 增加/减少指定的月份，并确保日期是有效日期
func AddMonthSafe(t time.Time, months int) time.Time {
	year := t.Year()
	month := int(t.Month()) + months

	for month > 12 {
		month -= 12
		year++
	}
	for month < 1 {
		month += 12
		year--
	}

	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	day := t.Day()
	if day > daysInMonth {
		day = daysInMonth
	}

	return time.Date(year, time.Month(month), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// AddYearSafe 增加/减少指定的年份，并确保日期是有效日期
func AddYearSafe(t time.Time, years int) time.Time {
	year, month, day := t.Date()
	year += years

	if month == time.February && day == 29 {
		if !IsLeapYear(year) {
			day = 28
		}
	}

	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// GetNowDate 获取当天日期，返回格式：yyyy-mm-dd
func GetNowDate() string {
	return time.Now().Format("2006-01-02")
}

// GetNowTime 获取当时时间，返回格式：hh:mm:ss
func GetNowTime() string {
	return time.Now().Format("15:04:05")
}

// GetNowDateTime 获取当时日期和时间，返回格式：yyyy-mm-dd hh:mm:ss
func GetNowDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetTodayStartTime 返回当天开始时间， 格式: yyyy-mm-dd 00:00:00
func GetTodayStartTime() string {
	return time.Now().Format("2006-01-02") + " 00:00:00"
}

// GetTodayEndTime 返回当天结束时间，格式: yyyy-mm-dd 23:59:59
func GetTodayEndTime() string {
	return time.Now().Format("2006-01-02") + " 23:59:59"
}

// GetZeroHourTimestamp 获取零点时间戳(timestamp of 00:00)
func GetZeroHourTimestamp() int64 {
	ts := time.Now().Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", ts)
	return t.UTC().Unix() - 8*3600
}

// GetNightTimestamp 获取午夜时间戳(timestamp of 23:59)
func GetNightTimestamp() int64 {
	return GetZeroHourTimestamp() + 86400 - 1
}

// FormatTimeToStr 将日期格式化成字符串，`format`
func FormatTimeToStr(t time.Time, format string, timezone ...string) string {
	tf, ok := timeFormat[strings.ToLower(format)]
	if !ok {
		return ""
	}

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return ""
		}
		return t.In(loc).Format(tf)
	}
	return t.Format(tf)
}

// FormatStrToTime 将字符串格式化成时间，`format` 参数格式
func FormatStrToTime(str, format string, timezone ...string) (time.Time, error) {
	tf, ok := timeFormat[strings.ToLower(format)]
	if !ok {
		return time.Time{}, fmt.Errorf("format %s not support", format)
	}

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return time.Time{}, err
		}

		return time.ParseInLocation(tf, str, loc)
	}

	return time.Parse(tf, str)
}

// BeginOfMinute 返回指定时间的分钟开始时间
func BeginOfMinute(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, t.Location())
}

// EndOfMinute 返回指定时间的分钟结束时间
func EndOfMinute(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), t.Minute(), 59, int(time.Second-time.Nanosecond), t.Location())
}

// BeginOfHour 返回指定时间的小时开始时间
func BeginOfHour(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())
}

// EndOfHour 返回指定时间的小时结束时间
func EndOfHour(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// BeginOfDay 返回指定时间的当天开始时间
func BeginOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// EndOfDay 返回指定时间的当天结束时间
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// BeginOfWeek 返回指定时间的每周开始时间,默认开始时间星期日
func BeginOfWeek(t time.Time, beginFrom time.Weekday) time.Time {
	y, m, d := t.AddDate(0, 0, int(beginFrom-t.Weekday())).Date()
	beginOfWeek := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	if beginOfWeek.After(t) {
		return beginOfWeek.AddDate(0, 0, -7)
	}
	return beginOfWeek
}

// EndOfWeek 返回指定时间的星期结束时间,默认结束时间星期六
func EndOfWeek(t time.Time, endWith time.Weekday) time.Time {
	y, m, d := t.AddDate(0, 0, int(endWith-t.Weekday())).Date()
	var endWithWeek = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
	if endWithWeek.Before(t) {
		endWithWeek = endWithWeek.AddDate(0, 0, 7)
	}
	return endWithWeek
}

// BeginOfMonth 返回指定时间的当月开始时间
func BeginOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 返回指定时间的当月结束时间
func EndOfMonth(t time.Time) time.Time {
	return BeginOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// BeginOfYear 返回指定时间的当年开始时间
func BeginOfYear(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear 返回指定时间的当年结束时间
func EndOfYear(t time.Time) time.Time {
	return BeginOfYear(t).AddDate(1, 0, 0).Add(-time.Nanosecond)
}

// IsLeapYear 验证是否是闰年
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// BetweenSeconds 返回两个时间的间隔秒数
func BetweenSeconds(t1 time.Time, t2 time.Time) int64 {
	index := t2.Unix() - t1.Unix()
	return index
}

// DayOfYear 返回参数日期是一年中的第几天
func DayOfYear(t time.Time) int {
	y, m, d := t.Date()
	firstDay := time.Date(y, 1, 1, 0, 0, 0, 0, t.Location())
	nowDate := time.Date(y, m, d, 0, 0, 0, 0, t.Location())

	return int(nowDate.Sub(firstDay).Hours() / 24)
}

// NowDateOrTime 根据指定的格式和时区返回当前时间字符串
func NowDateOrTime(format string, timezone ...string) string {
	tf, ok := timeFormat[strings.ToLower(format)]
	if !ok {
		return ""
	}
	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return ""
		}
		return time.Now().In(loc).Format(tf)
	}
	return time.Now().Format(tf)
}

// Timestamp 返回当前秒级时间戳
func Timestamp(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}

		t = t.In(loc)
	}

	return t.Unix()
}

// TimestampMilli 返回当前毫秒级时间戳
func TimestampMilli(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}
		t = t.In(loc)
	}

	return int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
}

// TimestampMicro 返回当前微秒级时间戳
func TimestampMicro(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}
		t = t.In(loc)
	}

	return int64(time.Nanosecond) * t.UnixNano() / int64(time.Microsecond)
}

// TimestampNano 返回当前纳秒级时间戳
func TimestampNano(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}
		t = t.In(loc)
	}

	return t.UnixNano()
}

// TrackFuncTime 测试函数执行时间
func TrackFuncTime(pre time.Time) func() {
	callerName := getCallerName()
	return func() {
		elapsed := time.Since(pre)
		fmt.Printf("Function %s execution time:\t %v", callerName, elapsed)
	}
}

func getCallerName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "Unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "Unknown"
	}

	fullName := fn.Name()
	if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
		return fullName[lastDot+1:]
	}

	return fullName
}

// DaysBetween 返回两个日期之间的天数差
func DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	days := int(duration.Hours() / 24)

	return days
}

// GenerateDatetimesBetween 生成从start到end的所有日期时间的字符串列表。
// layout参数表示时间格式，例如"2006-01-02 15:04:05"，
// interval参数表示时间间隔，例如"1h"表示1小时，"30m"表示30分钟。
func GenerateDatetimesBetween(start, end time.Time, layout string, interval string) ([]string, error) {
	var result []string

	if start.After(end) {
		start, end = end, start
	}

	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	for current := start; !current.After(end); current = current.Add(duration) {
		result = append(result, current.Format(layout))
	}

	return result, nil
}

// Min 返回最早时间
func Min(t1 time.Time, times ...time.Time) time.Time {
	minTime := t1

	for _, t := range times {
		if t.Before(minTime) {
			minTime = t
		}
	}

	return minTime
}

// Max 返回最晚时间
func Max(t1 time.Time, times ...time.Time) time.Time {
	maxTime := t1

	for _, t := range times {
		if t.After(maxTime) {
			maxTime = t
		}
	}

	return maxTime
}

// MaxMin 返回最早和最晚时间
func MaxMin(t1 time.Time, times ...time.Time) (maxTime time.Time, minTime time.Time) {
	maxTime = t1
	minTime = t1

	for _, t := range times {
		if t.Before(minTime) {
			minTime = t
		}

		if t.After(maxTime) {
			maxTime = t
		}
	}

	return maxTime, minTime
}
