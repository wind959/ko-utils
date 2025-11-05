package jsonutil

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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

// UnmarshalBytes 将 JSON 字节数组反序列化为 Go 对象
func UnmarshalBytes(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
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

// PrettyPrintBytes 格式化 JSON 字节数组
func PrettyPrintBytes(data []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		return nil, fmt.Errorf("json pretty print failed: %v", err)
	}
	return out.Bytes(), nil
}

// Compress 压缩 JSON 字符串（移除空格和换行）
func Compress(data string) (string, error) {
	var out bytes.Buffer
	if err := json.Compact(&out, []byte(data)); err != nil {
		return "", fmt.Errorf("json compress failed: %v", err)
	}
	return out.String(), nil
}

// CompressBytes 压缩 JSON 字节数组（移除空格和换行）
func CompressBytes(data []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Compact(&out, data); err != nil {
		return nil, fmt.Errorf("json compress failed: %v", err)
	}
	return out.Bytes(), nil
}

// ToXML 将 JSON 字符串转换为 XML 格式
func ToXML(jsonStr string) (string, error) {
	// 先将JSON解析为Go对象
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 创建XML根元素
	xmlBuilder := &strings.Builder{}
	xmlBuilder.WriteString("<root>")

	// 递归转换对象到XML
	if err := convertToXML(obj, xmlBuilder, ""); err != nil {
		return "", err
	}

	xmlBuilder.WriteString("</root>")

	return xml.Header + xmlBuilder.String(), nil
}

// convertToXML 递归地将Go对象转换为XML字符串
func convertToXML(obj interface{}, builder *strings.Builder, indent string) error {
	switch v := obj.(type) {
	case nil:
		// nil值不输出
		return nil
	case string:
		builder.WriteString(escapeXML(v))
	case bool:
		builder.WriteString(strconv.FormatBool(v))
	case float64:
		// 检查是否为整数
		if v == float64(int64(v)) {
			builder.WriteString(strconv.FormatInt(int64(v), 10))
		} else {
			builder.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
		}
	case []interface{}:
		for i, item := range v {
			builder.WriteString(fmt.Sprintf("%s<item index=\"%d\">", indent, i))
			if err := convertToXML(item, builder, indent+"  "); err != nil {
				return err
			}
			builder.WriteString("</item>")
		}
	case map[string]interface{}:
		for key, value := range v {
			// 确保标签名是有效的XML名称
			tagName := sanitizeXMLName(key)
			builder.WriteString(fmt.Sprintf("%s<%s>", indent, tagName))
			if err := convertToXML(value, builder, indent+"  "); err != nil {
				return err
			}
			builder.WriteString(fmt.Sprintf("</%s>", tagName))
		}
	default:
		builder.WriteString(fmt.Sprintf("%v", v))
	}

	return nil
}

// escapeXML 转义XML特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// sanitizeXMLName 确保XML标签名是有效的
func sanitizeXMLName(name string) string {
	if name == "" {
		return "empty"
	}

	// XML标签名不能以数字开头
	if name[0] >= '0' && name[0] <= '9' {
		name = "tag_" + name
	}

	// 替换无效字符
	var result strings.Builder
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' ||
			(r == '.' && i > 0) || (r == ':' && i > 0) {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	return result.String()
}

// IsValid 检查字符串是否为有效的JSON
func IsValid(jsonStr string) bool {
	return json.Valid([]byte(jsonStr))
}

// IsValidBytes 检查字节数组是否为有效的JSON
func IsValidBytes(jsonBytes []byte) bool {
	return json.Valid(jsonBytes)
}

// GetBytes 将 Go 对象序列化为 JSON 字节数组
func GetBytes(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %v", err)
	}
	return data, nil
}

// GetBytesIndent 将 Go 对象序列化为格式化的 JSON 字节数组
func GetBytesIndent(v interface{}, prefix, indent string) ([]byte, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, fmt.Errorf("json marshal indent failed: %v", err)
	}
	return data, nil
}

// ReadJSON 从文件中读取JSON数据到指定对象
func ReadJSON(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from file %s: %v", filename, err)
	}

	return nil
}

// ReadJSONObject 从文件中读取JSON对象
func ReadJSONObject(filename string) (map[string]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON object from file %s: %v", filename, err)
	}

	return obj, nil
}

// ReadJSONArray 从文件中读取JSON数组
func ReadJSONArray(filename string) ([]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON array from file %s: %v", filename, err)
	}

	return arr, nil
}
