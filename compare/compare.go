package compare

import "github.com/wind959/ko-utils/convertor"

// operator type
const (
	equal          = "eq"
	lessThan       = "lt"
	greaterThan    = "gt"
	lessOrEqual    = "le"
	greaterOrEqual = "ge"
)

// Equal 检查两个值是否相等(检查类型和值)
func Equal(left, right any) bool {
	return compareValue(equal, left, right)
}

// EqualValue 检查两个值是否相等(只检查值)
func EqualValue(left, right any) bool {
	ls, rs := convertor.ToString(left), convertor.ToString(right)
	return ls == rs
}

// LessThan 验证参数`left`的值是否小于参数`right`的值。
func LessThan(left, right any) bool {
	return compareValue(lessThan, left, right)
}

// LessOrEqual 验证参数`left`的值是否小于或等于参数`right`的值。
func LessOrEqual(left, right any) bool {
	return compareValue(lessOrEqual, left, right)
}

// GreaterThan 验证参数`left`的值是否大于参数`right`的值。
func GreaterThan(left, right any) bool {
	return compareValue(greaterThan, left, right)
}

// GreaterOrEqual 验证参数`left`的值是否大于或参数`right`的值。
func GreaterOrEqual(left, right any) bool {
	return compareValue(greaterOrEqual, left, right)
}
