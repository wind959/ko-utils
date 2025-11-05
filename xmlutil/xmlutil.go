package xmlutil

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Document represents an XML document
type Document struct {
	Root *Element
}

// Element represents an XML element
type Element struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:"-"`
	Children []Element  `xml:",any"`
	Text     string     `xml:",chardata"`
}

// ReadFile 读取XML文件
func ReadFile(filename string) (*Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(file)
}

// Parse 解析XML字符串为Document对象
func Parse(reader io.Reader) (*Document, error) {
	doc := &Document{}
	decoder := xml.NewDecoder(reader)

	// 去除XML文本中的无效字符
	decoder.CharsetReader = nil

	// 解码XML
	err := decoder.Decode(&doc.Root)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// ParseString 解析XML字符串为Document对象
func ParseString(xmlStr string) (*Document, error) {
	// 去除无效字符
	cleaned := CleanInvalidChars(xmlStr)
	return Parse(strings.NewReader(cleaned))
}

// ToString 将XML文档转换为String
func (doc *Document) ToString() (string, error) {
	if doc.Root == nil {
		return "", fmt.Errorf("document has no root element")
	}

	output, err := xml.MarshalIndent(doc.Root, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(output), nil
}

// WriteToFile 将XML文档写入到文件
func (doc *Document) WriteToFile(filename string) error {
	xmlStr, err := doc.ToString()
	if err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(xmlStr), 0644)
}

// CreateDocument 创建XML文档(默认UTF-8编码)
func CreateDocument(rootName string) *Document {
	doc := &Document{
		Root: &Element{
			XMLName: xml.Name{Local: rootName},
		},
	}
	return doc
}

// CleanInvalidChars 去除XML文本中的无效字符
func CleanInvalidChars(s string) string {
	var cleaned []rune
	for _, r := range s {
		// 保留有效的XML字符
		// XML 1.0 valid chars: #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]
		if r == 0x09 || r == 0x0A || r == 0x0D ||
			(r >= 0x20 && r <= 0xD7FF) ||
			(r >= 0xE000 && r <= 0xFFFD) ||
			(r >= 0x10000 && r <= 0x10FFFF) {
			cleaned = append(cleaned, r)
		}
	}
	return string(cleaned)
}

// GetElementsByTagName 根据节点名获得子节点列表
func (elem *Element) GetElementsByTagName(tagName string) []*Element {
	var result []*Element

	// 递归搜索所有子元素
	var findElements func(*Element)
	findElements = func(e *Element) {
		for i := range e.Children {
			child := &e.Children[i]
			if child.XMLName.Local == tagName {
				result = append(result, child)
			}
			findElements(child)
		}
	}

	findElements(elem)
	return result
}

// GetElementByTagName 根据节点名获得第一个子节点
func (elem *Element) GetElementByTagName(tagName string) *Element {
	elements := elem.GetElementsByTagName(tagName)
	if len(elements) > 0 {
		return elements[0]
	}
	return nil
}

// GetElementTextByTagName 根据节点名获得第一个子节点的文本值
func (elem *Element) GetElementTextByTagName(tagName string) string {
	element := elem.GetElementByTagName(tagName)
	if element != nil {
		return element.Text
	}
	return ""
}

// NodeListToElementList 将NodeList转换为Element列表
// 注意：在Go中，我们直接返回Element切片而不是NodeList
func NodeListToElementList(elements []*Element) []*Element {
	return elements
}

// AddChild 添加子元素
func (elem *Element) AddChild(child *Element) {
	elem.Children = append(elem.Children, *child)
}

// SetAttribute 设置属性
func (elem *Element) SetAttribute(name, value string) {
	elem.Attrs = append(elem.Attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

// GetAttribute 获取属性值
func (elem *Element) GetAttribute(name string) string {
	for _, attr := range elem.Attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// GetAllElements 获取所有元素(包括嵌套的)
func (elem *Element) GetAllElements() []*Element {
	var result []*Element
	result = append(result, elem)

	for i := range elem.Children {
		child := &elem.Children[i]
		result = append(result, child.GetAllElements()...)
	}

	return result
}
