package slice

import (
	"fmt"
	"github.com/wind959/ko-utils/random"
	"golang.org/x/exp/constraints"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	memoryHashMap     = make(map[string]map[any]int)
	memoryHashCounter = make(map[string]int)
	muForMemoryHash   sync.RWMutex
)

// Contain 判断slice是否包含value
func Contain[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// ContainBy 根据predicate函数判断切片是否包含某个值
func ContainBy[T any](slice []T, predicate func(item T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// ContainSubSlice 判断slice是否包含subslice
func ContainSubSlice[T comparable](slice, subSlice []T) bool {
	if len(subSlice) == 0 {
		return true
	}
	if len(slice) == 0 {
		return false
	}

	elementCount := make(map[T]int, len(slice))
	for _, item := range slice {
		elementCount[item]++
	}

	for _, item := range subSlice {
		if elementCount[item] == 0 {
			return false
		}
		elementCount[item]--
	}

	return true
}

// Chunk 按照size参数均分slice
func Chunk[T any](slice []T, size int) [][]T {
	result := [][]T{}

	if len(slice) == 0 || size <= 0 {
		return result
	}

	currentChunk := []T{}

	for _, item := range slice {
		if len(currentChunk) == size {
			result = append(result, currentChunk)
			currentChunk = []T{}
		}
		currentChunk = append(currentChunk, item)
	}

	if len(currentChunk) > 0 {
		result = append(result, currentChunk)
	}

	return result
}

// Compact 去除slice中的假值（false values are false, nil, 0, ""）
func Compact[T comparable](slice []T) []T {
	var zero T

	result := make([]T, 0, len(slice))

	for _, v := range slice {
		if v != zero {
			result = append(result, v)
		}
	}

	return result[:len(result):len(result)]
}

// Concat 创建一个新的切片，将传入的切片拼接起来返回
func Concat[T any](slices ...[]T) []T {
	totalLen := 0
	for _, v := range slices {
		totalLen += len(v)
		if totalLen < 0 {
			panic("len out of range")
		}
	}
	result := make([]T, 0, totalLen)

	for _, v := range slices {
		result = append(result, v...)
	}

	return result
}

// Difference 创建一个切片，其元素不包含在另一个给定切片中
func Difference[T comparable](slice, comparedSlice []T) []T {
	result := []T{}

	if len(slice) == 0 {
		return result
	}

	comparedMap := make(map[T]struct{}, len(comparedSlice))
	for _, v := range comparedSlice {
		comparedMap[v] = struct{}{}
	}

	for _, v := range slice {
		if _, found := comparedMap[v]; !found {
			result = append(result, v)
		}
	}
	return result
}

// DifferenceBy 将两个slice中的每个元素调用iteratee函数，并比较它们的返回值，如果不相等返回在slice中对应的值
func DifferenceBy[T comparable](slice []T, comparedSlice []T, iteratee func(index int, item T) T) []T {
	result := make([]T, 0)

	comparedMap := make(map[T]struct{}, len(comparedSlice))
	for _, item := range comparedSlice {
		comparedMap[iteratee(0, item)] = struct{}{}
	}

	for i, item := range slice {
		transformedItem := iteratee(i, item)
		if _, found := comparedMap[transformedItem]; !found {
			result = append(result, item)
		}
	}

	return result
}

// DifferenceWith 接受比较器函数，该比较器被调用以将切片的元素与值进行比较。 结果值的顺序和引用由第一个切片确定
func DifferenceWith[T any](slice []T, comparedSlice []T, comparator func(item1, item2 T) bool) []T {
	getIndex := func(arr []T, item T, comparison func(v1, v2 T) bool) int {
		for i, v := range arr {
			if comparison(item, v) {
				return i
			}
		}
		return -1
	}

	result := make([]T, 0, len(slice))

	comparedMap := make(map[int]T, len(comparedSlice))
	for _, v := range comparedSlice {
		comparedMap[getIndex(comparedSlice, v, comparator)] = v
	}

	for _, v := range slice {
		found := false
		for _, existing := range comparedSlice {
			if comparator(v, existing) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, v)
		}
	}

	return result
}

// Equal  检查两个切片是否相等，相等条件：切片长度相同，元素顺序和值都相同
func Equal[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}

// EqualWith 检查两个切片是否相等，相等条件：对两个切片的元素调用比较函数comparator，返回true。
func EqualWith[T, U any](slice1 []T, slice2 []U, comparator func(T, U) bool) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i, v := range slice1 {
		if !comparator(v, slice2[i]) {
			return false
		}
	}

	return true
}

// EqualUnordered 检查两个切片是否相等，元素数量相同，值相等，不考虑元素顺序。
func EqualUnordered[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	seen := make(map[T]int)
	for _, v := range slice1 {
		seen[v]++
	}

	for _, v := range slice2 {
		if seen[v] == 0 {
			return false
		}
		seen[v]--
	}

	return true
}

// Every 如果切片中的所有值都通过谓词函数，则返回true。 函数签名应该是func(index int, value any) bool
func Every[T any](slice []T, predicate func(index int, item T) bool) bool {
	for i, v := range slice {
		if !predicate(i, v) {
			return false
		}
	}

	return true
}

func None[T any](slice []T, predicate func(index int, item T) bool) bool {
	l := 0
	for i, v := range slice {
		if !predicate(i, v) {
			l++
		}
	}

	return l == len(slice)
}

// Some 如果列表中的任何值通过谓词函数，则返回true
func Some[T any](slice []T, predicate func(index int, item T) bool) bool {
	for i, v := range slice {
		if predicate(i, v) {
			return true
		}
	}

	return false
}

// Filter 返回切片中通过predicate函数真值测试的所有元素
func Filter[T any](slice []T, predicate func(index int, item T) bool) []T {
	result := make([]T, 0)

	for i, v := range slice {
		if predicate(i, v) {
			result = append(result, v)
		}
	}

	return result
}

// Count 返回切片中指定元素的个数
func Count[T comparable](slice []T, item T) int {
	count := 0

	for _, v := range slice {
		if item == v {
			count++
		}
	}

	return count
}

// CountBy 遍历切片，对每个元素执行函数predicate. 返回符合函数返回值为true的元素的个数
func CountBy[T any](slice []T, predicate func(index int, item T) bool) int {
	count := 0

	for i, v := range slice {
		if predicate(i, v) {
			count++
		}
	}

	return count
}

// GroupBy 迭代切片的元素，每个元素将按条件分组，返回两个切片
func GroupBy[T any](slice []T, groupFn func(index int, item T) bool) ([]T, []T) {
	if len(slice) == 0 {
		return make([]T, 0), make([]T, 0)
	}

	groupB := make([]T, 0)
	groupA := make([]T, 0)

	for i, v := range slice {
		ok := groupFn(i, v)
		if ok {
			groupA = append(groupA, v)
		} else {
			groupB = append(groupB, v)
		}
	}

	return groupA, groupB
}

// GroupWith 创建一个map，key是iteratee遍历slice中的每个元素返回的结果。
// 分组值的顺序是由他们出现在slice中的顺序确定的。
// 每个键对应的值负责生成key的元素组成的数组。iteratee调用1个参数： (value)
func GroupWith[T any, U comparable](slice []T, iteratee func(item T) U) map[U][]T {
	result := make(map[U][]T)

	for _, v := range slice {
		key := iteratee(v)
		if _, ok := result[key]; !ok {
			result[key] = []T{}
		}
		result[key] = append(result[key], v)
	}

	return result
}

// FindBy 遍历slice的元素，返回第一个通过predicate函数真值测试的元素
func FindBy[T any](slice []T, predicate func(index int, item T) bool) (v T, ok bool) {
	index := -1

	for i, v := range slice {
		if predicate(i, v) {
			index = i
			break
		}
	}

	if index == -1 {
		return v, false
	}

	return slice[index], true
}

// FindLastBy 从遍历slice的元素，返回最后一个通过predicate函数真值测试的元素。
func FindLastBy[T any](slice []T, predicate func(index int, item T) bool) (v T, ok bool) {
	index := -1

	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(i, slice[i]) {
			index = i
			break
		}
	}

	if index == -1 {
		return v, false
	}

	return slice[index], true
}

