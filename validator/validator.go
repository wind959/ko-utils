package validator

import (
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	alphaMatcher           *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z]+$`)
	letterRegexMatcher     *regexp.Regexp = regexp.MustCompile(`[a-zA-Z]`)
	alphaNumericMatcher    *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	numberRegexMatcher     *regexp.Regexp = regexp.MustCompile(`\d`)
	intStrMatcher          *regexp.Regexp = regexp.MustCompile(`^[\+-]?\d+$`)
	urlMatcher             *regexp.Regexp = regexp.MustCompile(`^((ftp|http|https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(([a-zA-Z0-9]+([-\.][a-zA-Z0-9]+)*)|((www\.)?))?(([a-z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-z\x{00a1}-\x{ffff}]{2,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`)
	dnsMatcher             *regexp.Regexp = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	emailMatcher           *regexp.Regexp = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	chineseMobileMatcher   *regexp.Regexp = regexp.MustCompile(`^1(?:3\d|4[4-9]|5[0-35-9]|6[67]|7[013-8]|8\d|9\d)\d{8}$`)
	chineseIdMatcher       *regexp.Regexp = regexp.MustCompile(`^(\d{17})([0-9]|X|x)$`)
	chineseMatcher         *regexp.Regexp = regexp.MustCompile("[\u4e00-\u9fa5]")
	chinesePhoneMatcher    *regexp.Regexp = regexp.MustCompile(`\d{3}-\d{8}|\d{4}-\d{7}|\d{4}-\d{8}`)
	creditCardMatcher      *regexp.Regexp = regexp.MustCompile(`^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|(222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11}|6[27][0-9]{14})$`)
	base64Matcher          *regexp.Regexp = regexp.MustCompile(`^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$`)
	base64URLMatcher       *regexp.Regexp = regexp.MustCompile(`^([A-Za-z0-9_-]{4})*([A-Za-z0-9_-]{2}(==)?|[A-Za-z0-9_-]{3}=?)?$`)
	binMatcher             *regexp.Regexp = regexp.MustCompile(`^(0b)?[01]+$`)
	hexMatcher             *regexp.Regexp = regexp.MustCompile(`^(#|0x|0X)?[0-9a-fA-F]+$`)
	visaMatcher            *regexp.Regexp = regexp.MustCompile(`^4[0-9]{12}(?:[0-9]{3})?$`)
	masterCardMatcher      *regexp.Regexp = regexp.MustCompile(`^5[1-5][0-9]{14}$`)
	americanExpressMatcher *regexp.Regexp = regexp.MustCompile(`^3[47][0-9]{13}$`)
	unionPay               *regexp.Regexp = regexp.MustCompile("^62[0-5]\\d{13,16}$")
	chinaUnionPay          *regexp.Regexp = regexp.MustCompile(`^62[0-9]{14,17}$`)
)

var (
	factor         = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	verifyStr      = [11]string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
	birthStartYear = 1900
	provinceKv     = map[string]struct{}{
		"11": {},
		"12": {},
		"13": {},
		"14": {},
		"15": {},
		"21": {},
		"22": {},
		"23": {},
		"31": {},
		"32": {},
		"33": {},
		"34": {},
		"35": {},
		"36": {},
		"37": {},
		"41": {},
		"42": {},
		"43": {},
		"44": {},
		"45": {},
		"46": {},
		"50": {},
		"51": {},
		"52": {},
		"53": {},
		"54": {},
		"61": {},
		"62": {},
		"63": {},
		"64": {},
		"65": {},
		//"71": {},
		//"81": {},
		//"82": {},
	}
)

// IsAlpha 验证字符串是否只包含英文字母
func IsAlpha(str string) bool {
	return alphaMatcher.MatchString(str)
}

// IsAllUpper 验证字符串是否全是大写英文字母
func IsAllUpper(str string) bool {
	for _, r := range str {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return str != ""
}

// IsAllLower 验证字符串是否全是小写英文字母
func IsAllLower(str string) bool {
	for _, r := range str {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return str != ""
}

// IsASCII 验证字符串是否只包含ASCII字符
func IsASCII(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// IsPrintable 检查字符串是否全部为可打印字符
func IsPrintable(str string) bool {
	for _, r := range str {
		if !unicode.IsPrint(r) {
			if r == '\n' || r == '\r' || r == '\t' || r == '`' {
				continue
			}
			return false
		}
	}
	return true
}

// ContainUpper 验证字符串是否包含至少一个英文大写字母。
func ContainUpper(str string) bool {
	for _, r := range str {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// ContainLower 验证字符串是否包含至少一个英文小写字母
func ContainLower(str string) bool {
	for _, r := range str {
		if unicode.IsLower(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// ContainLetter 验证字符串是否包含至少一个英文字母
func ContainLetter(str string) bool {
	return letterRegexMatcher.MatchString(str)
}

// IsJSON 验证字符串是否是有效json
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsNumberStr 验证字符串是否是可以转换为数字
func IsNumberStr(s string) bool {
	return IsIntStr(s) || IsFloatStr(s)
}

// IsFloatStr 验证字符串是否是可以转换为浮点数
func IsFloatStr(str string) bool {
	_, e := strconv.ParseFloat(str, 64)
	return e == nil
}

// IsIntStr 验证字符串是否是可以转换为整数
func IsIntStr(str string) bool {
	return intStrMatcher.MatchString(str)
}

// IsIp 验证字符串是否是ip地址
func IsIp(ipstr string) bool {
	ip := net.ParseIP(ipstr)
	return ip != nil
}

// IsIpPort 检查字符串是否是ip:port格式
func IsIpPort(str string) bool {
	host, port, err := net.SplitHostPort(str)
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	return ip != nil && IsPort(port)
}

// IsIpV4 验证字符串是否是ipv4地址
func IsIpV4(ipstr string) bool {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

// IsIpV6 验证字符串是否是ipv6地址
func IsIpV6(ipstr string) bool {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return false
	}
	return ip.To4() == nil && len(ip) == net.IPv6len
}

// IsPort 检查字符串是否是ip:port格式
func IsPort(str string) bool {
	if i, err := strconv.ParseInt(str, 10, 64); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsUrl 验证字符串是否是url
func IsUrl(str string) bool {
	if str == "" || len(str) >= 2083 || len(str) <= 3 || strings.HasPrefix(str, ".") {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return urlMatcher.MatchString(str)
}

// IsDns 验证字符串是否是有效dns
func IsDns(dns string) bool {
	return dnsMatcher.MatchString(dns)
}

// IsEmail 验证字符串是否是有效电子邮件地址
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsChineseMobile 验证字符串是否是中国手机号码
func IsChineseMobile(mobileNum string) bool {
	return chineseMobileMatcher.MatchString(mobileNum)
}

// IsChineseIdNum 验证字符串是否是中国身份证号码
func IsChineseIdNum(id string) bool {
	// All characters should be numbers, and the last digit can be either x or X
	if !chineseIdMatcher.MatchString(id) {
		return false
	}

	_, ok := provinceKv[id[0:2]]
	if !ok {
		return false
	}
	birthStr := fmt.Sprintf("%s-%s-%s", id[6:10], id[10:12], id[12:14])
	birthday, err := time.Parse("2006-01-02", birthStr)
	if err != nil || birthday.After(time.Now()) || birthday.Year() < birthStartYear {
		return false
	}
	sum := 0
	for i, c := range id[:17] {
		v, _ := strconv.Atoi(string(c))
		sum += v * factor[i]
	}

	return verifyStr[sum%11] == strings.ToUpper(id[17:18])
}

// ContainChinese 验证字符串是否包含中文字符
func ContainChinese(s string) bool {
	return chineseMatcher.MatchString(s)
}

// IsChinesePhone 验证字符串是否是中国电话座机号码
func IsChinesePhone(phone string) bool {
	return chinesePhoneMatcher.MatchString(phone)
}

// IsCreditCard 验证字符串是否是信用卡号码
func IsCreditCard(creditCart string) bool {
	return creditCardMatcher.MatchString(creditCart)
}

// IsBase64 验证字符串是否是base64编码
func IsBase64(base64 string) bool {
	return base64Matcher.MatchString(base64)
}

// IsEmptyString 验证字符串是否是空字符串
func IsEmptyString(str string) bool {
	return len(str) == 0
}

// IsRegexMatch 验证字符串是否可以匹配正则表达式
func IsRegexMatch(str, regex string) bool {
	reg := regexp.MustCompile(regex)
	return reg.MatchString(str)
}

// IsStrongPassword 验证字符串是否是强密码：(alpha(lower+upper) + number + special chars(!@#$%^&*()?><))。
func IsStrongPassword(password string, length int) bool {
	if len(password) < length {
		return false
	}
	var num, lower, upper, special bool
	for _, r := range password {
		switch {
		case unicode.IsDigit(r):
			num = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsSymbol(r), unicode.IsPunct(r):
			special = true
		}
	}

	return num && lower && upper && special
}

// IsWeakPassword 验证字符串是否是弱密码：（only letter or only number or letter + number）
func IsWeakPassword(password string) bool {
	var num, letter, special bool
	for _, r := range password {
		switch {
		case unicode.IsDigit(r):
			num = true
		case unicode.IsLetter(r):
			letter = true
		case unicode.IsSymbol(r), unicode.IsPunct(r):
			special = true
		}
	}

	return (num || letter) && !special
}

// IsZeroValue 判断传入的参数值是否为零值
func IsZeroValue(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice, reflect.Map:
		return rv.IsNil()
	}

	return reflect.DeepEqual(rv.Interface(), reflect.Zero(rv.Type()).Interface())
}

// IsGBK 检查数据编码是否为gbk（汉字内部代码扩展规范）。该函数的实现取决于双字节是否在gbk的编码范围内，而utf-8编码格式的每个字节都在gbk编码范围内。
// 因此，应该首先调用utf8.valid检查它是否是utf-8编码，然后调用IsGBK检查gbk编码
func IsGBK(data []byte) bool {
	i := 0
	for i < len(data) {
		if data[i] <= 0xff {
			i++
			continue
		} else {
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}

	return true
}

// IsNumber 验证参数是否是数字(integer or float)
func IsNumber(v any) bool {
	return IsInt(v) || IsFloat(v)
}

// IsFloat 验证参数是否是浮点数(float32, float34)
func IsFloat(v any) bool {
	switch v.(type) {
	case float32, float64:
		return true
	}
	return false
}

// IsInt 验证参数是否是整数(int, unit)
func IsInt(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return true
	}
	return false
}

// IsBin 检查字符串是否是有效的二进制数
func IsBin(v string) bool {
	return binMatcher.MatchString(v)
}

// IsHex 检查字符串是否是有效的十六进制数
func IsHex(v string) bool {
	return hexMatcher.MatchString(v)
}

// IsBase64URL 检查字符串是否是有效的base64 url
func IsBase64URL(v string) bool {
	return base64URLMatcher.MatchString(v)
}

// IsJWT 检查字符串是否是有效的JSON Web Token (JWT)
func IsJWT(v string) bool {
	strings := strings.Split(v, ".")
	if len(strings) != 3 {
		return false
	}

	for _, s := range strings {
		if !IsBase64URL(s) {
			return false
		}
	}

	return true
}

// IsVisa 检查字符串是否是有效的visa卡号
func IsVisa(v string) bool {
	return visaMatcher.MatchString(v)
}

// IsMasterCard 检查字符串是否是有效的mastercard卡号
func IsMasterCard(v string) bool {
	return masterCardMatcher.MatchString(v)
}

// IsAmericanExpress 检查字符串是否是有效的american express卡号
func IsAmericanExpress(v string) bool {
	return americanExpressMatcher.MatchString(v)
}

// IsUnionPay 检查字符串是否是有效的美国银联卡号
func IsUnionPay(v string) bool {
	return unionPay.MatchString(v)
}

// IsChinaUnionPay 检查字符串是否是有效的中国银联卡号
func IsChinaUnionPay(v string) bool {
	return chinaUnionPay.MatchString(v)
}
