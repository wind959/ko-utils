package strutil

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// used in `Shuffle` function
var rng = rand.New(rand.NewSource(int64(time.Now().UnixNano())))
var (
	// DefaultTrimChars are the characters which are stripped by Trim* functions in default.
	DefaultTrimChars = string([]byte{
		'\t', // Tab.
		'\v', // Vertical tab.
		'\n', // New line (line feed).
		'\r', // Carriage return.
		'\f', // New page.
		' ',  // Ordinary space.
		0x00, // NUL-byte.
		0x85, // Delete.
		0xA0, // Non-breaking space.
	})
)

// IntToString int转string
func IntToString(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

// Int32ToString int32转string
func Int32ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

// Int64ToString int64转string
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// StringToInt string转int
func StringToInt(i string) int {
	j, _ := strconv.Atoi(i)
	return j
}

// StringToInt64 string转int64
func StringToInt64(i string) int64 {
	j, _ := strconv.ParseInt(i, 10, 64)
	return j
}

// StringToInt32 string转int32
func StringToInt32(i string) int32 {
	j, _ := strconv.ParseInt(i, 10, 64)
	return int32(j)
}

// IsContain 判断字符串是否在字符串列表中
func IsContain(target string, List []string) bool {
	for _, element := range List {

		if target == element {
			return true
		}
	}
	return false
}

// StructToJsonString 结构体转json字符串
func StructToJsonString(param interface{}) string {
	dataType, _ := json.Marshal(param)
	dataString := string(dataType)
	return dataString
}

// StructToJsonBytes 结构体转json字节数组
func StructToJsonBytes(param interface{}) []byte {
	dataType, _ := json.Marshal(param)
	return dataType
}

// JsonStringToStruct json字符串转结构体
func JsonStringToStruct(s string, args interface{}) error {
	err := json.Unmarshal([]byte(s), args)
	return err
}

// Shuffle 打乱字符串的顺序
func Shuffle(str string) string {
	runes := []rune(str)

	for i := len(runes) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// CamelCase 驼峰命名
func CamelCase(s string) string {
	var builder strings.Builder

	strs := splitIntoStrings(s, false)
	for i, str := range strs {
		if i == 0 {
			builder.WriteString(strings.ToLower(str))
		} else {
			builder.WriteString(Capitalize(str))
		}
	}
	return builder.String()
}

// Capitalize 将字符串的第一个字符转换为大写，其余字符转换为小写。
func Capitalize(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

// UpperFirst 将首字母转大写
func UpperFirst(s string) string {
	if len(s) == 0 {
		return ""
	}

	r, size := utf8.DecodeRuneInString(s)
	r = unicode.ToUpper(r)

	return string(r) + s[size:]
}

// LowerFirst 将字符串的第一个字符转换为小写
func LowerFirst(s string) string {
	if len(s) == 0 {
		return ""
	}

	r, size := utf8.DecodeRuneInString(s)
	r = unicode.ToLower(r)

	return string(r) + s[size:]
}

// Pad 如果 Pad 比 size 短，则在左侧和右侧填充字符串。
// 如果填充字符超过大小，则填充字符将被截断。
func Pad(source string, size int, padStr string) string {
	return padAtPosition(source, size, padStr, 0)
}

// PadStart 会在字符串长度小于指定大小时在其左侧进行填充。
// 如果填充字符超出指定大小，则会进行截断。
func PadStart(source string, size int, padStr string) string {
	return padAtPosition(source, size, padStr, 1)
}

// PadEnd 函数会在字符串长度小于指定大小时在其右侧进行填充。
// 如果填充字符超出指定大小，则会进行截断。
func PadEnd(source string, size int, padStr string) string {
	return padAtPosition(source, size, padStr, 2)
}

// KebabCase 将字符串转换为连字符分隔的小写形式，非字母和数字字符将被忽略。
func KebabCase(s string) string {
	result := splitIntoStrings(s, false)
	return strings.Join(result, "-")
}

// UpperKebabCase 将字符串转换为大写 KEBAB-CASE 格式，非字母和数字将被忽略。
func UpperKebabCase(s string) string {
	result := splitIntoStrings(s, true)
	return strings.Join(result, "-")
}

// SnakeCase 将字符串转换为 snake_case 格式，非字母和数字部分将被忽略。
func SnakeCase(s string) string {
	result := splitIntoStrings(s, false)
	return strings.Join(result, "_")
}

// UpperSnakeCase 可将字符串转换为大写的 SNAKE_CASE 格式，非字母和数字部分将被忽略。
func UpperSnakeCase(s string) string {
	result := splitIntoStrings(s, true)
	return strings.Join(result, "_")
}

// Reverse 函数将字符串反转。
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// IsString 函数判断给定值是否为字符串。
func IsString(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case string:
		return true
	default:
		return false
	}
}

// Wrap 函数将字符串用给定的字符包装起来。
func Wrap(str string, wrapWith string) string {
	if str == "" || wrapWith == "" {
		return str
	}
	var sb strings.Builder
	sb.WriteString(wrapWith)
	sb.WriteString(str)
	sb.WriteString(wrapWith)

	return sb.String()
}

// Unwrap 函数将字符串从给定的字符包装中移除。
func Unwrap(str string, wrapToken string) string {
	if wrapToken == "" || !strings.HasPrefix(str, wrapToken) || !strings.HasSuffix(str, wrapToken) {
		return str
	}
	if len(str) < 2*len(wrapToken) {
		return str
	}
	return str[len(wrapToken) : len(str)-len(wrapToken)]
}

// SplitEx 可以对给定的字符串进行分割操作，该操作可控制分割结果中是否包含空字符串。
func SplitEx(s, sep string, removeEmptyString bool) []string {
	if sep == "" {
		return []string{}
	}
	n := strings.Count(s, sep) + 1
	a := make([]string, n)
	n--
	i := 0
	sepSave := 0
	ignore := false
	for i < n {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}
		ignore = false
		if removeEmptyString {
			if s[:m+sepSave] == "" {
				ignore = true
			}
		}
		if !ignore {
			a[i] = s[:m+sepSave]
			s = s[m+len(sep):]
			i++
		} else {
			s = s[m+len(sep):]
		}
	}
	var ret []string
	if removeEmptyString {
		if s != "" {
			a[i] = s
			ret = a[:i+1]
		} else {
			ret = a[:i]
		}
	} else {
		a[i] = s
		ret = a[:i+1]
	}
	return ret
}

// Substring 函数用于从字符串中提取子串。
func Substring(s string, offset int, length uint) string {
	rs := []rune(s)
	size := len(rs)
	if offset < 0 {
		offset += size
	}
	if offset < 0 {
		offset = 0
	}
	if offset > size {
		return ""
	}
	end := offset + int(length)
	if end > size {
		end = size
	}
	return strings.ReplaceAll(string(rs[offset:end]), "\x00", "")
}

// SplitWords 函数将字符串拆分成单词，其中每个单词仅包含字母字符。
func SplitWords(s string) []string {
	var word string
	var words []string
	var r rune
	var size, pos int

	isWord := false

	for len(s) > 0 {
		r, size = utf8.DecodeRuneInString(s)

		switch {
		case isLetter(r):
			if !isWord {
				isWord = true
				word = s
				pos = 0
			}

		case isWord && (r == '\'' || r == '-'):
			// is word

		default:
			if isWord {
				isWord = false
				words = append(words, word[:pos])
			}
		}
		pos += size
		s = s[size:]
	}
	if isWord {
		words = append(words, word[:pos])
	}
	return words
}

// StrToBytes 将字符串转换为字节片而不分配内存。
func StrToBytes(str string) (b []byte) {
	return *(*[]byte)(unsafe.Pointer(&str))
}

// BytesToString 将字节切片转换为字符串，而不分配内存。
func BytesToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

// IsBlank 函数判断给定字符串是否为空
func IsBlank(str string) bool {
	if len(str) == 0 {
		return true
	}
	if str == "" {
		return true
	}
	runes := []rune(str)
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// IsNotBlank 函数判断给定字符串是否不为空
func IsNotBlank(str string) bool {
	return !IsBlank(str)
}

// HasPrefixAny 函数判断给定字符串是否以给定前缀之一开头
func HasPrefixAny(str string, prefixes []string) bool {
	if len(str) == 0 || len(prefixes) == 0 {
		return false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	return false
}

// HasSuffixAny 函数判断给定字符串是否以给定后缀之一结尾
func HasSuffixAny(str string, suffixes []string) bool {
	if len(str) == 0 || len(suffixes) == 0 {
		return false
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(str, suffix) {
			return true
		}
	}
	return false
}

// Trim 从字符串的开头和结尾去除空格（或其他字符）。
// 可选参数 'characterMask' 指定其他剥离字符。
func Trim(str string, characterMask ...string) string {
	trimChars := DefaultTrimChars

	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.Trim(str, trimChars)
}

// HideString 函数将字符串中的指定部分隐藏。
// 将范围替换为origin[start: end]。[开始,结束)
func HideString(origin string, start, end int, replaceChar string) string {
	size := len(origin)

	if start > size-1 || start < 0 || end < 0 || start > end {
		return origin
	}
	if end > size {
		end = size
	}
	if replaceChar == "" {
		return origin
	}
	startStr := origin[0:start]
	endStr := origin[end:size]
	replaceSize := end - start
	replaceStr := strings.Repeat(replaceChar, replaceSize)
	return startStr + replaceStr + endStr
}

// ContainsAll 函数判断给定字符串是否包含所有给定子串。
func ContainsAll(str string, substrs []string) bool {
	for _, v := range substrs {
		if !strings.Contains(str, v) {
			return false
		}
	}
	return true
}

// ContainsAny 如果 target string 包含任何一个 substrs，则返回 true。
func ContainsAny(str string, substrs []string) bool {
	for _, v := range substrs {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}

// Rotate 函数将字符串向右旋转指定的位数。
func Rotate(str string, shift int) string {
	if shift == 0 {
		return str
	}
	runes := []rune(str)
	length := len(runes)
	if length == 0 {
		return str
	}
	shift = shift % length
	if shift < 0 {
		shift = length + shift
	}
	var sb strings.Builder
	sb.Grow(length)
	sb.WriteString(string(runes[length-shift:]))
	sb.WriteString(string(runes[:length-shift]))
	return sb.String()
}

// After 返回源字符串中特定字符串首次出现时的位置之后的子字符串
func After(s, char string) string {
	if char == "" {
		return s
	}
	if i := strings.Index(s, char); i >= 0 {
		return s[i+len(char):]
	}
	return s
}

// AfterLast 返回源字符串中指定字符串最后一次出现时的位置之后的子字符串。
func AfterLast(s, char string) string {
	if char == "" {
		return s
	}
	if i := strings.LastIndex(s, char); i >= 0 {
		return s[i+len(char):]
	}
	return s
}

// Before 返回源字符串中指定字符串第一次出现时的位置之前的子字符串。
func Before(s, char string) string {
	if char == "" {
		return s
	}
	if i := strings.Index(s, char); i >= 0 {
		return s[:i]
	}
	return s
}

// BeforeLast 返回源字符串中指定字符串最后一次出现时的位置之前的子字符串
func BeforeLast(s, char string) string {
	if char == "" {
		return s
	}
	if i := strings.LastIndex(s, char); i >= 0 {
		return s[:i]
	}
	return s
}

// IsEqual 判断传入的两个字符串是否相等
func IsEqual(str1, str2 string) bool {
	return str1 == str2
}

// IsEqualAll 判断传入的多个字符串是否相等
func IsEqualAll(strs ...string) bool {
	if len(strs) == 0 {
		return true
	}
	first := strs[0]
	for _, str := range strs[1:] {
		if str != first {
			return false
		}
	}
	return true
}