// Flatten 将切片压平一层
func Flatten(slice any) any {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic("Flatten: input must be a slice")
	}

	elemType := sv.Type().Elem()
	if elemType.Kind() == reflect.Slice {
		elemType = elemType.Elem()
	}

	result := reflect.MakeSlice(reflect.SliceOf(elemType), 0, sv.Len())

	for i := 0; i < sv.Len(); i++ {
		item := sv.Index(i)
		if item.Kind() == reflect.Slice {
			for j := 0; j < item.Len(); j++ {
				result = reflect.Append(result, item.Index(j))
			}
		} else {
			result = reflect.Append(result, item)
		}
	}

	return result.Interface()
}

// FlattenDeep flattens slice 递归
func FlattenDeep(slice any) any {
	sv := sliceValue(slice)
	st := sliceElemType(sv.Type())

	tmp := reflect.MakeSlice(reflect.SliceOf(st), 0, 0)

	result := flattenRecursive(sv, tmp)

	return result.Interface()
}

func flattenRecursive(value reflect.Value, result reflect.Value) reflect.Value {
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		kind := item.Kind()

		if kind == reflect.Slice {
			result = flattenRecursive(item, result)
		} else {
			result = reflect.Append(result, item)
		}
	}

	return result
}

// ForEach 遍历切片的元素并为每个元素调用iteratee函数
func ForEach[T any](slice []T, iteratee func(index int, item T)) {
	for i := 0; i < len(slice); i++ {
		iteratee(i, slice[i])
	}
}

