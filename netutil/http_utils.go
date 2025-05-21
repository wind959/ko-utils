package netutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/wind959/ko-utils/strutil"
	"golang.org/x/net/http2"
	"net/http"
	"time"
)

// HttpClientConfig HTTP 客户端配置
type HttpClientConfig struct {
	Timeout           time.Duration     // 请求超时时间
	RetryCount        int               // 重试次数
	Proxy             string            // 代理地址
	TLSConfig         *tls.Config       // TLS 配置
	MaxHeaderListSize uint32            // HTTP/2 最大头部大小
	DefaultHeaders    map[string]string // 默认请求头
}

// HttpClient HTTP 客户端
type HttpClient struct {
	client *resty.Client
	config *HttpClientConfig
}

// NewHttpClient 创建 HTTP 客户端
func NewHttpClient(config *HttpClientConfig) *HttpClient {
	client := resty.New()
	if config.Timeout > 0 {
		client.SetTimeout(config.Timeout)
	}

	// 设置重试次数
	if config.RetryCount > 0 {
		client.SetRetryCount(config.RetryCount)
	}

	// 配置 TLS
	tlsConfig := config.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // 默认跳过 TLS 验证
		}
	}

	// 在 Transport 中启用连接池
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: 60 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}

	if config.MaxHeaderListSize <= 0 {
		// 过 Cloudflare
		config.MaxHeaderListSize = 262144 // 256 KB
	}

	// 创建 HTTP/2 Transport
	h2Transport := &http2.Transport{
		TLSClientConfig:   tlsConfig,
		MaxHeaderListSize: config.MaxHeaderListSize,
	}

	// 让 http.Transport 处理 HTTP/1.1，让 http2.Transport 处理 HTTP/2
	transport.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{
		"h2": func(authority string, c *tls.Conn) http.RoundTripper {
			return h2Transport
		},
	}
	client.SetTransport(transport)

	// 设置代码
	if strutil.IsNotBlank(config.Proxy) {
		client.SetProxy(config.Proxy)
	}

	// 设置默认请求头
	if len(config.DefaultHeaders) > 0 {
		client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
			for k, v := range config.DefaultHeaders {
				if _, exists := req.Header[k]; !exists {
					req.SetHeader(k, v)
				}
			}
			return nil
		})
	}
	return &HttpClient{
		client: client,
		config: config,
	}
}

// SetTimeout 动态设置超时时间
func (hc *HttpClient) SetTimeout(timeout time.Duration) {
	hc.client.SetTimeout(timeout)
}

// SetProxy 动态设置代理
func (hc *HttpClient) SetProxy(proxy string) {
	hc.client.SetProxy(proxy)
}

// AddRequestMiddleware 添加请求中间件
func (hc *HttpClient) AddRequestMiddleware(middleware func(*resty.Client, *resty.Request) error) {
	hc.client.OnBeforeRequest(middleware)
}

// AddResponseMiddleware 添加响应中间件
func (hc *HttpClient) AddResponseMiddleware(middleware func(*resty.Client, *resty.Response) error) {
	hc.client.OnAfterResponse(middleware)
}

// SetDynamicHeaders 动态设置请求头
func (hc *HttpClient) SetDynamicHeaders(headers map[string]string) {
	hc.AddRequestMiddleware(func(c *resty.Client, req *resty.Request) error {
		for k, v := range headers {
			if _, exists := req.Header[k]; !exists {
				req.SetHeader(k, v)
			}
		}
		return nil
	})
}

