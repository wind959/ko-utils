package netutil

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/http2"
)

type HttpClient struct {
	*resty.Client
}

type HttpClientConfig struct {
	Timeout               time.Duration     // 请求超时时间
	RetryCount            int               // 重试次数
	Proxy                 string            // 代理地址
	EnableCustomTransport bool              // 是否启用自定义 Transport（默认 false）
	TLSConfig             *tls.Config       // TLS 配置,仅在 EnableCustomTransport=true 时生效
	MaxHeaderListSize     uint32            // HTTP/2 Header 最大大小
	DefaultHeaders        map[string]string // 默认请求头
}

// DefaultHttpClientConfig 默认 HTTP 配置
var DefaultHttpClientConfig = &HttpClientConfig{
	Timeout:               30 * time.Second,
	RetryCount:            5,
	EnableCustomTransport: false,     // 默认禁用自定义 Transport
	MaxHeaderListSize:     256 << 10, // 256KB （默认值）, 过Cloudflare
	TLSConfig: &tls.Config{
		InsecureSkipVerify: false,
	},
}

// NewHttpClient 创建一个 resty 客户端
func NewHttpClient(cfg *HttpClientConfig) *HttpClient {
	// 合并配置
	config := mergeConfig(DefaultHttpClientConfig, cfg)

	client := resty.New()

	// 基础配置
	client.SetTimeout(config.Timeout)
	client.SetRetryCount(config.RetryCount)

	// Proxy
	if config.Proxy != "" {
		client.SetProxy(config.Proxy)
	}

	// ===== 自定义 Transport（可选）=====
	if config.EnableCustomTransport {
		tlsCfg := config.TLSConfig
		if tlsCfg == nil {
			tlsCfg = &tls.Config{}
		}
		transport := &http.Transport{
			TLSClientConfig:     tlsCfg,
			TLSHandshakeTimeout: 60 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		}

		h2Transport := &http2.Transport{
			TLSClientConfig:   tlsCfg,
			MaxHeaderListSize: config.MaxHeaderListSize,
		}

		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{
			"h2": func(authority string, c *tls.Conn) http.RoundTripper {
				return h2Transport
			},
		}

		client.SetTransport(transport)
	}

	// 默认请求头（不覆盖用户显式设置）
	if len(config.DefaultHeaders) > 0 {
		headers := config.DefaultHeaders
		client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			for k, v := range headers {
				if _, ok := r.Header[k]; !ok {
					r.SetHeader(k, v)
				}
			}
			return nil
		})
	}
	return &HttpClient{Client: client}
}

// R 创建一个 resty 请求
func (c *HttpClient) R(ctx context.Context) *resty.Request {
	return c.Client.R().SetContext(ctx)
}

// Get 发送 GET 请求
func (c *HttpClient) Get(ctx context.Context, url string, out any) (*resty.Response, error) {
	return c.R(ctx).
		SetResult(out).
		Get(url)
}

// Post 发送 POST 请求
func (c *HttpClient) Post(ctx context.Context, url string, body any, out any) (*resty.Response, error) {
	return c.R(ctx).
		SetBody(body).
		SetResult(out).
		Post(url)
}

// Put 发送 PUT 请求
func (c *HttpClient) Put(ctx context.Context, url string, body any, out any) (*resty.Response, error) {
	return c.R(ctx).
		SetBody(body).
		SetResult(out).
		Put(url)
}

// Delete 发送 DELETE 请求
func (c *HttpClient) Delete(ctx context.Context, url string, out any) (*resty.Response, error) {
	return c.R(ctx).
		SetResult(out).
		Delete(url)
}

func (c *HttpClient) Patch(ctx context.Context, url string, body any, out any) (*resty.Response, error) {
	return c.R(ctx).SetBody(body).SetResult(out).Patch(url)
}

// Head 发送 HEAD 请求
func (c *HttpClient) Head(ctx context.Context, url string) (*resty.Response, error) {
	return c.R(ctx).Head(url)
}

// Options 发送 OPTIONS 请求
func (c *HttpClient) Options(ctx context.Context, url string) (*resty.Response, error) {
	return c.R(ctx).Options(url)
}

// Do 发送自定义请求
func (c *HttpClient) Do(
	ctx context.Context,
	method string,
	url string,
	body any,
	out any,
) (*resty.Response, error) {

	r := c.R(ctx)

	if body != nil {
		r.SetBody(body)
	}
	if out != nil {
		r.SetResult(out)
	}

	return r.Execute(method, url)
}

// SetTimeout 设置请求超时时间
func (c *HttpClient) SetTimeout(timeout time.Duration) {
	c.Client.SetTimeout(timeout)
}

// SetProxy 动态设置代理
func (c *HttpClient) SetProxy(proxy string) {
	c.Client.SetProxy(proxy)
}

// AddRequestMiddleware 添加请求中间件
func (c *HttpClient) AddRequestMiddleware(middleware func(*resty.Client, *resty.Request) error) {
	c.Client.OnBeforeRequest(middleware)
}

// AddResponseMiddleware 添加响应中间件
func (c *HttpClient) AddResponseMiddleware(middleware func(*resty.Client, *resty.Response) error) {
	c.Client.OnAfterResponse(middleware)
}

// SetHeader 设置请求头
func (c *HttpClient) SetHeader(key, value string) {
	c.Client.SetHeader(key, value)
}

// SetDynamicHeaders 在每次请求前动态注入 Header
func (c *HttpClient) SetDynamicHeaders(
	fn func(ctx context.Context) map[string]string,
) {
	c.AddRequestMiddleware(func(_ *resty.Client, req *resty.Request) error {
		headers := fn(req.Context())
		if len(headers) == 0 {
			return nil
		}
		for k, v := range headers {
			if _, exists := req.Header[k]; !exists {
				req.SetHeader(k, v)
			}
		}
		return nil
	})
}

func mergeConfig(base, override *HttpClientConfig) *HttpClientConfig {
	cfg := *base

	// 深拷贝 headers
	if len(base.DefaultHeaders) > 0 {
		cfg.DefaultHeaders = make(map[string]string, len(base.DefaultHeaders))
		for k, v := range base.DefaultHeaders {
			cfg.DefaultHeaders[k] = v
		}
	}

	if override == nil {
		return &cfg
	}

	if override.Timeout > 0 {
		cfg.Timeout = override.Timeout
	}
	if override.RetryCount > 0 {
		cfg.RetryCount = override.RetryCount
	}
	if override.Proxy != "" {
		cfg.Proxy = override.Proxy
	}

	cfg.EnableCustomTransport = override.EnableCustomTransport

	if override.TLSConfig != nil {
		cfg.TLSConfig = override.TLSConfig
	}
	if override.MaxHeaderListSize > 0 {
		cfg.MaxHeaderListSize = override.MaxHeaderListSize
	}
	if len(override.DefaultHeaders) > 0 {
		if cfg.DefaultHeaders == nil {
			cfg.DefaultHeaders = make(map[string]string)
		}
		for k, v := range override.DefaultHeaders {
			cfg.DefaultHeaders[k] = v
		}
	}

	return &cfg
}