// ForEachWithBreak 遍历切片的元素并为每个元素调用iteratee函数，当iteratee函数返回false时，终止遍历
func ForEachWithBreak[T any](slice []T, iteratee func(index int, item T) bool) {
	for i := 0; i < len(slice); i++ {
		if !iteratee(i, slice[i]) {
			break
		}
	}
}

// Map 对slice中的每个元素执行map函数以创建一个新切片
func Map[T any, U any](slice []T, iteratee func(index int, item T) U) []U {
	result := make([]U, len(slice), cap(slice))

	for i := 0; i < len(slice); i++ {
		result[i] = iteratee(i, slice[i])
	}

	return result
}

// FilterMap 返回一个将filter和map操作应用于给定切片的切片。
// iteratee回调函数应该返回两个值：1，结果值。2，结果值是否应该被包含在返回的切片中。
func FilterMap[T any, U any](slice []T, iteratee func(index int, item T) (U, bool)) []U {
	result := []U{}

	for i, v := range slice {
		if a, ok := iteratee(i, v); ok {
			result = append(result, a)
		}
	}

	return result
}

// FlatMap 将切片转换为其它类型切片
func FlatMap[T any, U any](slice []T, iteratee func(index int, item T) []U) []U {
	result := make([]U, 0, len(slice))

	for i, v := range slice {
		result = append(result, iteratee(i, v)...)
	}

	return result
}

// ReduceBy 对切片元素执行reduce操作
func ReduceBy[T any, U any](slice []T, initial U, reducer func(index int, item T, agg U) U) U {
	accumulator := initial

	for i, v := range slice {
		accumulator = reducer(i, v, accumulator)
	}

	return accumulator
}

// ReduceRight 类似ReduceBy操作，迭代切片元素顺序从右至左
func ReduceRight[T any, U any](slice []T, initial U, reducer func(index int, item T, agg U) U) U {
	accumulator := initial

	for i := len(slice) - 1; i >= 0; i-- {
		accumulator = reducer(i, slice[i], accumulator)
	}

	return accumulator
}

// Replace 返回切片的副本，其中前n个不重叠的old替换为new
func Replace[T comparable](slice []T, old T, new T, n int) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i := range result {
		if result[i] == old && n != 0 {
			result[i] = new
			n--
		}
	}

	return result
}

// ReplaceAll 返回切片的副本，将其中old全部替换为new
func ReplaceAll[T comparable](slice []T, old T, new T) []T {
	return Replace(slice, old, new, -1)
}

