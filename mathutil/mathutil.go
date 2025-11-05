package mathutil

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
	"strconv"
	"strings"
)

// Exponent 指数计算（x的n次方）
func Exponent(x, n int64) int64 {
	if n == 0 {
		return 1
	}

	t := Exponent(x, n/2)

	if n%2 == 1 {
		return t * t * x
	}

	return t * t
}

// Fibonacci 计算斐波那契数列的第n个数
func Fibonacci(first, second, n int) int {
	if n <= 0 {
		return 0
	}
	if n < 3 {
		return 1
	} else if n == 3 {
		return first + second
	} else {
		return Fibonacci(second, first+second, n-1)
	}
}

// Factorial 计算阶乘
func Factorial(n uint) uint {
	if n == 0 || n == 1 {
		return 1
	}

	result := uint(1)
	for i := uint(2); i <= n; i++ {
		result *= i
	}

	return result
}

// Percent 计算百分比，保留n位小数
func Percent(val, total float64, n int) float64 {
	if total == 0 {
		return float64(0)
	}
	tmp := val / total * 100
	result := RoundToFloat(tmp, n)

	return result
}

// RoundToString 四舍五入，保留n位小数，返回字符串
func RoundToString[T constraints.Float | constraints.Integer](x T, n int) string {
	tmp := math.Pow(10.0, float64(n))
	x *= T(tmp)
	r := math.Round(float64(x))
	result := strconv.FormatFloat(r/tmp, 'f', n, 64)
	return result
}

// RoundToFloat 四舍五入，保留n位小数
func RoundToFloat[T constraints.Float | constraints.Integer](x T, n int) float64 {
	tmp := math.Pow(10.0, float64(n))
	x *= T(tmp)
	r := math.Round(float64(x))
	return r / tmp
}

// TruncRound 四舍五入 n 位小数
func TruncRound[T constraints.Float | constraints.Integer](x T, n int) T {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n+1)+"f", x)
	temp := strings.Split(floatStr, ".")
	var newFloat string
	if len(temp) < 2 || n >= len(temp[1]) {
		newFloat = floatStr
	} else {
		newFloat = temp[0] + "." + temp[1][:n]
	}
	result, _ := strconv.ParseFloat(newFloat, 64)
	return T(result)
}

// FloorToFloat 向下舍入（去尾法），保留n位小数
func FloorToFloat[T constraints.Float | constraints.Integer](x T, n int) float64 {
	tmp := math.Pow(10.0, float64(n))
	x *= T(tmp)
	r := math.Floor(float64(x))
	return r / tmp
}

// FloorToString 向下舍入（去尾法），保留n位小数，返回字符串

func FloorToString[T constraints.Float | constraints.Integer](x T, n int) string {
	tmp := math.Pow(10.0, float64(n))
	x *= T(tmp)
	r := math.Floor(float64(x))
	result := strconv.FormatFloat(r/tmp, 'f', n, 64)
	return result
}

// CeilToFloat 向上舍入（进一法），保留n位小数
func CeilToFloat[T constraints.Float | constraints.Integer](x T, n int) float64 {
	tmp := math.Pow(10.0, float64(n))
	x *= T(tmp)
	r := math.Ceil(float64(x))
	return r / tmp
}

// CeilToString 向上舍入（进一法），保留n位小数，返回字符串
func CeilToString[T constraints.Float | constraints.Integer](x T, n int) string {
	tmp := math.Pow(10.0, float64(n))
	x *= T(tmp)
	r := math.Ceil(float64(x))
	result := strconv.FormatFloat(r/tmp, 'f', n, 64)
	return result
}

// Max 返回参数中的最大数
func Max[T constraints.Ordered](items ...T) T {
	if len(items) < 1 {
		panic("mathutil.Max: empty list")
	}

	max := items[0]

	for _, v := range items {
		if max < v {
			max = v
		}
	}

	return max
}

// MaxBy 使用给定的比较器函数返回切片的最大值
func MaxBy[T any](slice []T, comparator func(T, T) bool) T {
	var max T

	if len(slice) == 0 {
		return max
	}

	max = slice[0]

	for i := 1; i < len(slice); i++ {
		val := slice[i]

		if comparator(val, max) {
			max = val
		}
	}

	return max
}

// Min 返回参数中的最小数
func Min[T constraints.Ordered](items ...T) T {
	if len(items) < 1 {
		panic("mathutil.min: empty list")
	}
	min := items[0]
	for _, v := range items {
		if min > v {
			min = v
		}
	}
	return min
}

