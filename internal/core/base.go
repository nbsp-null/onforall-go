package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// ModuleType 模块类型
type ModuleType string

const (
	ModuleTypeSearch    ModuleType = "search"     // 搜索引擎、OSINT、病毒扫描、资产平台等
	ModuleTypeBrute     ModuleType = "brute"      // 爆破模块
	ModuleTypeDNSLookup ModuleType = "dns_lookup" // DNS 反射查询
	ModuleTypeResolve   ModuleType = "resolve"    // 域名到 IP 解析
	ModuleTypeCheck     ModuleType = "check"      // 检查模块
	ModuleTypeCrawl     ModuleType = "crawl"      // 爬虫模块
	ModuleTypeEnrich    ModuleType = "enrich"     // 信息丰富模块
)

// Module 模块接口
type Module interface {
	Name() string
	Type() ModuleType
	Run(domain string) ([]string, error)
	IsEnabled() bool
	SetEnabled(enabled bool)
}

// BaseModule 基础模块类（对应 Python 的 Module 基类）
type BaseModule struct {
	name       string
	moduleType ModuleType
	enabled    bool
	config     *config.Config
	domain     string
	subdomains map[string]bool
	infos      map[string]interface{}
	results    []interface{}
	startTime  time.Time
	endTime    time.Time
	elapsed    time.Duration

	// HTTP 相关
	httpClient *http.Client
	userAgents []string
	cookie     *http.Cookie
	header     map[string]string
	proxy      *url.URL
	delay      time.Duration
	timeout    time.Duration
	retryCount int

	// 线程安全
	mutex sync.RWMutex
}

// NewBaseModule 创建基础模块
func NewBaseModule(name string, moduleType ModuleType, cfg *config.Config) *BaseModule {
	return &BaseModule{
		name:       name,
		moduleType: moduleType,
		enabled:    true,
		config:     cfg,
		subdomains: make(map[string]bool),
		infos:      make(map[string]interface{}),
		results:    make([]interface{}, 0),
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
		header:     make(map[string]string),
		delay:      time.Duration(rand.Intn(3)+1) * time.Second,
		timeout:    time.Duration(cfg.DNSResolveTimeout) * time.Second,
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

// SetDomain 设置域名
func (b *BaseModule) SetDomain(domain string) {
	b.domain = domain
}

// GetDomain 获取域名
func (b *BaseModule) GetDomain() string {
	return b.domain
}

// Begin 开始执行
func (b *BaseModule) Begin() {
	b.startTime = time.Now()
	b.LogDebug("Starting %s module to collect subdomains of %s", b.name, b.domain)
}

// Finish 结束执行
func (b *BaseModule) Finish() {
	b.endTime = time.Now()
	b.elapsed = b.endTime.Sub(b.startTime)
	b.LogDebug("Finished %s module to collect %s's subdomains", b.name, b.domain)
	b.LogInfo("%s module took %v seconds found %d subdomains", b.name, b.elapsed, len(b.subdomains))
}

// AddSubdomain 添加子域名
func (b *BaseModule) AddSubdomain(subdomain string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.subdomains[subdomain] = true
}

// GetSubdomains 获取所有子域名
func (b *BaseModule) GetSubdomains() []string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	subdomains := make([]string, 0, len(b.subdomains))
	for subdomain := range b.subdomains {
		subdomains = append(subdomains, subdomain)
	}
	return subdomains
}

// AddInfo 添加信息
func (b *BaseModule) AddInfo(key string, value interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.infos[key] = value
}

// GetInfo 获取信息
func (b *BaseModule) GetInfo(key string) interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.infos[key]
}

// AddResult 添加结果
func (b *BaseModule) AddResult(result interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.results = append(b.results, result)
}

// GetResults 获取所有结果
func (b *BaseModule) GetResults() []interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	results := make([]interface{}, len(b.results))
	copy(results, b.results)
	return results
}

// HTTP 请求相关方法

// GetRandomUserAgent 获取随机 User-Agent
func (b *BaseModule) GetRandomUserAgent() string {
	return b.userAgents[rand.Intn(len(b.userAgents))]
}

// SetHeader 设置请求头
func (b *BaseModule) SetHeader(key, value string) {
	b.header[key] = value
}

// GetHeader 获取请求头
func (b *BaseModule) GetHeader() map[string]string {
	headers := make(map[string]string)
	for k, v := range b.header {
		headers[k] = v
	}
	return headers
}

// SetProxy 设置代理
func (b *BaseModule) SetProxy(proxyURL string) error {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}
	b.proxy = proxy
	return nil
}

// SetDelay 设置延迟
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

// HTTPPostJSON 执行 HTTP POST JSON 请求
func (b *BaseModule) HTTPPostJSON(urlStr string, data interface{}, headers map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
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
		headers["Content-Type"] = "application/json"
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

// 子域名处理相关方法

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

// API 相关方法

// GetAPIKey 获取 API 密钥
func (b *BaseModule) GetAPIKey(keyName string) string {
	return b.config.APIKeys[keyName]
}

// HaveAPI 检查是否有 API 密钥
func (b *BaseModule) HaveAPI(keys ...string) bool {
	for _, key := range keys {
		if b.GetAPIKey(key) == "" {
			b.LogDebug("%s module is not configured", b.name)
			return false
		}
	}
	return true
}

// 日志相关方法

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
