package urlutil

import (
	"net/url"
	"regexp"
	"strings"
)

// Normalize 标准化URL链接
func Normalize(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	// 保存原始URL的查询参数部分
	var query string
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		query = rawURL[idx:]
		rawURL = rawURL[:idx]
	}
	// 处理协议部分，确保协议后面只有两个斜杠
	protocolRegex := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*):/+`)
	rawURL = protocolRegex.ReplaceAllString(rawURL, "$1://")
	// 查找协议分隔符的位置
	protocolEnd := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`).FindStringIndex(rawURL)
	protocolPart := ""
	pathPart := rawURL
	if len(protocolEnd) > 0 {
		protocolPart = rawURL[:protocolEnd[1]]
		pathPart = rawURL[protocolEnd[1]:]
	}
	// 处理路径部分：
	// 1. 将反斜杠替换为正斜杠
	pathPart = strings.ReplaceAll(pathPart, "\\", "/")
	// 2. 将多个连续的斜杠替换为单个斜杠（但保留协议后的双斜杠）
	pathPart = regexp.MustCompile(`/+`).ReplaceAllString(pathPart, "/")
	// 3. 移除末尾的斜杠（除非是根路径）
	if len(pathPart) > 1 && strings.HasSuffix(pathPart, "/") {
		pathPart = strings.TrimRight(pathPart, "/")
	}
	return protocolPart + pathPart + query
}

// NormalizeWithPort 标准化带端口的URL链接
func NormalizeWithPort(rawURL string) string {
	return Normalize(rawURL)
}

// AddScheme 如果URL没有协议，则添加默认的http协议
func AddScheme(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	// 检查是否已经有协议
	if regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(rawURL) {
		return rawURL
	}
	// 如果URL以//开头，添加http协议
	if strings.HasPrefix(rawURL, "//") {
		return "http:" + rawURL
	}
	// 否则添加http://
	return "http://" + rawURL
}

// RemoveScheme 移除URL中的协议部分
func RemoveScheme(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	// 移除协议部分
	schemeRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)
	return schemeRegex.ReplaceAllString(rawURL, "")
}

// GetDomain 获取URL中的域名部分
func GetDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	// 添加协议如果不存在
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(rawURL) {
		rawURL = "http://" + rawURL
	}
	// 解析域名部分
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Host
}

// Encode 对URL进行编码
func Encode(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// 重新组合URL，自动进行编码
	return u.String(), nil
}

// Decode 对URL进行解码
func Decode(encodedURL string) (string, error) {
	decoded, err := url.QueryUnescape(encodedURL)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

// ToURI 转URL或URL字符串为URI
// 该方法会解析并规范化URL，确保返回合法的URI格式
func ToURI(rawURL string) (string, error) {
	if rawURL == "" {
		return "", nil
	}

	// 使用url.Parse解析URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 重新组装URL，这会自动规范化URL格式
	return u.String(), nil
}

// IsValid 检查URL是否有效
func IsValid(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}

// IsAbsolute 检查URL是否为绝对路径
func IsAbsolute(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.IsAbs()
}
