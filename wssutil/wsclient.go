package wssutil

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wind959/ko-utils/jsonutil"
	"golang.org/x/net/proxy"
)

// ClientConfig WebSocket客户端配置
type ClientConfig struct {
	// 连接超时
	HandshakeTimeout time.Duration // 握手超时
	ReadTimeout      time.Duration // 读取超时
	WriteTimeout     time.Duration // 写入超时

	ReadBufferSize  int // 读缓冲区大小
	WriteBufferSize int // 写缓冲区大小

	Subprotocols []string // 子协议

	TLSConfig *tls.Config // TLS配置

	ProxyURL  string     // 代理地址
	ProxyAuth *ProxyAuth // 代理认证

	Headers http.Header    // 自定义请求头
	Jar     http.CookieJar // Cookie管理

	EnableCompression bool // 启用压缩
}

// ProxyAuth 代理认证信息
type ProxyAuth struct {
	Username string
	Password string
}

// DefaultConfig 默认配置
var DefaultConfig = &ClientConfig{
	HandshakeTimeout:  10 * time.Second,
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	Headers:           make(http.Header),
	EnableCompression: true,
}

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	conn   *websocket.Conn
	config *ClientConfig
	wsURL  string
	dialer *websocket.Dialer // 保存用户修改后的dialer
}

// Dialer 返回当前的Dialer实例
func (c *WebSocketClient) Dialer() *websocket.Dialer {
	if c.dialer == nil {
		// 基于config创建默认dialer
		c.dialer = &websocket.Dialer{
			HandshakeTimeout:  c.config.HandshakeTimeout,
			ReadBufferSize:    c.config.ReadBufferSize,
			WriteBufferSize:   c.config.WriteBufferSize,
			Subprotocols:      c.config.Subprotocols,
			TLSClientConfig:   c.config.TLSConfig,
			EnableCompression: c.config.EnableCompression,
			Jar:               c.config.Jar,
		}
	}
	// 注意：这个方法返回的dialer在Connect()时使用
	// 修改dialer字段会影响后续Connect()调用
	// 线程不安全，请在Connect()前修改
	return c.dialer
}

// ClientOption 选项模式
type ClientOption func(*ClientConfig)

// NewWebSocketClient 创建WebSocket客户端
func NewWebSocketClient(opts ...ClientOption) *WebSocketClient {
	config := &ClientConfig{
		HandshakeTimeout:  DefaultConfig.HandshakeTimeout,
		ReadBufferSize:    DefaultConfig.ReadBufferSize,
		WriteBufferSize:   DefaultConfig.WriteBufferSize,
		Headers:           make(http.Header),
		EnableCompression: DefaultConfig.EnableCompression,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	return &WebSocketClient{
		config: config,
	}
}

// WithHandshakeTimeout 设置握手超时
func WithHandshakeTimeout(d time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.HandshakeTimeout = d
	}
}

// WithTimeouts 设置所有超时
func WithTimeouts(handshake, read, write time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.HandshakeTimeout = handshake
		c.ReadTimeout = read
		c.WriteTimeout = write
	}
}

// WithBufferSize 设置缓冲区大小
func WithBufferSize(readBuf, writeBuf int) ClientOption {
	return func(c *ClientConfig) {
		c.ReadBufferSize = readBuf
		c.WriteBufferSize = writeBuf
	}
}

// WithProxy 设置代理
func WithProxy(proxyURL string) ClientOption {
	return func(c *ClientConfig) {
		c.ProxyURL = proxyURL
	}
}

// WithProxyAuth 设置代理认证
func WithProxyAuth(username, password string) ClientOption {
	return func(c *ClientConfig) {
		c.ProxyAuth = &ProxyAuth{
			Username: username,
			Password: password,
		}
	}
}

// WithHeaders 设置HTTP头
func WithHeaders(headers http.Header) ClientOption {
	return func(c *ClientConfig) {
		if headers != nil {
			c.Headers = headers
		}
	}
}

// WithHeader 设置单个HTTP头
func WithHeader(key, value string) ClientOption {
	return func(c *ClientConfig) {
		c.Headers.Add(key, value)
	}
}

// WithSubprotocols 设置子协议
func WithSubprotocols(protocols ...string) ClientOption {
	return func(c *ClientConfig) {
		c.Subprotocols = protocols
	}
}

// WithTLSConfig 设置TLS配置
func WithTLSConfig(tlsConfig *tls.Config) ClientOption {
	return func(c *ClientConfig) {
		c.TLSConfig = tlsConfig
	}
}

// WithCompression 设置压缩
func WithCompression(enable bool) ClientOption {
	return func(c *ClientConfig) {
		c.EnableCompression = enable
	}
}

// WithCookieJar 设置Cookie管理
func WithCookieJar(jar http.CookieJar) ClientOption {
	return func(c *ClientConfig) {
		c.Jar = jar
	}
}

// WithSkipVerify 跳过TLS验证（用于测试）
func WithSkipVerify() ClientOption {
	return func(c *ClientConfig) {
		if c.TLSConfig == nil {
			c.TLSConfig = &tls.Config{}
		}
		c.TLSConfig.InsecureSkipVerify = true
	}
}

