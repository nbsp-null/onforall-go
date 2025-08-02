package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// Client HTTP 客户端
type Client struct {
	timeout     time.Duration
	userAgent   string
	httpClient  *fasthttp.Client
	httpClient2 *http.Client
}

// NewClient 创建新的 HTTP 客户端
func NewClient() *Client {
	return &Client{
		timeout:   10 * time.Second,
		userAgent: "OneForAll-Go/1.0.0",
		httpClient: &fasthttp.Client{
			ReadTimeout:         10 * time.Second,
			WriteTimeout:        10 * time.Second,
			MaxIdleConnDuration: 30 * time.Second,
		},
		httpClient2: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Request 发送 HTTP 请求
func (c *Client) Request(domain, ip string) (int, string, error) {
	// 尝试 HTTP
	status, title, err := c.requestHTTP(domain, ip, "http")
	if err == nil && status > 0 {
		return status, title, nil
	}

	// 尝试 HTTPS
	status, title, err = c.requestHTTP(domain, ip, "https")
	if err == nil && status > 0 {
		return status, title, nil
	}

	return 0, "", fmt.Errorf("failed to request %s", domain)
}

// requestHTTP 发送指定协议的 HTTP 请求
func (c *Client) requestHTTP(domain, ip, scheme string) (int, string, error) {
	url := fmt.Sprintf("%s://%s", scheme, domain)

	// 创建请求
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetUserAgent(c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "close")

	// 发送请求
	err := c.httpClient.DoTimeout(req, resp, c.timeout)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP request failed: %v", err)
	}

	status := resp.StatusCode()
	if status == 0 {
		return 0, "", fmt.Errorf("invalid response status")
	}

	// 提取标题
	title := c.extractTitle(resp.Body())

	return status, title, nil
}

// RequestWithPort 发送指定端口的 HTTP 请求
func (c *Client) RequestWithPort(domain, ip string, port int) (int, string, error) {
	// 尝试 HTTP
	status, title, err := c.requestHTTPWithPort(domain, ip, port, "http")
	if err == nil && status > 0 {
		return status, title, nil
	}

	// 尝试 HTTPS
	status, title, err = c.requestHTTPWithPort(domain, ip, port, "https")
	if err == nil && status > 0 {
		return status, title, nil
	}

	return 0, "", fmt.Errorf("failed to request %s:%d", domain, port)
}

// requestHTTPWithPort 发送指定端口和协议的 HTTP 请求
func (c *Client) requestHTTPWithPort(domain, ip string, port int, scheme string) (int, string, error) {
	url := fmt.Sprintf("%s://%s:%d", scheme, domain, port)

	// 创建请求
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetUserAgent(c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "close")

	// 发送请求
	err := c.httpClient.DoTimeout(req, resp, c.timeout)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP request failed: %v", err)
	}

	status := resp.StatusCode()
	if status == 0 {
		return 0, "", fmt.Errorf("invalid response status")
	}

	// 提取标题
	title := c.extractTitle(resp.Body())

	return status, title, nil
}

// extractTitle 从 HTML 中提取标题
func (c *Client) extractTitle(body []byte) string {
	bodyStr := string(body)

	// 查找 <title> 标签
	titleStart := strings.Index(strings.ToLower(bodyStr), "<title>")
	if titleStart == -1 {
		return ""
	}

	titleStart += 7 // 跳过 "<title>"
	titleEnd := strings.Index(strings.ToLower(bodyStr[titleStart:]), "</title>")
	if titleEnd == -1 {
		return ""
	}

	title := bodyStr[titleStart : titleStart+titleEnd]

	// 清理标题
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	// 移除多余空格
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}

	// 限制标题长度
	if len(title) > 100 {
		title = title[:100] + "..."
	}

	return title
}

// RequestMultiplePorts 发送多端口请求
func (c *Client) RequestMultiplePorts(domain, ip string, ports []int) map[int]int {
	results := make(map[int]int)

	for _, port := range ports {
		status, _, err := c.RequestWithPort(domain, ip, port)
		if err == nil && status > 0 {
			results[port] = status
		}
	}

	return results
}

// SetTimeout 设置超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.ReadTimeout = timeout
	c.httpClient.WriteTimeout = timeout
	c.httpClient2.Timeout = timeout
}

// SetUserAgent 设置 User-Agent
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// GetUserAgent 获取 User-Agent
func (c *Client) GetUserAgent() string {
	return c.userAgent
}

// GetTimeout 获取超时时间
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}
