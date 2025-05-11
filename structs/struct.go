package structs

import (
	"fmt"
	"github.com/wind959/ko-utils/pointer"
	"reflect"
)

// defaultTagName is the default tag for struct fields to lookup.
var defaultTagName = "json"

// Struct is abstract struct for provide several high level functions
type Struct struct {
	raw     any
	rtype   reflect.Type
	rvalue  reflect.Value
	TagName string
}

// New `Struct`结构体的构造函数
func New(value any, tagName ...string) *Struct {
	value = pointer.ExtractPointer(value)
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	tn := defaultTagName

	if len(tagName) > 0 {
		tn = tagName[0]
	}

	return &Struct{
		raw:     value,
		rtype:   t,
		rvalue:  v,
		TagName: tn,
	}
}

// ToMap 将一个合法的struct对象转换为map[string]any
func (s *Struct) ToMap() (map[string]any, error) {
	if !s.IsStruct() {
		return nil, fmt.Errorf("invalid struct %v", s)
	}

	result := make(map[string]any)
	fields := s.Fields()
	for _, f := range fields {
		if !f.IsExported() || f.tag.IsEmpty() || f.tag.Name == "-" {
			continue
		}

		if f.IsZero() && f.tag.HasOption("omitempty") {
			continue
		}

		if f.IsNil() {
			continue
		}

		result[f.tag.Name] = f.mapValue(f.Value())
	}

	return result, nil
}

// Fields 获取一个struct对象的属性列表
func (s *Struct) Fields() []*Field {
	fieldNum := s.rvalue.NumField()
	fields := make([]*Field, 0, fieldNum)
	for i := 0; i < fieldNum; i++ {
		v := s.rvalue.Field(i)
		sf := s.rtype.Field(i)
		field := newField(v, sf, s.TagName)
		fields = append(fields, field)
	}
	return fields
}

// Field 根据属性名获取一个struct对象的属性
func (s *Struct) Field(name string) (*Field, bool) {
	f, ok := s.rtype.FieldByName(name)
	if !ok {
		return nil, false
	}
	return newField(s.rvalue.FieldByName(name), f, s.TagName), true
}

// IsStruct 判断是否为一个合法的struct对象
func (s *Struct) IsStruct() bool {
	k := s.rvalue.Kind()
	if k == reflect.Invalid {
		return false
	}
	return k == reflect.Struct
}

// ToMap 将一个合法的struct对象转换为map[string]any
func ToMap(v any) (map[string]any, error) {
	return New(v).ToMap()
}