// Connect 建立WebSocket连接（同步阻塞）
func (c *WebSocketClient) Connect(ctx context.Context, wsURL string) error {
	c.wsURL = wsURL

	dialer := c.Dialer()

	userHasProxyConfig := dialer.Proxy != nil ||
		dialer.NetDial != nil ||
		dialer.NetDialContext != nil ||
		dialer.NetDialTLSContext != nil

	if c.config.ProxyURL != "" && !userHasProxyConfig {
		if err := c.configureProxy(dialer); err != nil {
			return err
		}
	}
	// 发起连接
	conn, resp, err := dialer.DialContext(ctx, wsURL, c.config.Headers)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return errors.New("websocket handshake failed: " + resp.Status + ", body: " + string(body))
		}
		return err
	}
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	c.conn = conn
	// 设置超时（如果配置了）
	if c.config.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
	}
	if c.config.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	}
	return nil
}

// configureProxy 配置代理
func (c *WebSocketClient) configureProxy(dialer *websocket.Dialer) error {
	proxyURL, err := url.Parse(c.config.ProxyURL)
	if err != nil {
		return err
	}

	switch proxyURL.Scheme {
	case "http", "https":
		// HTTP/HTTPS代理
		if c.config.ProxyAuth != nil {
			proxyURL.User = url.UserPassword(
				c.config.ProxyAuth.Username,
				c.config.ProxyAuth.Password,
			)
		}
		dialer.Proxy = http.ProxyURL(proxyURL)

	case "socks5", "socks5h":
		// SOCKS5代理
		dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			auth := &proxy.Auth{}
			if c.config.ProxyAuth != nil {
				auth.User = c.config.ProxyAuth.Username
				auth.Password = c.config.ProxyAuth.Password
			} else if proxyURL.User != nil {
				auth.User = proxyURL.User.Username()
				auth.Password, _ = proxyURL.User.Password()
			}

			socksDialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
			if err != nil {
				return nil, err
			}
			return socksDialer.Dial(network, addr)
		}

	default:
		return errors.New("unsupported proxy type: " + proxyURL.Scheme)
	}

	return nil
}

// ReadMessage 读取一条消息（同步阻塞）
func (c *WebSocketClient) ReadMessage() (messageType int, data []byte, err error) {
	if c.conn == nil {
		return 0, nil, errors.New("not connected")
	}
	if c.config.ReadTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
	}

	return c.conn.ReadMessage()
}

// ReadMessageText 读取文本消息
func (c *WebSocketClient) ReadMessageText() (string, error) {
	msgType, data, err := c.ReadMessage()
	if err != nil {
		return "", err
	}
	if msgType != websocket.TextMessage {
		return "", errors.New("not a text message")
	}
	return string(data), nil
}

// WriteMessage 发送消息
func (c *WebSocketClient) WriteMessage(messageType int, data []byte) error {
	if c.conn == nil {
		return errors.New("not connected")
	}

	if c.config.WriteTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	}

	return c.conn.WriteMessage(messageType, data)
}

// WriteText 发送文本消息
func (c *WebSocketClient) WriteText(text string) error {
	return c.WriteMessage(websocket.TextMessage, []byte(text))
}

func (c *WebSocketClient) WriteBinary(data []byte) error {
	return c.WriteMessage(websocket.BinaryMessage, data)
}

// WriteJSON 发送JSON数据到WebSocket服务器
func (c *WebSocketClient) WriteJSON(v interface{}) error {
	if c.conn == nil {
		return errors.New("not connected")
	}

	if c.config.WriteTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	}

	return c.conn.WriteJSON(v)
}

// WriteJSONWithUtil 使用jsonutil写入JSON
func (c *WebSocketClient) WriteJSONWithUtil(v interface{}) error {
	data, err := jsonutil.Marshal(v)
	if err != nil {
		return err
	}
	return c.WriteText(data)
}

// SetReadDeadline 设置读取截止时间
func (c *WebSocketClient) SetReadDeadline(t time.Time) error {
	if c.conn == nil {
		return errors.New("not connected")
	}
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline 设置写入截止时间
func (c *WebSocketClient) SetWriteDeadline(t time.Time) error {
	if c.conn == nil {
		return errors.New("not connected")
	}
	return c.conn.SetWriteDeadline(t)
}

// SetPongHandler 设置Pong处理器
func (c *WebSocketClient) SetPongHandler(h func(string) error) {
	if c.conn != nil {
		c.conn.SetPongHandler(h)
	}
}

// SetPingHandler 设置Ping处理器
func (c *WebSocketClient) SetPingHandler(h func(string) error) {
	if c.conn != nil {
		c.conn.SetPingHandler(h)
	}
}

// Close 关闭连接
func (c *WebSocketClient) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// RawConn 获取原始连接（供高级用户使用）
func (c *WebSocketClient) RawConn() *websocket.Conn {
	return c.conn
}

// IsConnected 检查是否连接
func (c *WebSocketClient) IsConnected() bool {
	return c.conn != nil
}

// Config 获取当前配置（只读）
func (c *WebSocketClient) Config() ClientConfig {
	return *c.config
}

// URL 获取连接URL
func (c *WebSocketClient) URL() string {
	return c.wsURL
}
