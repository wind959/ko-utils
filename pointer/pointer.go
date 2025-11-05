package pointer

import "reflect"

// Of 返回传入参数的指针值
func Of[T any](v T) *T {
	if IsNil(v) {
		return nil
	}
	return &v
}

// Unwrap 返回传入指针指向的值
func Unwrap[T any](p *T) T {
	return *p
}

// UnwrapOrDefault 返回指针的值，如果指针为零值，则返回相应零值
func UnwrapOrDefault[T any](p *T) T {
	var v T

	if p == nil {
		return v
	}
	return *p
}

// UnwrapOr 返回指针的值，如果指针为零值，则返回fallback
func UnwrapOr[T any](p *T, fallback ...T) T {
	if !IsNil(p) {
		return *p
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	var t T
	return t
}

// ExtractPointer 返回传入interface的底层值
func ExtractPointer(value any) any {
	if IsNil(value) {
		return value
	}
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if t.Kind() != reflect.Pointer {
		return value
	}

	if v.Elem().IsValid() {
		return ExtractPointer(v.Elem().Interface())
	}

	return nil
}

func IsNil(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}
