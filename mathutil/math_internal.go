package mathutil

import "golang.org/x/exp/constraints"

func gcd[T constraints.Integer](a, b T) T {
	if b == 0 {
		return a
	}
	return gcd[T](b, a%b)
}

func lcm[T constraints.Integer](a, b T) T {
	if a == 0 || b == 0 {
		panic("lcm function: provide non zero integers only.")
	}
	return a * b / gcd(a, b)
}