// Repeat 创建一个切片，包含n个传入的item
func Repeat[T any](item T, n int) []T {
	result := make([]T, n)

	for i := range result {
		result[i] = item
	}

	return result
}

// InterfaceSlice 将值转换为接口切片
func InterfaceSlice(slice any) []any {
	sv := sliceValue(slice)
	if sv.IsNil() {
		return nil
	}

	result := make([]any, sv.Len())
	for i := 0; i < sv.Len(); i++ {
		result[i] = sv.Index(i).Interface()
	}

	return result
}

// StringSlice 将接口切片转换为字符串切片
func StringSlice(slice any) []string {
	v := sliceValue(slice)

	result := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		v, ok := v.Index(i).Interface().(string)
		if !ok {
			panic("invalid element type")
		}
		result[i] = v
	}

	return result
}

// IntSlice 将接口切片转换为int切片
func IntSlice(slice any) []int {
	sv := sliceValue(slice)

	result := make([]int, sv.Len())
	for i := 0; i < sv.Len(); i++ {
		v, ok := sv.Index(i).Interface().(int)
		if !ok {
			panic("invalid element type")
		}
		result[i] = v
	}

	return result
}

// DeleteAt  删除切片中指定索引的元素（不修改原切片）
func DeleteAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice[:len(slice)-1]
	}

	result := append([]T(nil), slice...)
	copy(result[index:], result[index+1:])

	// Set the last element to zero value, clean up the memory.
	result[len(result)-1] = zeroValue[T]()

	return result[:len(result)-1]
}

func zeroValue[T any]() T {
	var zero T
	return zero
}

// DeleteRange 删除切片中指定索引范围的元素（不修改原切片）.
func DeleteRange[T any](slice []T, start, end int) []T {
	result := make([]T, 0, len(slice)-(end-start))

	for i := 0; i < start; i++ {
		result = append(result, slice[i])
	}

	for i := end; i < len(slice); i++ {
		result = append(result, slice[i])
	}

	return result
}

// Drop 从切片的头部删除n个元素.
func Drop[T any](slice []T, n int) []T {
	size := len(slice)

	if size <= n {
		return []T{}
	}

	if n <= 0 {
		return slice
	}

	result := make([]T, 0, size-n)

	return append(result, slice[n:]...)
}

// DropRight 从切片的尾部删除n个元素.
func DropRight[T any](slice []T, n int) []T {
	size := len(slice)

	if size <= n {
		return []T{}
	}

	if n <= 0 {
		return slice
	}

	result := make([]T, 0, size-n)

	return append(result, slice[:size-n]...)
}

// DropWhile 从切片的头部删除n个元素，这个n个元素满足predicate函数返回true
func DropWhile[T any](slice []T, predicate func(item T) bool) []T {
	i := 0

	for ; i < len(slice); i++ {
		if !predicate(slice[i]) {
			break
		}
	}

	result := make([]T, 0, len(slice)-i)

	return append(result, slice[i:]...)
}

// DropRightWhile 从切片的尾部删除n个元素，这个n个元素满足predicate函数返回true
func DropRightWhile[T any](slice []T, predicate func(item T) bool) []T {
	i := len(slice) - 1

	for ; i >= 0; i-- {
		if !predicate(slice[i]) {
			break
		}
	}

	result := make([]T, 0, i+1)

	return append(result, slice[:i+1]...)
}

// InsertAt 将元素插入到索引处的切片中
func InsertAt[T any](slice []T, index int, value any) []T {
	size := len(slice)

	if index < 0 || index > size {
		return slice
	}

	switch v := value.(type) {
	case T:
		result := make([]T, size+1)
		copy(result, slice[:index])
		result[index] = v
		copy(result[index+1:], slice[index:])
		return result
	case []T:
		result := make([]T, size+len(v))
		copy(result, slice[:index])
		copy(result[index:], v)
		copy(result[index+len(v):], slice[index:])
		return result
	default:
		return slice
	}
}

