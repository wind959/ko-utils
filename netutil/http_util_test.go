package netutil

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

func demo1() {
	// 创建 HTTP 客户端
	client := NewHttpClient(&HttpClientConfig{
		Timeout:    10 * time.Second,
		RetryCount: 3,
		DefaultHeaders: map[string]string{
			"User-Agent": "MyApp/1.0",
		},
	})

	// 发送 GET 请求
	ctx := context.Background()
	resp, err := client.Get(ctx, "https://example.com", nil)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	fmt.Println("Response:", resp.String())
}

// 异步请求
func demo2() {
	// 创建 HTTP 客户端
	client := NewHttpClient(&HttpClientConfig{
		Timeout:    10 * time.Second,
		RetryCount: 3,
		DefaultHeaders: map[string]string{
			"User-Agent": "MyApp/1.0",
		},
	})

	// 异步发送 GET 请求
	client.GetAsync(context.Background(), "https://example.com", nil, func(resp *resty.Response, err error) {
		if err != nil {
			fmt.Println("Async GET request failed:", err)
			return
		}
		fmt.Println("Async GET response:", resp.String())
	})

	// 异步发送 POST 请求
	client.PostJsonAsync(context.Background(), "https://example.com", map[string]string{"key": "value"}, nil, func(resp *resty.Response, err error) {
		if err != nil {
			fmt.Println("Async POST request failed:", err)
			return
		}
		fmt.Println("Async POST response:", resp.String())
	})

	// 等待异步请求完成
	time.Sleep(5 * time.Second)
}
