package strutil

import "testing"

func TestName(t *testing.T) {

	kebabCase := UpperKebabCase("up-ka")
	t.Log(kebabCase)

	kebabCase1 := UpperKebabCase("up")
	t.Log(kebabCase1)

	lowerkebabCase := KebabCase("UP")
	t.Log(lowerkebabCase)

	reverse := Reverse("up")
	t.Log(reverse)
}