// UpdateAt 更新索引处的切片元素。 如果index < 0或 index <= len(slice)，将返回错误
func UpdateAt[T any](slice []T, index int, value T) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}

	result := make([]T, len(slice))
	copy(result, slice)

	result[index] = value

	return result
}

// Unique 删除切片中的重复元素
func Unique[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[T]struct{}, len(slice))
	result := slice[:0]

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// UniqueBy 根据迭代函数返回的值，从输入切片中移除重复元素。此函数保持元素的顺序。
func UniqueBy[T any, U comparable](slice []T, iteratee func(item T) U) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[U]struct{}, len(slice))
	result := slice[:0]

	for _, item := range slice {
		key := iteratee(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// UniqueByComparator 使用提供的比较器函数从输入切片中移除重复元素。此函数保持元素的顺序
func UniqueByComparator[T comparable](slice []T, comparator func(item T, other T) bool) []T {
	if len(slice) == 0 {
		return slice
	}

	result := make([]T, 0, len(slice))
	for _, item := range slice {
		isDuplicate := false
		for _, existing := range result {
			if comparator(item, existing) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			result = append(result, item)
		}
	}

	return result
}

// UniqueByField 根据struct字段对struct切片去重复
func UniqueByField[T any](slice []T, field string) ([]T, error) {
	seen := map[any]struct{}{}

	var result []T
	for _, item := range slice {
		val, err := getField(item, field)
		if err != nil {
			return nil, fmt.Errorf("get field %s failed: %v", field, err)
		}
		if _, ok := seen[val]; !ok {
			seen[val] = struct{}{}
			result = append(result, item)
		}
	}

	return result, nil
}

func getField[T any](item T, field string) (interface{}, error) {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data type %T not support, shuld be struct or pointer to struct", item)
	}

	f := v.FieldByName(field)
	if !f.IsValid() {
		return nil, fmt.Errorf("field name %s not found", field)
	}

	return v.FieldByName(field).Interface(), nil
}

// Union 合并多个切片
func Union[T comparable](slices ...[]T) []T {
	result := []T{}
	contain := map[T]struct{}{}

	for _, slice := range slices {
		for _, item := range slice {
			if _, ok := contain[item]; !ok {
				contain[item] = struct{}{}
				result = append(result, item)
			}
		}
	}

	return result
}

// UnionBy 对切片的每个元素调用函数后，合并多个切片
func UnionBy[T any, V comparable](predicate func(item T) V, slices ...[]T) []T {
	result := []T{}
	contain := map[V]struct{}{}

	for _, slice := range slices {
		for _, item := range slice {
			val := predicate(item)
			if _, ok := contain[val]; !ok {
				contain[val] = struct{}{}
				result = append(result, item)
			}
		}
	}

	return result
}

// Merge 合并多个切片（不会消除重复元素).
func Merge[T any](slices ...[]T) []T {
	return Concat(slices...)
}

// Intersection 多个切片的交集
func Intersection[T comparable](slices ...[]T) []T {
	result := []T{}
	elementCount := make(map[T]int)

	for _, slice := range slices {
		seen := make(map[T]bool)

		for _, item := range slice {
			if !seen[item] {
				seen[item] = true
				elementCount[item]++
			}
		}
	}

	for _, item := range slices[0] {
		if elementCount[item] == len(slices) {
			result = append(result, item)
			elementCount[item] = 0
		}
	}

	return result
}

// SymmetricDifference 返回一个切片，其中的元素存在于参数切片中，但不同时存储在于参数切片中（交集取反）
func SymmetricDifference[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}
	if len(slices) == 1 {
		return Unique(slices[0])
	}

	result := make([]T, 0)

	intersectSlice := Intersection(slices...)

	for i := 0; i < len(slices); i++ {
		slice := slices[i]
		for _, v := range slice {
			if !Contain(intersectSlice, v) {
				result = append(result, v)
			}
		}

	}

	return Unique(result)
}

// Reverse 反转切片中的元素顺序
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// ReverseCopy 反转切片中的元素顺序, 不改变原slice
func ReverseCopy[T any](slice []T) []T {
	result := make([]T, len(slice))

	for i, j := 0, len(slice)-1; i < len(slice); i, j = i+1, j-1 {
		result[i] = slice[j]
	}

	return result
}

