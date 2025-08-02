package modules

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// BaseModule 基础模块结构
type BaseModule struct {
	name       string
	moduleType ModuleType
	enabled    bool
	config     *config.Config
	httpClient *http.Client
	userAgents []string
	delay      time.Duration
	retryCount int
}

// NewBaseModule 创建基础模块
func NewBaseModule(name string, moduleType ModuleType, cfg *config.Config) *BaseModule {
	return &BaseModule{
		name:       name,
		moduleType: moduleType,
		enabled:    true,
		config:     cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.DNSResolveTimeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Googlebot/2.1 (+http://www.google.com/bot.html)",
		},
		delay:      time.Duration(rand.Intn(3)+1) * time.Second,
		retryCount: 3,
	}
}

// Name 获取模块名称
func (b *BaseModule) Name() string {
	return b.name
}

// Type 获取模块类型
func (b *BaseModule) Type() ModuleType {
	return b.moduleType
}

// IsEnabled 检查模块是否启用
func (b *BaseModule) IsEnabled() bool {
	return b.enabled
}

// SetEnabled 设置模块启用状态
func (b *BaseModule) SetEnabled(enabled bool) {
	b.enabled = enabled
}

// GetRandomUserAgent 获取随机 User-Agent
func (b *BaseModule) GetRandomUserAgent() string {
	return b.userAgents[rand.Intn(len(b.userAgents))]
}

// GetHTTPClient 获取 HTTP 客户端
func (b *BaseModule) GetHTTPClient() *http.Client {
	return b.httpClient
}

// SetDelay 设置延迟时间
func (b *BaseModule) SetDelay(delay time.Duration) {
	b.delay = delay
}

// Sleep 随机延迟
func (b *BaseModule) Sleep() {
	time.Sleep(b.delay)
}

// HTTPGet 执行 HTTP GET 请求
func (b *BaseModule) HTTPGet(urlStr string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// 设置默认 User-Agent
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["User-Agent"]; !exists {
		headers["User-Agent"] = b.GetRandomUserAgent()
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 随机延迟
	b.Sleep()

	// 重试机制
	var resp *http.Response
	for i := 0; i < b.retryCount; i++ {
		resp, err = b.httpClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i < b.retryCount-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return resp, err
}

// HTTPPost 执行 HTTP POST 请求
func (b *BaseModule) HTTPPost(urlStr string, data url.Values, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置默认请求头
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["User-Agent"]; !exists {
		headers["User-Agent"] = b.GetRandomUserAgent()
	}
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 随机延迟
	b.Sleep()

	// 重试机制
	var resp *http.Response
	for i := 0; i < b.retryCount; i++ {
		resp, err = b.httpClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i < b.retryCount-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return resp, err
}

// ExtractSubdomains 从文本中提取子域名
func (b *BaseModule) ExtractSubdomains(text, domain string) []string {
	var subdomains []string

	// 构建正则表达式模式
	pattern := fmt.Sprintf(`([a-zA-Z0-9.-]+\.%s)`, regexp.QuoteMeta(domain))
	re := regexp.MustCompile(pattern)

	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		// 清理和验证子域名
		subdomain := strings.TrimSpace(match)
		if b.IsValidSubdomain(subdomain, domain) {
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains
}

// IsValidSubdomain 验证子域名是否有效
func (b *BaseModule) IsValidSubdomain(subdomain, domain string) bool {
	// 基本验证
	if subdomain == domain {
		return false
	}

	if !strings.HasSuffix(subdomain, "."+domain) {
		return false
	}

	// 检查是否包含无效字符
	invalidChars := []string{" ", "\t", "\n", "\r", "\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(subdomain, char) {
			return false
		}
	}

	return true
}

// Deduplicate 去重
func (b *BaseModule) Deduplicate(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ReadResponseBody 读取响应体
func (b *BaseModule) ReadResponseBody(resp *http.Response) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("response is nil")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetAPIKey 获取 API 密钥
func (b *BaseModule) GetAPIKey(keyName string) string {
	return b.config.APIKeys[keyName]
}

// LogDebug 记录调试日志
func (b *BaseModule) LogDebug(format string, args ...interface{}) {
	logger.Debugf("[%s] "+format, append([]interface{}{b.name}, args...)...)
}

// LogInfo 记录信息日志
func (b *BaseModule) LogInfo(format string, args ...interface{}) {
	logger.Infof("[%s] "+format, append([]interface{}{b.name}, args...)...)
}

// LogError 记录错误日志
func (b *BaseModule) LogError(format string, args ...interface{}) {
	logger.Errorf("[%s] "+format, append([]interface{}{b.name}, args...)...)
}
