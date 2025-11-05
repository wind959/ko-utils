package maputil

import (
	"fmt"
	"github.com/wind959/ko-utils/slice"
	"golang.org/x/exp/constraints"
	"reflect"
	"sort"
	"strings"
)

// Keys 返回map中所有key的切片
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))

	var i int
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// Values 返回map中所有value的切片
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))

	var i int
	for _, v := range m {
		values[i] = v
		i++
	}

	return values
}

// KeysBy 创建一个切片，其元素是每个map的key调用mapper函数的结果
func KeysBy[K comparable, V any, T any](m map[K]V, mapper func(item K) T) []T {
	keys := make([]T, 0, len(m))

	for k := range m {
		keys = append(keys, mapper(k))
	}

	return keys
}

// ValuesBy 创建一个切片，其元素是每个map的value调用mapper函数的结果
func ValuesBy[K comparable, V any, T any](m map[K]V, mapper func(item V) T) []T {
	keys := make([]T, 0, len(m))

	for _, v := range m {
		keys = append(keys, mapper(v))
	}

	return keys
}

// Merge 合并多个maps, 相同的key会被后来的key覆盖
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	size := 0
	for i := range maps {
		size += len(maps[i])
	}

	result := make(map[K]V, size)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// ForEach 对map中的每对key和value执行iteratee函数
func ForEach[K comparable, V any](m map[K]V, iteratee func(key K, value V)) {
	for k, v := range m {
		iteratee(k, v)
	}
}

// Filter 迭代map中的每对key和value, 返回符合predicate函数的key, value
func Filter[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) map[K]V {
	result := make(map[K]V)

	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// FilterByKeys 迭代map, 返回一个新map，其key都是给定的key值
func FilterByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)

	for k, v := range m {
		if slice.Contain(keys, k) {
			result[k] = v
		}
	}
	return result
}

// FilterByValues 迭代map, 返回一个新map，其value都是给定的value值
func FilterByValues[K comparable, V comparable](m map[K]V, values []V) map[K]V {
	result := make(map[K]V)

	for k, v := range m {
		if slice.Contain(values, v) {
			result[k] = v
		}
	}
	return result
}

// OmitBy Filter的反向操作, 迭代map中的每对key和value, 删除符合predicate函数的key, value, 返回新map
func OmitBy[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) map[K]V {
	result := make(map[K]V)

	for k, v := range m {
		if !predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// OmitByKeys FilterByKeys的反向操作, 迭代map, 返回一个新map，其key不包括给定的key值
func OmitByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)

	for k, v := range m {
		if !slice.Contain(keys, k) {
			result[k] = v
		}
	}
	return result
}

// OmitByValues FilterByValues的反向操作, 迭代map, 返回一个新map，其value不包括给定的value值
func OmitByValues[K comparable, V comparable](m map[K]V, values []V) map[K]V {
	result := make(map[K]V)

	for k, v := range m {
		if !slice.Contain(values, v) {
			result[k] = v
		}
	}
	return result
}

// Intersect 多个map的交集操作
func Intersect[K comparable, V any](maps ...map[K]V) map[K]V {
	if len(maps) == 0 {
		return map[K]V{}
	}
	if len(maps) == 1 {
		return maps[0]
	}

	var result map[K]V

	reducer := func(m1, m2 map[K]V) map[K]V {
		m := make(map[K]V)
		for k, v1 := range m1 {
			if v2, ok := m2[k]; ok && reflect.DeepEqual(v1, v2) {
				m[k] = v1
			}
		}
		return m
	}

	reduceMaps := make([]map[K]V, 2)
	result = reducer(maps[0], maps[1])

	for i := 2; i < len(maps); i++ {
		reduceMaps[0] = result
		reduceMaps[1] = maps[i]
		result = reducer(reduceMaps[0], reduceMaps[1])
	}

	return result
}

// Minus 返回一个map，其中的key存在于mapA，不存在于mapB
func Minus[K comparable, V any](mapA, mapB map[K]V) map[K]V {
	result := make(map[K]V)

	for k, v := range mapA {
		if _, ok := mapB[k]; !ok {
			result[k] = v
		}
	}
	return result
}

// IsDisjoint 验证两个map是否具有不同的key
func IsDisjoint[K comparable, V any](mapA, mapB map[K]V) bool {
	for k := range mapA {
		if _, ok := mapB[k]; ok {
			return false
		}
	}
	return true
}

// Entry is a key/value pairs.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// Entries 将map转换为键/值对切片
func Entries[K comparable, V any](m map[K]V) []Entry[K, V] {
	entries := make([]Entry[K, V], 0, len(m))

	for k, v := range m {
		entries = append(entries, Entry[K, V]{
			Key:   k,
			Value: v,
		})
	}

	return entries
}

// FromEntries 基于键/值对的切片创建map
func FromEntries[K comparable, V any](entries []Entry[K, V]) map[K]V {
	result := make(map[K]V, len(entries))

	for _, v := range entries {
		result[v.Key] = v.Value
	}

	return result
}

