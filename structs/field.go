package structs

import (
	"github.com/wind959/ko-utils/pointer"
	"reflect"
)

type Field struct {
	Struct
	field reflect.StructField
	tag   *Tag
}

func newField(v reflect.Value, f reflect.StructField, tagName string) *Field {
	tag := f.Tag.Get(tagName)
	field := &Field{
		field: f,
		tag:   newTag(tag),
	}
	field.rvalue = v
	field.rtype = v.Type()
	field.TagName = tagName
	return field
}

// Tag 获取`Field`的`Tag`，默认的tag key是json
func (f *Field) Tag() *Tag {
	return f.tag
}

// Value 获取`Field`属性的值
func (f *Field) Value() any {
	return f.rvalue.Interface()
}

// IsEmbedded 判断属性是否为嵌入
func (f *Field) IsEmbedded() bool {
	return len(f.field.Index) > 1
}

// IsExported 判断属性是否导出
func (f *Field) IsExported() bool {
	return f.field.IsExported()
}

// IsZero 判断属性是否为零值
func (f *Field) IsZero() bool {
	z := reflect.Zero(f.rvalue.Type()).Interface()
	v := f.Value()
	return reflect.DeepEqual(z, v)
}

// IsNil 如果给定的字段值为 nil，则返回 true 。
func (f *Field) IsNil() bool {
	v := f.Value()
	if v == nil || (reflect.ValueOf(v)).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil() {
		return true
	}

	return false
}

// Name 获取属性名
func (f *Field) Name() string {
	return f.field.Name
}

// Kind 获取属性Kind
func (f *Field) Kind() reflect.Kind {
	return f.rvalue.Kind()
}

// IsSlice 判断属性是否是切片
func (f *Field) IsSlice() bool {
	k := f.rvalue.Kind()
	return k == reflect.Slice
}

// IsTargetType 判断属性是否是目标类型
func (f *Field) IsTargetType(targetType reflect.Kind) bool {
	return f.rvalue.Kind() == targetType
}

// mapValue covert field value to map
func (f *Field) mapValue(value any) any {
	val := pointer.ExtractPointer(value)
	v := reflect.ValueOf(val)
	var ret any

	switch v.Kind() {
	case reflect.Struct:
		s := New(val)
		s.TagName = f.TagName
		m, _ := s.ToMap()
		ret = m
	case reflect.Map:
		mapEl := v.Type().Elem()
		switch mapEl.Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Chan:
			// iterate the map
			m := make(map[string]any, v.Len())
			for _, key := range v.MapKeys() {
				m[key.String()] = f.mapValue(v.MapIndex(key).Interface())
			}
			ret = m
		default:
			ret = v.Interface()
		}
	case reflect.Slice, reflect.Array:
		sEl := v.Type().Elem()
		switch sEl.Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Chan:
			slices := make([]any, v.Len())
			for i := 0; i < v.Len(); i++ {
				slices[i] = f.mapValue(v.Index(i).Interface())
			}
			ret = slices
		default:
			ret = v.Interface()
		}
	default:
		if v.Kind().String() != "invalid" {
			ret = v.Interface()
		}
	}
	return ret
}
