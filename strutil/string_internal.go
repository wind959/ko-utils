package strutil

import (
	"strings"
	"unicode"
)

func splitIntoStrings(s string, upperCase bool) []string {
	var runes [][]rune
	lastCharType := 0
	charType := 0

	//根据unicode字符的类型拆分为字段
	for _, r := range s {
		switch true {
		case isLower(r):
			charType = 1
		case isUpper(r):
			charType = 2
		case isDigit(r):
			charType = 3
		default:
			charType = 4
		}

		if charType == lastCharType {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastCharType = charType
	}

	for i := 0; i < len(runes)-1; i++ {
		if isUpper(runes[i][0]) && isLower(runes[i+1][0]) {
			length := len(runes[i]) - 1
			temp := runes[i][length]
			runes[i+1] = append([]rune{temp}, runes[i+1]...)
			runes[i] = runes[i][:length]
		}
	}

	//过滤所有非字母和非数字
	var result []string
	for _, rs := range runes {
		if len(rs) > 0 && (unicode.IsLetter(rs[0]) || isDigit(rs[0])) {
			if upperCase {
				result = append(result, string(toUpperAll(rs)))
			} else {
				result = append(result, string(toLowerAll(rs)))
			}
		}
	}

	return result
}

// 判断是否为数字
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// 判断是否为小写
func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// 判断是否为大写
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// 转小写
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

// 转小写
func toLowerAll(rs []rune) []rune {
	for i := range rs {
		rs[i] = toLower(rs[i])
	}
	return rs
}

// 转大写
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

// 转大写
func toUpperAll(rs []rune) []rune {
	for i := range rs {
		rs[i] = toUpper(rs[i])
	}
	return rs
}

// padAtPosition 函数用于在字符串的指定位置填充字符，使其达到指定的长度。
//
// 参数:
//   - str: 需要填充的原始字符串。
//   - length: 填充后字符串的目标长度。
//   - padStr: 用于填充的字符或字符串。如果为空字符串，则默认使用空格填充。
//   - position: 填充位置，0表示居中填充，1表示在字符串右侧填充，2表示在字符串左侧填充。
//
// 返回值:
//   - 返回填充后的字符串。如果原始字符串长度已经大于或等于目标长度，则直接返回原始字符串。
func padAtPosition(str string, length int, padStr string, position int) string {
	// 如果原始字符串长度已经大于或等于目标长度，直接返回原始字符串
	if len(str) >= length {
		return str
	}

	// 如果填充字符串为空，则默认使用空格填充
	if padStr == "" {
		padStr = " "
	}

	// 计算需要填充的总长度
	totalPad := length - len(str)
	startPad := 0

	// 根据填充位置计算左侧填充的长度
	if position == 0 {
		startPad = totalPad / 2 // 居中填充，左侧填充一半
	} else if position == 1 {
		startPad = totalPad // 右侧填充，左侧不填充
	} else if position == 2 {
		startPad = 0 // 左侧填充，右侧不填充
	}
	endPad := totalPad - startPad // 计算右侧填充的长度

	// 定义一个函数，用于生成指定长度的填充字符串
	repeatPad := func(n int) string {
		repeated := strings.Repeat(padStr, (n+len(padStr)-1)/len(padStr))
		return repeated[:n]
	}

	// 生成左侧和右侧的填充字符串
	left := repeatPad(startPad)
	right := repeatPad(endPad)

	// 返回填充后的字符串
	return left + str + right
}

// isLetter  断一个Unicode字符是否为字母，但排除了某些特定范围的字符（如日文、中文等）。
// 首先检查字符是否为字母，如果不是则返回false；
// 如果是，则进一步检查是否在排除的范围内，如果在则返回false，否则返回true
func isLetter(r rune) bool {
	if !unicode.IsLetter(r) {
		return false
	}

	switch {

	case r >= '\u3034' && r < '\u30ff':
		return false

	case r >= '\u3400' && r < '\u4dbf':
		return false

	case r >= '\u4e00' && r < '\u9fff':
		return false

	case r >= '\uf900' && r < '\ufaff':
		return false

	case r >= '\uff66' && r < '\uff9f':
		return false
	}

	return true
}