// Transform 将map转换为其他类型的map
func Transform[K1 comparable, V1 any, K2 comparable, V2 any](m map[K1]V1, iteratee func(key K1, value V1) (K2, V2)) map[K2]V2 {
	result := make(map[K2]V2, len(m))

	for k1, v1 := range m {
		k2, v2 := iteratee(k1, v1)
		result[k2] = v2
	}

	return result
}

// MapKeys 操作map的每个key，然后转为新的map
func MapKeys[K comparable, V any, T comparable](m map[K]V, iteratee func(key K, value V) T) map[T]V {
	result := make(map[T]V, len(m))

	for k, v := range m {
		result[iteratee(k, v)] = v
	}

	return result
}

// MapValues 操作map的每个value，然后转为新的map
func MapValues[K comparable, V any, T any](m map[K]V, iteratee func(key K, value V) T) map[K]T {
	result := make(map[K]T, len(m))

	for k, v := range m {
		result[k] = iteratee(k, v)
	}

	return result
}

// HasKey 检查map是否包含某个key
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, haskey := m[key]
	return haskey
}

// MapToStruct 将map转成struct
func MapToStruct(m map[string]any, structObj any) error {
	for k, v := range m {
		err := setStructField(structObj, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func setStructField(structObj any, fieldName string, fieldValue any) error {
	structVal := reflect.ValueOf(structObj).Elem()

	fName := getFieldNameByJsonTag(structObj, fieldName)
	if fName == "" {
		return fmt.Errorf("Struct field json tag don't match map key : %s in obj", fieldName)
	}

	fieldVal := structVal.FieldByName(fName)

	if !fieldVal.IsValid() {
		return fmt.Errorf("No such field: %s in obj", fieldName)
	}

	if !fieldVal.CanSet() {
		return fmt.Errorf("Cannot set %s field value", fieldName)
	}

	val := reflect.ValueOf(fieldValue)

	if fieldVal.Type() != val.Type() {

		if val.CanConvert(fieldVal.Type()) {
			fieldVal.Set(val.Convert(fieldVal.Type()))
			return nil
		}

		if m, ok := fieldValue.(map[string]any); ok {

			if fieldVal.Kind() == reflect.Struct {
				return MapToStruct(m, fieldVal.Addr().Interface())
			}

			if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}

				return MapToStruct(m, fieldVal.Interface())
			}

		}

		return fmt.Errorf("Map value type don't match struct field type")
	}

	fieldVal.Set(val)

	return nil
}

func getFieldNameByJsonTag(structObj any, jsonTag string) string {
	s := reflect.TypeOf(structObj).Elem()

	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tag := field.Tag
		name, _, _ := strings.Cut(tag.Get("json"), ",")
		if name == jsonTag {
			return field.Name
		}
	}

	return ""
}

// ToSortedSlicesDefault 将map的key和value转化成两个根据key的值从小到大排序的切片，value切片中元素的位置与key对应
func ToSortedSlicesDefault[K constraints.Ordered, V any](m map[K]V) ([]K, []V) {
	keys := make([]K, 0, len(m))

	// store the map’s keys into a slice
	for k := range m {
		keys = append(keys, k)
	}

	// sort the slice of keys
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// adjust the order of values according to the sorted keys
	sortedValues := make([]V, len(keys))
	for i, k := range keys {
		sortedValues[i] = m[k]
	}
	return keys, sortedValues
}

// ToSortedSlicesWithComparator 将map的key和value转化成两个使用比较器函数根据key的值自定义排序规则的切片，
// value切片中元素的位置与key对应
func ToSortedSlicesWithComparator[K comparable, V any](m map[K]V, comparator func(a, b K) bool) ([]K, []V) {
	keys := make([]K, 0, len(m))

	// store the map’s keys into a slice
	for k := range m {
		keys = append(keys, k)
	}

	// sort the key slice using the provided comparison function
	sort.Slice(keys, func(i, j int) bool {
		return comparator(keys[i], keys[j])
	})

	// adjust the order of values according to the sorted keys
	sortedValues := make([]V, len(keys))
	for i, k := range keys {
		sortedValues[i] = m[k]
	}

	return keys, sortedValues
}

// GetOrSet 返回给定键的值，如果不存在则设置该值
func GetOrSet[K comparable, V any](m map[K]V, key K, value V) V {
	if v, ok := m[key]; ok {
		return v
	}

	m[key] = value

	return value
}

// SortByKey 对传入的map根据key进行排序，返回排序后的map
func SortByKey[K constraints.Ordered, V any](m map[K]V, less func(a, b K) bool) (sortedKeysMap map[K]V) {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})

	sortedKeysMap = make(map[K]V, len(m))
	for _, k := range keys {
		sortedKeysMap[k] = m[k]
	}

	return
}

