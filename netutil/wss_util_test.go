package netutil

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestWebSocketClient(t *testing.T) {
	// 创建 WebSocket 客户端
	client := NewWebSocketClient("", http.Header{})
	// 设置消息处理回调
	client.SetOnMessage(func(message []byte) {
		fmt.Println("Received message:", string(message))
	})

	// 设置连接成功回调
	client.SetOnConnect(func() {
		fmt.Println("Connected to WebSocket server")
	})

	// 设置错误处理回调
	client.SetOnError(func(err error) {
		fmt.Println("WebSocket error:", err)
	})

	// 连接到 WebSocket 服务器
	ctx := context.Background()
	err := client.Connect(ctx, "ws://echo.websocket.org")
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}

	// 发送消息
	err = client.SendMessage(ctx, []byte("Hello, WebSocket!"))
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}

	// 等待一段时间以接收消息
	time.Sleep(5 * time.Second)

	// 关闭连接
	err = client.Close()
	if err != nil {
		fmt.Println("Failed to close connection:", err)
		return
	}
}