// Shuffle 随机打乱切片中的元素顺序
func Shuffle[T any](slice []T) []T {
	rand.Seed(time.Now().UnixNano())

	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	return slice
}

// ShuffleCopy 随机打乱切片中的元素顺序, 不改变原切片.
func ShuffleCopy[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// IsAscending 检查切片元素是否按升序排列
func IsAscending[T constraints.Ordered](slice []T) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i-1] > slice[i] {
			return false
		}
	}

	return true
}

// IsDescending 检查切片元素是否按降序排列
func IsDescending[T constraints.Ordered](slice []T) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i-1] < slice[i] {
			return false
		}
	}

	return true
}

// IsSorted 检查切片元素是否是有序的（升序或降序）
func IsSorted[T constraints.Ordered](slice []T) bool {
	return IsAscending(slice) || IsDescending(slice)
}

// IsSortedByKey 通过iteratee函数，检查切片元素是否是有序的
func IsSortedByKey[T any, K constraints.Ordered](slice []T, iteratee func(item T) K) bool {
	size := len(slice)

	isAscending := func(data []T) bool {
		for i := 0; i < size-1; i++ {
			if iteratee(data[i]) > iteratee(data[i+1]) {
				return false
			}
		}

		return true
	}

	isDescending := func(data []T) bool {
		for i := 0; i < size-1; i++ {
			if iteratee(data[i]) < iteratee(data[i+1]) {
				return false
			}
		}

		return true
	}

	return isAscending(slice) || isDescending(slice)
}

// Sort 对任何有序类型（数字或字符串）的切片进行排序，使用快速排序算法。 默认排序顺序为升序 (asc)，
// 如果需要降序，请将参数 `sortOrder` 设置为 `desc`。
// Ordered类型：数字（所有整数浮点数）或字符串。
func Sort[T constraints.Ordered](slice []T, sortOrder ...string) {
	if len(sortOrder) > 0 && sortOrder[0] == "desc" {
		quickSort(slice, 0, len(slice)-1, "desc")
	} else {
		quickSort(slice, 0, len(slice)-1, "asc")
	}
}

// SortBy 按照less函数确定的升序规则对切片进行排序。排序不保证稳定性
func SortBy[T any](slice []T, less func(a, b T) bool) {
	quickSortBy(slice, 0, len(slice)-1, less)
}

// SortByField 按字段对结构体切片进行排序。slice元素应为struct，排序字段field类型应为int、uint、string或bool。
// 默认排序类型是升序（asc），
// 如果是降序，设置 sortType 为 desc
func SortByField[T any](slice []T, field string, sortType ...string) error {
	sv := sliceValue(slice)
	t := sv.Type().Elem()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("data type %T not support, shuld be struct or pointer to struct", slice)
	}

	// Find the field.
	sf, ok := t.FieldByName(field)
	if !ok {
		return fmt.Errorf("field name %s not found", field)
	}

	// Create a less function based on the field's kind.
	var compare func(a, b reflect.Value) bool
	switch sf.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(sortType) > 0 && sortType[0] == "desc" {
			compare = func(a, b reflect.Value) bool { return a.Int() > b.Int() }
		} else {
			compare = func(a, b reflect.Value) bool { return a.Int() < b.Int() }
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if len(sortType) > 0 && sortType[0] == "desc" {
			compare = func(a, b reflect.Value) bool { return a.Uint() > b.Uint() }
		} else {
			compare = func(a, b reflect.Value) bool { return a.Uint() < b.Uint() }
		}
	case reflect.Float32, reflect.Float64:
		if len(sortType) > 0 && sortType[0] == "desc" {
			compare = func(a, b reflect.Value) bool { return a.Float() > b.Float() }
		} else {
			compare = func(a, b reflect.Value) bool { return a.Float() < b.Float() }
		}
	case reflect.String:
		if len(sortType) > 0 && sortType[0] == "desc" {
			compare = func(a, b reflect.Value) bool { return a.String() > b.String() }
		} else {
			compare = func(a, b reflect.Value) bool { return a.String() < b.String() }
		}
	case reflect.Bool:
		if len(sortType) > 0 && sortType[0] == "desc" {
			compare = func(a, b reflect.Value) bool { return a.Bool() && !b.Bool() }
		} else {
			compare = func(a, b reflect.Value) bool { return !a.Bool() && b.Bool() }
		}
	default:
		return fmt.Errorf("field type %s not supported", sf.Type)
	}

	sort.Slice(slice, func(i, j int) bool {
		a := sv.Index(i)
		b := sv.Index(j)
		if t.Kind() == reflect.Ptr {
			a = a.Elem()
			b = b.Elem()
		}
		a = a.FieldByIndex(sf.Index)
		b = b.FieldByIndex(sf.Index)
		return compare(a, b)
	})

	return nil
}