// Get 发送 GET 请求
func (hc *HttpClient) Get(ctx context.Context, url string, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// PostJson 发送 JSON 格式的 POST 请求
func (hc *HttpClient) PostJson(ctx context.Context, url string, payload interface{}, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx).SetBody(payload)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("POST request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// PostForm 发送表单格式的 POST 请求
func (hc *HttpClient) PostForm(ctx context.Context, url string, formData map[string]string, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx).SetFormData(formData)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("POST form request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// PutJson 发送 JSON 格式的 PUT 请求
func (hc *HttpClient) PutJson(ctx context.Context, url string, payload interface{}, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx).SetBody(payload)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Put(url)
	if err != nil {
		return nil, fmt.Errorf("PUT request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// Delete 发送 DELETE 请求
func (hc *HttpClient) Delete(ctx context.Context, url string, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Delete(url)
	if err != nil {
		return nil, fmt.Errorf("DELETE request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// PatchJson 发送 JSON 格式的 PATCH 请求
func (hc *HttpClient) PatchJson(ctx context.Context, url string, payload interface{}, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx).SetBody(payload)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Patch(url)
	if err != nil {
		return nil, fmt.Errorf("PATCH request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// Head 发送 HEAD 请求
func (hc *HttpClient) Head(ctx context.Context, url string, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Head(url)
	if err != nil {
		return nil, fmt.Errorf("HEAD request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// Options 发送 OPTIONS 请求
func (hc *HttpClient) Options(ctx context.Context, url string, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx)
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Options(url)
	if err != nil {
		return nil, fmt.Errorf("OPTIONS request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// PostXml 发送 XML 格式的 POST 请求
func (hc *HttpClient) PostXml(ctx context.Context, url string, payload interface{}, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx).SetBody(payload).SetHeader("Content-Type", "application/xml")
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("POST XML request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// PostMultipart 发送 multipart 格式的 POST 请求
func (hc *HttpClient) PostMultipart(ctx context.Context, url string, formData map[string]string, headers map[string]string) (*resty.Response, error) {
	request := hc.client.R().SetContext(ctx).SetFormData(formData).SetHeader("Content-Type", "multipart/form-data")
	if headers != nil {
		request.SetHeaders(headers)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("POST multipart request url: %s failed: %v", url, err)
	}
	return resp, nil
}

// ================= 以下是 异步 请求 =========================//

// GetAsync 异步发送 GET 请求
func (hc *HttpClient) GetAsync(ctx context.Context, url string, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.Get(ctx, url, headers)
		callback(resp, err)
	}()
}

// PostJsonAsync 异步发送 JSON 格式的 POST 请求
func (hc *HttpClient) PostJsonAsync(ctx context.Context, url string, payload interface{}, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.PostJson(ctx, url, payload, headers)
		callback(resp, err)
	}()
}

// PostFormAsync 异步发送表单格式的 POST 请求
func (hc *HttpClient) PostFormAsync(ctx context.Context, url string, formData map[string]string, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.PostForm(ctx, url, formData, headers)
		callback(resp, err)
	}()
}

// PutJsonAsync 异步发送 JSON 格式的 PUT 请求
func (hc *HttpClient) PutJsonAsync(ctx context.Context, url string, payload interface{}, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.PutJson(ctx, url, payload, headers)
		callback(resp, err)
	}()
}

// DeleteAsync 异步发送 DELETE 请求
func (hc *HttpClient) DeleteAsync(ctx context.Context, url string, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.Delete(ctx, url, headers)
		callback(resp, err)
	}()
}

// PatchJsonAsync 异步发送 JSON 格式的 PATCH 请求
func (hc *HttpClient) PatchJsonAsync(ctx context.Context, url string, payload interface{}, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.PatchJson(ctx, url, payload, headers)
		callback(resp, err)
	}()
}

// HeadAsync 异步发送 HEAD 请求
func (hc *HttpClient) HeadAsync(ctx context.Context, url string, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.Head(ctx, url, headers)
		callback(resp, err)
	}()
}

// OptionsAsync 异步发送 OPTIONS 请求
func (hc *HttpClient) OptionsAsync(ctx context.Context, url string, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.Options(ctx, url, headers)
		callback(resp, err)
	}()
}

// PostXmlAsync 异步发送 XML 格式的 POST 请求
func (hc *HttpClient) PostXmlAsync(ctx context.Context, url string, payload interface{}, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.PostXml(ctx, url, payload, headers)
		callback(resp, err)
	}()
}

// PostMultipartAsync 异步发送 multipart 格式的 POST 请求
func (hc *HttpClient) PostMultipartAsync(ctx context.Context, url string, formData map[string]string, headers map[string]string, callback func(*resty.Response, error)) {
	go func() {
		resp, err := hc.PostMultipart(ctx, url, formData, headers)
		callback(resp, err)
	}()
}
