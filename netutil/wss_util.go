package netutil

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/wind959/ko-utils/jsonutil"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ConnPool 连接池
type ConnPool struct {
	pool sync.Pool
}

// NewConnPool 创建一个新的连接池
func NewConnPool() *ConnPool {
	return &ConnPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &websocket.Conn{}
			},
		},
	}
}

// Get 从连接池中获取一个连接
func (p *ConnPool) Get() *websocket.Conn {
	return p.pool.Get().(*websocket.Conn)
}

// Put 将连接放回连接池
func (p *ConnPool) Put(conn *websocket.Conn) {
	p.pool.Put(conn)
}

// WebSocketClient 封装了 WebSocket 客户端的功能
type WebSocketClient struct {
	connPool          *ConnPool // 连接池
	conn              *websocket.Conn
	proxyURL          string        // 代理 socks, http,https
	headers           http.Header   // 请求头
	messageChan       chan []byte   // 消息通道
	onMessage         func([]byte)  // 消息处理回调
	onConnect         func()        // 连接成功回调
	onDisconnect      func()        // 断开连接回调
	onError           func(error)   // 错误处理回调
	reconnect         bool          // 是否自动重连
	maxRetries        int           // 最大重试次数
	reconnectChan     chan struct{} // 重连信号通道
	mu                sync.Mutex    // 互斥锁
	reconnectInterval time.Duration // 重试间隔
	onRetryFailed     func(int)     // 重试失败回调函数
}

// NewWebSocketClient 创建一个新的 WebSocket 客户端
func NewWebSocketClient(proxyURL string, headers http.Header) (*WebSocketClient, error) {
	client := &WebSocketClient{
		connPool:          NewConnPool(), // 初始化连接池
		proxyURL:          proxyURL,
		headers:           headers,
		messageChan:       make(chan []byte, 100),
		reconnectChan:     make(chan struct{}),
		maxRetries:        5,               // 默认最大重试次数
		reconnectInterval: 5 * time.Second, // 默认重试间隔
	}
	return client, nil
}

// Connect 连接到 WebSocket 服务器
func (c *WebSocketClient) Connect(ctx context.Context, wsURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 从连接池中获取连接
	conn := c.connPool.Get()
	defer func() {
		if conn != nil {
			c.connPool.Put(conn)
		}
	}()

	// 创建 WebSocket Dialer
	dialer := websocket.Dialer{
		HandshakeTimeout: 1000 * time.Second,
	}

	// 配置代理
	if c.proxyURL != "" {
		proxyURL, err := url.Parse(c.proxyURL)
		if err != nil {
			return err
		}

		switch proxyURL.Scheme {
		case "http", "https":
			// HTTP/HTTPS 代理
			dialer.Proxy = http.ProxyURL(proxyURL)
		case "socks5":
			// SOCKS5 代理
			dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				auth := &proxy.Auth{}
				if proxyURL.User != nil {
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
			return errors.New("unsupported proxy type")
		}
	}

	// 设置自定义 Header
	header := http.Header{}
	for key, values := range c.headers {
		for _, value := range values {
			header.Add(key, value)
		}
	}

	// 连接到 WebSocket 服务器
	conn, _, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		if c.onError != nil {
			c.onError(err)
		}
		return err
	}

	c.conn = conn
	fmt.Printf("✅ WebSocket Connected to %s\n", wsURL)

	go c.readMessages()

	if c.onConnect != nil {
		c.onConnect()
	}

	if c.reconnect {
		go c.handleReconnect()
	}

	return nil
}

// readMessages 读取 WebSocket 消息
func (c *WebSocketClient) readMessages() {
	defer close(c.messageChan)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if c.onError != nil {
				c.onError(err)
			}
			if c.reconnect {
				c.reconnectChan <- struct{}{}
			}
			// 接连断开，销毁连接
			c.Close()
			return
		}
		if c.onMessage != nil {
			c.onMessage(message)
		} else {
			c.messageChan <- message
		}
	}
}

// SendMessage 发送消息到 WebSocket 服务器
func (c *WebSocketClient) SendMessage(ctx context.Context, message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return errors.New("not connected")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

// SendStrMsg 发送消息到 WebSocket 服务器
func (c *WebSocketClient) SendStrMsg(ctx context.Context, message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return errors.New("not connected")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}

// SendJSON 发送JSON数据到WebSocket服务器
func (c *WebSocketClient) SendJSON(ctx context.Context, v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return errors.New("not connected")
	}
	data, err := jsonutil.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return c.conn.WriteMessage(websocket.TextMessage, []byte(data))
	}
}

// GetMessageChan 获取消息通道
func (c *WebSocketClient) GetMessageChan() <-chan []byte {
	return c.messageChan
}

// Close 关闭 WebSocket 连接
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}
	// 销毁连接
	err := c.conn.Close()
	c.conn = nil
	return err
}

// SetOnMessage 设置消息处理回调
func (c *WebSocketClient) SetOnMessage(handler func([]byte)) {
	c.onMessage = handler
}

// SetOnConnect 设置连接成功回调
func (c *WebSocketClient) SetOnConnect(handler func()) {
	c.onConnect = handler
}

// SetOnDisconnect 设置断开连接回调
func (c *WebSocketClient) SetOnDisconnect(handler func()) {
	c.onDisconnect = handler
}

// SetOnError 设置错误处理回调
func (c *WebSocketClient) SetOnError(handler func(error)) {
	c.onError = handler
}

// SetReconnect 设置是否自动重连
func (c *WebSocketClient) SetReconnect(reconnect bool) {
	c.reconnect = reconnect
}

// SetMaxRetries 设置最大重试次数
func (c *WebSocketClient) SetMaxRetries(maxRetries int) {
	c.maxRetries = maxRetries
}

// SetReconnectInterval 设置重试间隔
func (c *WebSocketClient) SetReconnectInterval(interval time.Duration) {
	c.reconnectInterval = interval
}

// SetOnRetryFailed 设置重试失败回调函数
func (c *WebSocketClient) SetOnRetryFailed(handler func(int)) {
	c.onRetryFailed = handler
}

// handleReconnect 处理自动重连
func (c *WebSocketClient) handleReconnect() {
	retryCount := 0
	for range c.reconnectChan {
		if retryCount >= c.maxRetries {
			if c.onRetryFailed != nil {
				c.onRetryFailed(retryCount)
			}
			return
		}

		time.Sleep(c.reconnectInterval) // 使用用户设置的重试间隔
		if err := c.Connect(context.Background(), c.conn.RemoteAddr().String()); err != nil {
			if c.onError != nil {
				c.onError(err)
			}
			retryCount++
		} else {
			retryCount = 0 // 重连成功后重置重试次数
		}
	}
}