// MinBy 使用给定的比较器函数返回切片的最小值
func MinBy[T any](slice []T, comparator func(T, T) bool) T {
	var t T

	if len(slice) == 0 {
		return t
	}
	t = slice[0]
	for i := 1; i < len(slice); i++ {
		val := slice[i]

		if comparator(val, t) {
			t = val
		}
	}
	return t
}

// Sum 传入参数之和
func Sum[T constraints.Integer | constraints.Float](numbers ...T) T {
	var sum T

	for _, v := range numbers {
		sum += v
	}

	return sum
}

// Average 计算平均数. 可能需要对结果调用RoundToFloat方法四舍五入
func Average[T constraints.Integer | constraints.Float](numbers ...T) float64 {
	var sum float64
	for _, num := range numbers {
		sum += float64(num)
	}
	return sum / float64(len(numbers))
}

// Range 根据指定的起始值和数量，创建一个数字切片
func Range[T constraints.Integer | constraints.Float](start T, count int) []T {
	size := count
	if count < 0 {
		size = -count
	}

	result := make([]T, size)

	for i, j := 0, start; i < size; i, j = i+1, j+1 {
		result[i] = j
	}

	return result
}

// RangeWithStep 根据指定的起始值，结束值，步长，创建一个数字切片
func RangeWithStep[T constraints.Integer | constraints.Float](start, end, step T) []T {
	result := []T{}

	if start >= end || step == 0 {
		return result
	}

	for i := start; i < end; i += step {
		result = append(result, i)
	}

	return result
}

// AngleToRadian 将角度值转为弧度值

func AngleToRadian(angle float64) float64 {
	radian := angle * (math.Pi / 180)
	return radian
}

// RadianToAngle 将弧度值转为角度值

func RadianToAngle(radian float64) float64 {
	angle := radian * (180 / math.Pi)
	return angle
}

// PointDistance 计算两个坐标点的距离
func PointDistance(x1, y1, x2, y2 float64) float64 {
	a := x1 - x2
	b := y1 - y2
	c := math.Pow(a, 2) + math.Pow(b, 2)

	return math.Sqrt(c)
}

// IsPrime 判断质数
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}

	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}

	return true
}

// GCD 计算最大公约数
func GCD[T constraints.Integer](integers ...T) T {
	result := integers[0]
	for k := range integers {
		result = gcd(integers[k], result)

		if result == 1 {
			return 1
		}
	}
	return result
}

// LCM 计算最小公倍数
func LCM[T constraints.Integer](integers ...T) T {
	result := integers[0]
	for k := range integers {
		result = lcm(integers[k], result)
	}
	return result
}

// Cos 计算弧度的余弦值
func Cos(radian float64, precision ...int) float64 {
	t := 1.0 / (2.0 * math.Pi)
	radian *= t
	radian -= 0.25 + math.Floor(radian+0.25)
	radian *= 16.0 * (math.Abs(radian) - 0.5)
	radian += 0.225 * radian * (math.Abs(radian) - 1.0)
	if len(precision) == 1 {
		return TruncRound(radian, precision[0])
	}
	return TruncRound(radian, 3)
}

// Sin 计算弧度的正弦值
func Sin(radian float64, precision ...int) float64 {
	return Cos((math.Pi/2)-radian, precision...)
}

// Log 计算以base为底n的对数
func Log(n, base float64) float64 {
	return math.Log(n) / math.Log(base)
}

// Abs 求绝对值
func Abs[T constraints.Integer | constraints.Float](x T) T {
	if x < 0 {
		return (-x)
	}

	return x
}

// Div 除法运算
func Div[T constraints.Float | constraints.Integer](x T, y T) float64 {
	return float64(x) / float64(y)
}

// Variance 计算方差
func Variance[T constraints.Float | constraints.Integer](numbers []T) float64 {
	n := len(numbers)
	if n == 0 {
		return 0
	}
	avg := Average(numbers...)
	var sum float64
	for _, v := range numbers {
		sum += (float64(v) - avg) * (float64(v) - avg)
	}
	return sum / float64(n)
}

// StdDev 计算标准差
func StdDev[T constraints.Float | constraints.Integer](numbers []T) float64 {
	return math.Sqrt(Variance(numbers))
}

// Permutation 计算排列数P(n, k)
func Permutation(n, k uint) uint {
	if n < k {
		return 0
	}
	nFactorial := Factorial(n)
	nMinusKFactorial := Factorial(n - k)
	return nFactorial / nMinusKFactorial
}

// Combination 计算组合数C(n, k)
func Combination(n, k uint) uint {
	if n < k {
		return 0
	}
	nFactorial := Factorial(n)
	kFactorial := Factorial(k)
	nMinusKFactorial := Factorial(n - k)
	return nFactorial / (kFactorial * nMinusKFactorial)
}