// Without 创建一个不包括所有给定值的切片
func Without[T comparable](slice []T, items ...T) []T {
	if len(items) == 0 || len(slice) == 0 {
		return slice
	}

	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if !Contain(items, v) {
			result = append(result, v)
		}
	}

	return result
}

// IndexOf 返回在切片中找到值的第一个匹配项的索引，如果找不到值，则返回-1
func IndexOf[T comparable](arr []T, val T) int {
	limit := 10
	// gets the hash value of the array as the key of the hash table.
	key := fmt.Sprintf("%p", arr)

	muForMemoryHash.RLock()
	// determines whether the hash table is empty. If so, the hash table is created.
	if memoryHashMap[key] == nil {

		muForMemoryHash.RUnlock()
		muForMemoryHash.Lock()

		if memoryHashMap[key] == nil {
			memoryHashMap[key] = make(map[any]int)
			// iterate through the array, adding the value and index of each element to the hash table.
			for i := len(arr) - 1; i >= 0; i-- {
				memoryHashMap[key][arr[i]] = i
			}
		}

		muForMemoryHash.Unlock()
	} else {
		muForMemoryHash.RUnlock()
	}

	muForMemoryHash.Lock()
	// update the hash table counter.
	memoryHashCounter[key]++
	muForMemoryHash.Unlock()

	// use the hash table to find the specified value. If found, the index is returned.
	muForMemoryHash.RLock()
	index, ok := memoryHashMap[key][val]
	muForMemoryHash.RUnlock()

	if ok {
		muForMemoryHash.RLock()
		// calculate the memory usage of the hash table.
		size := len(memoryHashMap)
		muForMemoryHash.RUnlock()

		// If the memory usage of the hash table exceeds the memory limit, the hash table with the lowest counter is cleared.
		if size > limit {
			muForMemoryHash.Lock()
			var minKey string
			var minVal int
			for k, v := range memoryHashCounter {
				if k == key {
					continue
				}
				if minVal == 0 || v < minVal {
					minKey = k
					minVal = v
				}
			}
			delete(memoryHashMap, minKey)
			delete(memoryHashCounter, minKey)
			muForMemoryHash.Unlock()
		}
		return index
	}
	return -1
}

// LastIndexOf 返回在切片中找到最后一个值的索引，如果找不到该值，则返回-1
func LastIndexOf[T comparable](slice []T, item T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if item == slice[i] {
			return i
		}
	}

	return -1
}

// ToSlicePointer 将可变参数转为指针切片
func ToSlicePointer[T any](items ...T) []*T {
	result := make([]*T, len(items))
	for i := range items {
		result[i] = &items[i]
	}

	return result
}

// ToSlice 将可变参数转为切片
func ToSlice[T any](items ...T) []T {
	result := make([]T, len(items))
	copy(result, items)

	return result
}

// AppendIfAbsent 当前切片中不包含值时，将该值追加到切片中
func AppendIfAbsent[T comparable](slice []T, item T) []T {
	if !Contain(slice, item) {
		slice = append(slice, item)
	}
	return slice
}

