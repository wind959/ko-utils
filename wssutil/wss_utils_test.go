package wssutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("=== WebSocket客户端示例 ===")

	// 示例1: 基本用法
	basicExample()

	// 示例2: 带配置的用法
	configuredExample()

	// 示例3: 并发用法
	concurrentExample()
}

func basicExample() {
	fmt.Println("\n--- 基本用法示例 ---")

	// 创建最简单的客户端
	client := NewWebSocketClient() // ✅ 零配置

	// 场景2: 带代理的客户端（原来需要 WithProxy）
	//withProxy := NewWebSocketClient(
	//	WithProxy("socks5://127.0.0.1:1080"),
	//)
	// 场景3: 带Header的客户端（原来需要 WithHeaders）
	//withHeaders := NewWebSocketClient(
	//	WithHeader("Authorization", "Bearer token"),
	//)
	// 场景4: 完整配置的客户端
	//fullConfig := NewWebSocketClient(
	//	WithHandshakeTimeout(15*time.Second),
	//	WithBufferSize(8192, 8192),
	//	WithProxy("http://proxy:8080"),
	//	WithHeader("User-Agent", "MyApp/1.0"),
	//	WithCompression(true),
	//	WithSkipVerify(),
	//)

	// 连接（这里使用一个测试服务器地址，实际使用时替换）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 注意: 这里需要一个真实的WebSocket服务器
	err := client.Connect(ctx, "ws://echo.websocket.org")
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer client.Close()

	fmt.Println("基本客户端创建完成")
}

func customer() {
	// 1. 正常使用你的Option
	client := NewWebSocketClient(
		WithHandshakeTimeout(15*time.Second),
		WithBufferSize(8192, 8192),
		WithCompression(true),
	)

	// 2.特殊配置时，直接修改Dialer方法
	dialer := client.Dialer()

	// 设置任何websocket.Dialer支持的字段
	dialer.EnableCompression = false // 覆盖Option的设置
	//dialer.Jar =           // 设置CookieJar
	//dialer.WriteBufferPool = myPool   // 设置缓冲池

	// 甚至完全自定义NetDial
	dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 用户自定义拨号逻辑
		d := &net.Dialer{Timeout: 5 * time.Second}
		return d.DialContext(ctx, network, addr)
	}

	// 3. 连接（使用用户修改后的dialer）
	ctx := context.Background()
	err := client.Connect(ctx, "wss://api.example.com/ws")
	if err != nil {
		// 处理错误
	}
}

func configuredExample() {
	fmt.Println("\n--- 带配置的用法示例 ---")

	// 创建Cookie Jar
	jar, _ := cookiejar.New(nil)

	// 创建带完整配置的客户端
	client := NewWebSocketClient(
		WithHandshakeTimeout(15*time.Second),
		WithBufferSize(8192, 8192),
		WithHeader("Authorization", "Bearer token-12345"),
		WithHeader("User-Agent", "MyApp/1.0"),
		WithSubprotocols("chat", "json"),
		WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}),
		WithCompression(true),
		WithCookieJar(jar),
		WithTimeouts(
			15*time.Second, // 握手超时
			30*time.Second, // 读取超时
			10*time.Second, // 写入超时
		),
	)

	fmt.Printf("客户端配置:\n")
	fmt.Printf("  - Handshake Timeout: %v\n", client.Config().HandshakeTimeout)
	fmt.Printf("  - Read Buffer Size: %d\n", client.Config().ReadBufferSize)
	fmt.Printf("  - Enable Compression: %v\n", client.Config().EnableCompression)
	fmt.Printf("  - Headers Count: %d\n", len(client.Config().Headers))
}

func concurrentExample() {
	fmt.Println("\n--- 并发用法示例 ---")

	// 创建多个客户端（模拟多个连接）
	var wg sync.WaitGroup
	clients := make([]*WebSocketClient, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 每个客户端有自己的配置
			client := NewWebSocketClient(
				WithHeader("Client-ID", fmt.Sprintf("client-%d", idx)),
				WithHandshakeTimeout(time.Duration(5+idx)*time.Second),
			)

			clients[idx] = client

			// 这里可以执行连接和通信
			fmt.Printf("客户端 %d 创建完成\n", idx)
		}(i)
	}

	wg.Wait()

	// 清理
	for _, client := range clients {
		if client != nil {
			client.Close()
		}
	}

	fmt.Println("所有客户端已关闭")
}

// 实际使用场景示例
func realWorldExample() {
	fmt.Println("\n--- 实际使用场景示例 ---")

	// 处理中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 创建客户端
	client := NewWebSocketClient(
		WithHandshakeTimeout(10*time.Second),
		WithHeader("X-API-Key", "your-api-key-here"),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 连接服务器（示例URL）
	wsURL := "wss://api.example.com/realtime"
	err := client.Connect(ctx, wsURL)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()

	fmt.Printf("已连接到: %s\n", client.URL())

	// 启动读取goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)

		for client.IsConnected() {
			msg, err := client.ReadMessageText()
			if err != nil {
				log.Printf("读取错误: %v", err)
				return
			}

			// 处理消息
			fmt.Printf("收到消息: %s\n", msg)

			// 可以根据消息类型进行不同处理
			// processMessage(msg)
		}
	}()

	// 启动写入goroutine（例如：发送心跳）
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if client.IsConnected() {
					// 发送心跳
					err := client.WriteText(`{"type":"ping","timestamp":` +
						fmt.Sprintf("%d", time.Now().Unix()) + `}`)
					if err != nil {
						log.Printf("心跳发送失败: %v", err)
						return
					}
				}
			case <-done:
				return
			case <-interrupt:
				return
			}
		}
	}()

	// 发送一些初始消息
	messages := []string{
		`{"type":"subscribe","channel":"ticker"}`,
		`{"type":"subscribe","channel":"orderbook"}`,
	}

	for _, msg := range messages {
		if err := client.WriteText(msg); err != nil {
			log.Printf("发送失败: %v", err)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 等待中断
	select {
	case <-interrupt:
		fmt.Println("\n收到中断信号，关闭连接...")
	case <-done:
		fmt.Println("连接已关闭")
	}

	// 发送关闭帧
	if client.IsConnected() {
		client.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		time.Sleep(100 * time.Millisecond)
	}
}

// 代理使用示例
func proxyExample() {
	fmt.Println("\n--- 代理使用示例 ---")

	// HTTP代理
	httpClient := NewWebSocketClient(
		WithProxy("http://proxy.example.com:8080"),
		WithProxyAuth("username", "password"),
	)
	fmt.Println("HTTP代理客户端创建完成")

	// SOCKS5代理
	socksClient := NewWebSocketClient(
		WithProxy("socks5://127.0.0.1:1080"),
	)
	fmt.Println("SOCKS5代理客户端创建完成")

	// 注意: 实际使用时需要取消注释连接代码
	_ = httpClient
	_ = socksClient
}

func init() {
	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 运行所有示例
func Example_main() {
	basicExample()
	configuredExample()
	concurrentExample()
	proxyExample()

	// 以运行实际示例（需要真实服务器）
	realWorldExample()
}
