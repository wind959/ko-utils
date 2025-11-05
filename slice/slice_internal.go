package slice

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"reflect"
)

type resultChunk[T comparable] struct {
	index int
	data  []T
}

func sliceValue(slice any) reflect.Value {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Invalid slice type, value of type %T", slice))
	}
	return v
}

func sliceElemType(reflectType reflect.Type) reflect.Type {
	for {
		if reflectType.Kind() != reflect.Slice {
			return reflectType
		}

		reflectType = reflectType.Elem()
	}
}

func quickSort[T constraints.Ordered](slice []T, lowIndex, highIndex int, order string) {
	if lowIndex < highIndex {
		p := partitionOrderedSlice(slice, lowIndex, highIndex, order)
		quickSort[T](slice, lowIndex, p-1, order)
		quickSort[T](slice, p+1, highIndex, order)
	}
}

func partitionOrderedSlice[T constraints.Ordered](slice []T, lowIndex, highIndex int, order string) int {
	p := slice[highIndex]
	i := lowIndex

	for j := lowIndex; j < highIndex; j++ {
		if order == "desc" {
			if slice[j] > p {
				swap(slice, i, j)
				i++
			}
		} else {
			if slice[j] < p {
				swap(slice, i, j)
				i++
			}
		}
	}

	swap(slice, i, highIndex)

	return i
}

func quickSortBy[T any](slice []T, lowIndex, highIndex int, less func(a, b T) bool) {
	if lowIndex < highIndex {
		p := partitionAnySlice(slice, lowIndex, highIndex, less)
		quickSortBy[T](slice, lowIndex, p-1, less)
		quickSortBy[T](slice, p+1, highIndex, less)
	}
}

func partitionAnySlice[T any](slice []T, lowIndex, highIndex int, less func(a, b T) bool) int {
	p := slice[highIndex]
	i := lowIndex

	for j := lowIndex; j < highIndex; j++ {

		if less(slice[j], p) {
			swap(slice, i, j)
			i++
		}
	}

	swap(slice, i, highIndex)

	return i
}

func swap[T any](slice []T, i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