// SetToDefaultIf 根据给定给定的predicate判定函数来修改切片中的元素。
// 对于满足的元素，将其替换为指定的默认值，同时保持元素在切片中的位置不变
// 。函数返回修改后的切片以及被修改的元素个数。
func SetToDefaultIf[T any](slice []T, predicate func(T) bool) ([]T, int) {
	var count int
	for i := 0; i < len(slice); i++ {
		if predicate(slice[i]) {
			var zeroValue T
			slice[i] = zeroValue
			count++
		}
	}
	return slice, count
}

// KeyBy 将切片每个元素调用函数后转为map
func KeyBy[T any, U comparable](slice []T, iteratee func(item T) U) map[U]T {
	result := make(map[U]T, len(slice))

	for _, v := range slice {
		k := iteratee(v)
		result[k] = v
	}

	return result
}

// Join 用指定的分隔符链接切片元素
func Join[T any](slice []T, separator string) string {
	str := Map(slice, func(_ int, item T) string {
		return fmt.Sprint(item)
	})

	return strings.Join(str, separator)
}

// Partition 根据给定的predicate判断函数分组切片元素
func Partition[T any](slice []T, predicates ...func(item T) bool) [][]T {
	l := len(predicates)

	result := make([][]T, l+1)

	for _, item := range slice {
		processed := false

		for i, f := range predicates {
			if f == nil {
				panic("predicate function must not be nill")
			}

			if f(item) {
				result[i] = append(result[i], item)
				processed = true
				break
			}
		}

		if !processed {
			result[l] = append(result[l], item)
		}
	}

	return result
}

// Break 根据判断函数将切片分成两部分。
// 它开始附加到与函数匹配的第一个元素之后的第二个切片。
// 第一个匹配之后的所有元素都包含在第二个切片中，无论它们是否与函数匹配。
func Break[T any](values []T, predicate func(T) bool) ([]T, []T) {
	a := make([]T, 0)
	b := make([]T, 0)
	if len(values) == 0 {
		return a, b
	}
	matched := false
	for _, value := range values {

		if !matched && predicate(value) {
			matched = true
		}

		if matched {
			b = append(b, value)
		} else {
			a = append(a, value)
		}
	}
	return a, b
}

// Random 随机返回切片中元素以及下标, 当切片长度为0时返回下标-1
func Random[T any](slice []T) (val T, idx int) {
	if len(slice) == 0 {
		return val, -1
	}

	idx = random.RandInt(0, len(slice))
	return slice[idx], idx
}

// RightPadding 在切片的右部添加元素
func RightPadding[T any](slice []T, paddingValue T, paddingLength int) []T {
	if paddingLength == 0 {
		return slice
	}
	for i := 0; i < paddingLength; i++ {
		slice = append(slice, paddingValue)
	}
	return slice
}

// LeftPadding 在切片的左部添加元素
func LeftPadding[T any](slice []T, paddingValue T, paddingLength int) []T {
	if paddingLength == 0 {
		return slice
	}

	paddedSlice := make([]T, len(slice)+paddingLength)
	i := 0
	for ; i < paddingLength; i++ {
		paddedSlice[i] = paddingValue
	}
	for j := 0; j < len(slice); j++ {
		paddedSlice[i] = slice[j]
		i++
	}

	return paddedSlice
}

// Frequency 计算切片中每个元素出现的频率
func Frequency[T comparable](slice []T) map[T]int {
	result := make(map[T]int)

	for _, v := range slice {
		result[v]++
	}

	return result
}

// JoinFunc 将切片元素用给定的分隔符连接成一个单一的字符串。
func JoinFunc[T any](slice []T, sep string, transform func(T) T) string {
	var buf strings.Builder
	for i, v := range slice {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(fmt.Sprint(transform(v)))
	}
	return buf.String()
}

// ConcatBy 将切片中的元素连接成一个值，使用指定的分隔符和连接器函数
func ConcatBy[T any](slice []T, sep T, connector func(T, T) T) T {
	var result T

	if len(slice) == 0 {
		return result
	}

	for i, v := range slice {
		result = connector(result, v)
		if i < len(slice)-1 {
			result = connector(result, sep)
		}
	}

	return result
}