var mapHandlers = map[reflect.Kind]func(reflect.Value, reflect.Value) error{
	reflect.String:     convertNormal,
	reflect.Int:        convertNormal,
	reflect.Int16:      convertNormal,
	reflect.Int32:      convertNormal,
	reflect.Int64:      convertNormal,
	reflect.Uint:       convertNormal,
	reflect.Uint16:     convertNormal,
	reflect.Uint32:     convertNormal,
	reflect.Uint64:     convertNormal,
	reflect.Float32:    convertNormal,
	reflect.Float64:    convertNormal,
	reflect.Uint8:      convertNormal,
	reflect.Int8:       convertNormal,
	reflect.Struct:     convertNormal,
	reflect.Complex64:  convertNormal,
	reflect.Complex128: convertNormal,
}

var _ = func() struct{} {
	mapHandlers[reflect.Map] = convertMap
	mapHandlers[reflect.Array] = convertSlice
	mapHandlers[reflect.Slice] = convertSlice
	return struct{}{}
}()

// MapTo 快速将map或者其他类型映射到结构体或者指定类型
func MapTo(src any, dst any) error {
	dstRef := reflect.ValueOf(dst)

	if dstRef.Kind() != reflect.Ptr {
		return fmt.Errorf("dst is not ptr")
	}

	dstElem := dstRef.Type().Elem()
	if dstElem.Kind() == reflect.Struct {
		srcMap := src.(map[string]interface{})
		return MapToStruct(srcMap, dst)
	}

	dstRef = reflect.Indirect(dstRef)

	srcRef := reflect.ValueOf(src)
	if srcRef.Kind() == reflect.Ptr || srcRef.Kind() == reflect.Interface {
		srcRef = srcRef.Elem()
	}

	if f, ok := mapHandlers[srcRef.Kind()]; ok {
		return f(srcRef, dstRef)
	}

	return fmt.Errorf("no implemented:%s", srcRef.Type())
}

func convertNormal(src reflect.Value, dst reflect.Value) error {
	if dst.CanSet() {
		if src.Type() == dst.Type() {
			dst.Set(src)
		} else if src.CanConvert(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
		} else {
			return fmt.Errorf("can not convert:%s:%s", src.Type().String(), dst.Type().String())
		}
	}
	return nil
}

func convertSlice(src reflect.Value, dst reflect.Value) error {
	if dst.Kind() != reflect.Array && dst.Kind() != reflect.Slice {
		return fmt.Errorf("error type:%s", dst.Type().String())
	}
	l := src.Len()
	target := reflect.MakeSlice(dst.Type(), l, l)
	if dst.CanSet() {
		dst.Set(target)
	}
	for i := 0; i < l; i++ {
		srcValue := src.Index(i)
		if srcValue.Kind() == reflect.Ptr || srcValue.Kind() == reflect.Interface {
			srcValue = srcValue.Elem()
		}
		if f, ok := mapHandlers[srcValue.Kind()]; ok {
			err := f(srcValue, dst.Index(i))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertMap(src reflect.Value, dst reflect.Value) error {
	if src.Kind() != reflect.Map || dst.Kind() != reflect.Struct {
		if src.Kind() == reflect.Interface && dst.IsValid() {
			return convertMap(src.Elem(), dst)
		} else {
			return fmt.Errorf("src or dst type error,%s,%s", src.Type().String(), dst.Type().String())
		}
	}
	dstType := dst.Type()
	num := dstType.NumField()

	exist := map[string]int{}

	for i := 0; i < num; i++ {
		k := dstType.Field(i).Tag.Get("json")
		if k == "" {
			k = dstType.Field(i).Name
		}
		if strings.Contains(k, ",") {
			taglist := strings.Split(k, ",")
			if taglist[0] == "" {
				k = dstType.Field(i).Name
			} else {
				k = taglist[0]

			}

		}
		exist[k] = i
	}

	keys := src.MapKeys()

	for _, key := range keys {
		if index, ok := exist[key.String()]; ok {
			v := dst.Field(index)

			if v.Kind() == reflect.Struct {
				err := convertMap(src.MapIndex(key), v)
				if err != nil {
					return err
				}
			} else {
				if v.CanSet() {
					if v.Type() == src.MapIndex(key).Elem().Type() {
						v.Set(src.MapIndex(key).Elem())
					} else if src.MapIndex(key).Elem().CanConvert(v.Type()) {
						v.Set(src.MapIndex(key).Elem().Convert(v.Type()))
					} else if f, ok := mapHandlers[src.MapIndex(key).Elem().Kind()]; ok && f != nil {
						err := f(src.MapIndex(key).Elem(), v)
						if err != nil {
							return err
						}
					} else {
						return fmt.Errorf("error type:d(%s)s(%s)", v.Type(), src.Type())
					}
				}
			}
		}
	}

	return nil
}

// GetOrDefault 返回给定键的值，如果键不存在，则返回默认值
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

// FindValuesBy returns a slice of values from the map that satisfy the given predicate function.
func FindValuesBy[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) []V {
	result := make([]V, 0)

	for k, v := range m {
		if predicate(k, v) {
			result = append(result, v)
		}
	}
	return result
}
