package netutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	cli := NewHttpClient(nil)

	var resp map[string]any
	_, err := cli.Get(context.Background(), "https://api.ipify.org?format=json", &resp)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	fmt.Println("Response:", resp)
}

func Test2(t *testing.T) {
	cli := NewHttpClient(&HttpClientConfig{
		DefaultHeaders: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36",
		},
	})
	var resp map[string]any
	_, err := cli.Get(context.Background(), "https://api.ipify.org?format=json", &resp)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	fmt.Println("Response:", resp)
}

func Test3(t *testing.T) {
	cli := NewHttpClient(nil)
	cli.SetDynamicHeaders(func(ctx context.Context) map[string]string {
		token, _ := ctx.Value("token").(string)
		if token == "" {
			return nil
		}
		return map[string]string{
			"Authorization": "Bearer " + token,
		}
	})
	var resp map[string]any
	ctx := context.WithValue(context.Background(), "token", "abc123")
	cli.Get(ctx, "https://api.ipify.org?format=json", &resp)
}

// Test4 每个请求一个 Trace ID
func Test4(t *testing.T) {

	cli := NewHttpClient(nil)
	cli.SetDynamicHeaders(func(ctx context.Context) map[string]string {
		return map[string]string{
			"X-Request-ID": "aaaaa",
		}
	})
}

// Test5 过Cloudflare / 大 Header / 内网证书
func Test5(t *testing.T) {
	cli := NewHttpClient(&HttpClientConfig{
		EnableCustomTransport: true,
		MaxHeaderListSize:     256 << 10,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true, // 明确表达
		},
	})
	// 测试过Cloudflare
	var resp map[string]any
	_, err := cli.Get(context.Background(), "https://api.ipify.org?format=json", &resp)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	fmt.Println("Response:", resp)
}
