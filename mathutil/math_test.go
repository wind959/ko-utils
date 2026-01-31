package mathutil

import "testing"

func TestExponent(t *testing.T) {
	amount := 3.1415926
	float := TruncateToFloat(amount, 3)
	t.Log(float)
	toString := TruncateToString(amount, 3)
	t.Log(toString)
}
