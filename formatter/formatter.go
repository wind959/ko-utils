package formatter

import (
	"encoding/json"
	"github.com/wind959/ko-utils/convertor"
	"golang.org/x/exp/constraints"
	"io"
	"strconv"
	"strings"
)

// Comma 用逗号每隔3位分割数字/字符串，支持添加前缀符号。
// 参数value必须是数字或者可以转为数字的字符串, 否则返回空字符串
func Comma[T constraints.Float | constraints.Integer | string](value T, prefixSymbol string) string {
	numString := convertor.ToString(value)
	_, err := strconv.ParseFloat(numString, 64)
	if err != nil {
		return ""
	}
	isNegative := strings.HasPrefix(numString, "-")
	if isNegative {
		numString = numString[1:]
	}
	index := strings.Index(numString, ".")
	if index == -1 {
		index = len(numString)
	}
	for index > 3 {
		index -= 3
		numString = numString[:index] + "," + numString[index:]
	}
	if isNegative {
		numString = "-" + numString
	}
	return prefixSymbol + numString
}

// Pretty 返回pretty JSON字符串
func Pretty(v any) (string, error) {
	out, err := json.MarshalIndent(v, "", "    ")
	return string(out), err
}

// PrettyToWriter Pretty encode数据到writer
func PrettyToWriter(v any, out io.Writer) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")

	if err := enc.Encode(v); err != nil {
		return err
	}

	return nil
}
