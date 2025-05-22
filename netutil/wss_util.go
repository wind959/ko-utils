package netutil

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/wind959/ko-utils/jsonutil"
	logutil "github.com/wind959/ko-utils/logger"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// WebSocketClient 封装了 WebSocket 客户端的功能
type WebSocketClient struct {
	conn              *websocket.Conn
	wsURL             string        // WebSocket URL
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
func NewWebSocketClient(proxyURL string, headers http.Header) *WebSocketClient {
	return &WebSocketClient{
		proxyURL:          proxyURL,
		headers:           headers,
		messageChan:       make(chan []byte, 100),
		reconnectChan:     make(chan struct{}),
		maxRetries:        5,               // 默认最大重试次数
		reconnectInterval: 5 * time.Second, // 默认重试间隔
	}
}

// Connect 连接到 WebSocket 服务器
func (c *WebSocketClient) Connect(ctx context.Context, wsURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.wsURL = wsURL
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
	conn, resp, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			logutil.Error("❌ 握手失败", zap.String("status", resp.Status), zap.ByteString("body", body))
		}
		if c.onError != nil {
			c.onError(err)
		}
		return err
	}

	c.conn = conn
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
	defer func() {
		if c.reconnect {
			c.mu.Lock()
			c.conn = nil
			c.mu.Unlock()
			c.reconnectChan <- struct{}{}
		}
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if c.onError != nil {
				c.onError(err)
			}
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
	return c.SendMessage(ctx, []byte(message))
}

// SendJSON 发送JSON数据到WebSocket服务器
func (c *WebSocketClient) SendJSON(ctx context.Context, v interface{}) error {
	data, err := jsonutil.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendMessage(ctx, []byte(data))

}

// GetMessageChan 获取消息通道
func (c *WebSocketClient) GetMessageChan() <-chan []byte {
	return c.messageChan
}

// Close 关闭 WebSocket 连接
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
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

		if err := c.Connect(context.Background(), c.wsURL); err != nil {
			retryCount++
			if c.onError != nil {
				c.onError(fmt.Errorf("重连失败 (第 %d 次)：%w", retryCount, err))
			}
		} else {
			logutil.Info("🔄 WebSocket 重连成功", zap.Int("retryCount", retryCount+1))
			retryCount = 0 // 成功后重置
			if c.onConnect != nil {
				c.onConnect()
			}
		}
	}
}
