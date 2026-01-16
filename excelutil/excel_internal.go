package excelutil

import (
	"reflect"
	"strconv"
	"strings"
)

// fillStruct 填充结构体
func (e *Excel) fillStruct(elem reflect.Value, row []string) error {

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if i >= len(row) {
			continue
		}

		value := strings.TrimSpace(row[i])
		if value == "" {
			continue
		}

		if err := e.setValue(field, value); err != nil {
			return err
		}
	}
	return nil
}

// setValue 设置值
func (e *Excel) setValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return nil
	}
	return nil
}

// structToRow 结构体转行数据
func (e *Excel) structToRow(elem reflect.Value) []interface{} {
	row := make([]interface{}, elem.NumField())
	for i := 0; i < elem.NumField(); i++ {
		row[i] = elem.Field(i).Interface()
	}
	return row
}
