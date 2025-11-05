package dateutil

import "time"

type theTime struct {
	unix int64
}

// NewUnixNow 创建一个当前时间的unix时间戳
func NewUnixNow() *theTime {
	return &theTime{unix: time.Now().Unix()}
}

// NewUnix 创建一个unix时间戳
func NewUnix(unix int64) *theTime {
	return &theTime{unix: unix}
}

// NewFormat 创建一个yyyy-mm-dd hh:mm:ss格式时间字符串的unix时间戳
func NewFormat(t string) (*theTime, error) {
	timeLayout := "2006-01-02 15:04:05"
	loc := time.FixedZone("CST", 8*3600)
	tt, err := time.ParseInLocation(timeLayout, t, loc)
	if err != nil {
		return nil, err
	}
	return &theTime{unix: tt.Unix()}, nil
}

// NewISO8601 创建一个iso8601格式时间字符串的unix时间戳
func NewISO8601(iso8601 string) (*theTime, error) {
	t, err := time.ParseInLocation(time.RFC3339, iso8601, time.UTC)
	if err != nil {
		return nil, err
	}
	return &theTime{unix: t.Unix()}, nil
}

// ToUnix 返回unix时间戳
func (t *theTime) ToUnix() int64 {
	return t.unix
}

// ToFormat 返回格式'yyyy-mm-dd hh:mm:ss'的日期字符串
func (t *theTime) ToFormat() string {
	return time.Unix(t.unix, 0).Format("2006-01-02 15:04:05")
}

// ToFormatForTpl 返回tpl格式指定的日期字符串
func (t *theTime) ToFormatForTpl(tpl string) string {
	return time.Unix(t.unix, 0).Format(tpl)
}

// ToIso8601 返回iso8601日期字符串
func (t *theTime) ToIso8601() string {
	return time.Unix(t.unix, 0).Format(time.RFC3339)
}
