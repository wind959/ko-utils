package objutils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// IsEqual 判断两个对象是否相等
func IsEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// IsAnyEqual 判断任意数量的对象是否相等
func IsAnyEqual(values ...interface{}) bool {
	if len(values) < 2 {
		return true
	}
	for i := 1; i < len(values); i++ {
		if !IsEqual(values[0], values[i]) {
			return false
		}
	}
	return true
}

// IsNil 判断对象是否为nil
func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// IsNotNil 判断对象是否不为nil
func IsNotNil(obj interface{}) bool {
	return !IsNil(obj)
}

// IsNotEmpty 判断对象是否不为空值
func IsNotEmpty(obj interface{}) bool {
	return !IsEmpty(obj)
}

// IsEmpty 判断对象是否为空值
func IsEmpty(obj interface{}) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Struct:
		return v.IsZero()
	}
	return false
}

// GetType 获取对象的类型名称
func GetType(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	return reflect.TypeOf(obj).String()
}

// GetKind 获取对象的种类
func GetKind(obj interface{}) reflect.Kind {
	if obj == nil {
		return reflect.Invalid
	}
	return reflect.TypeOf(obj).Kind()
}

// IsType 判断对象是否为指定类型
func IsType(obj interface{}, typ reflect.Type) bool {
	if obj == nil || typ == nil {
		return obj == nil && typ == nil
	}
	return reflect.TypeOf(obj) == typ
}

// IsKind 判断对象是否为指定种类
func IsKind(obj interface{}, kind reflect.Kind) bool {
	if obj == nil {
		return kind == reflect.Invalid
	}
	return reflect.TypeOf(obj).Kind() == kind
}

// DeepCopy 深拷贝一个对象
func DeepCopy(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}

	srcValue := reflect.ValueOf(src)
	srcType := srcValue.Type()

	// 创建新对象
	dstValue := reflect.New(srcType).Elem()

	// 执行拷贝
	err := deepCopyValue(dstValue, srcValue)
	if err != nil {
		return nil, err
	}

	return dstValue.Interface(), nil
}

// deepCopyValue 递归拷贝值
func deepCopyValue(dst, src reflect.Value) error {
	switch src.Kind() {
	case reflect.Interface:
		if src.IsNil() {
			return nil
		}
		originalValue := src.Elem()
		dst.Set(reflect.New(originalValue.Type()).Elem())
		return deepCopyValue(dst.Elem(), originalValue)
	case reflect.Ptr:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.New(src.Elem().Type()))
		return deepCopyValue(dst.Elem(), src.Elem())
	case reflect.Map:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			err := deepCopyValue(copyValue, originalValue)
			if err != nil {
				return err
			}
			dst.SetMapIndex(key, copyValue)
		}
	case reflect.Slice:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		fallthrough
	case reflect.Array:
		for i := 0; i < src.Len(); i++ {
			err := deepCopyValue(dst.Index(i), src.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if dst.Field(i).CanSet() {
				err := deepCopyValue(dst.Field(i), src.Field(i))
				if err != nil {
					return err
				}
			}
		}
	default:
		dst.Set(src)
	}
	return nil
}

// Contains 判断对象中是否包含元素
// 支持的对象类型包括：string, collection, map, array, slice等
func Contains(obj interface{}, element interface{}) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.String:
		// 字符串类型
		if elemStr, ok := element.(string); ok {
			return strings.Contains(v.String(), elemStr)
		}
		// 如果element不是string类型，尝试转换
		elemStr := ToString(element)
		return strings.Contains(v.String(), elemStr)

	case reflect.Array, reflect.Slice:
		// 数组和切片类型
		for i := 0; i < v.Len(); i++ {
			if IsEqual(v.Index(i).Interface(), element) {
				return true
			}
		}
		return false

	case reflect.Map:
		// 映射类型，检查key是否存在
		mapKey := reflect.ValueOf(element)
		if !mapKey.IsValid() {
			return false
		}
		val := v.MapIndex(mapKey)
		return val.IsValid()

	case reflect.Struct:
		// 结构体类型，检查是否有对应的字段
		for i := 0; i < v.NumField(); i++ {
			if IsEqual(v.Field(i).Interface(), element) {
				return true
			}
		}
		return false

	default:
		// 其他类型不支持contains操作
		return false
	}
}

// ToString 将任意类型转换为字符串
func ToString(obj interface{}) string {
	if obj == nil {
		return ""
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", obj)
	}
}
