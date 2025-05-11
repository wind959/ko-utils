package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Marshal 将 Go 对象序列化为 JSON 字符串
func Marshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("json marshal failed: %v", err)
	}
	return string(data), nil
}

// MarshalIndent 将 Go 对象序列化为格式化的 JSON 字符串
func MarshalIndent(v interface{}, prefix, indent string) (string, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return "", fmt.Errorf("json marshal indent failed: %v", err)
	}
	return string(data), nil
}

// Unmarshal 将 JSON 字符串反序列化为 Go 对象
func Unmarshal(data string, v interface{}) error {
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("json unmarshal failed: %v", err)
	}
	return nil
}

// PrettyPrint 格式化 JSON 字符串
func PrettyPrint(data string) (string, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(data), "", "  "); err != nil {
		return "", fmt.Errorf("json pretty print failed: %v", err)
	}
	return out.String(), nil
}

// Compress 压缩 JSON 字符串（移除空格和换行）
func Compress(data string) (string, error) {
	var out bytes.Buffer
	if err := json.Compact(&out, []byte(data)); err != nil {
		return "", fmt.Errorf("json compress failed: %v", err)
	}
	return out.String(), nil
}
