package puter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"puter2api/internal/types"
)

const (
	apiURL       = "https://api.puter.com/drivers/call"
	defaultModel = "claude-opus-4-5-20251101"
)

// Client Puter API 客户端
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// NewClient 创建新的客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	}
}

// Call 调用 Puter API 并返回完整响应文本
func (c *Client) Call(messages []types.PuterMessage, authToken string) (string, error) {
	puterReq := types.PuterRequest{
		Interface: "puter-chat-completion",
		Driver:    "claude",
		TestMode:  false,
		Method:    "complete",
		Args: types.PuterArgs{
			Messages: messages,
			Model:    defaultModel,
			Stream:   true,
		},
		AuthToken: authToken,
	}

	body, _ := json.Marshal(puterReq)
	startTime := time.Now()
	log.Printf("[Puter] 开始请求, messages=%d", len(messages))

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("[Puter] 创建请求失败: %v", err)
		return "", err
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[Puter] 请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[Puter] API 错误: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("puter API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	// 收集完整响应
	var fullText strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var chunk types.PuterStreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err == nil && chunk.Text != "" {
			fullText.WriteString(chunk.Text)
		}
	}

	responseText := fullText.String()
	elapsed := time.Since(startTime)
	log.Printf("[Puter] 请求完成, 耗时: %v, 响应: %d 字符", elapsed, len(responseText))

	return responseText, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "text/plain;actually=json")
	req.Header.Set("DNT", "1")
	req.Header.Set("Origin", "https://docs.puter.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://docs.puter.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-ch-ua-full-version-list", `"Chromium";v="142.0.7355.4", "Google Chrome";v="142.0.7355.4", "Not_A Brand";v="99.0.0.0"`)
	req.Header.Set("sec-ch-ua-arch", `"arm"`)
	req.Header.Set("sec-ch-ua-bitness", `"64"`)
	req.Header.Set("sec-ch-ua-model", `""`)
	req.Header.Set("sec-ch-ua-platform-version", `"15.0.0"`)
}
