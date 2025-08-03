package validator

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// DomainValidator 域名验证器
type DomainValidator struct {
	config      *config.Config
	client      *http.Client
	httpsClient *http.Client
}

// ValidationResult 验证结果
type ValidationResult struct {
	Subdomain   string   `json:"subdomain"`
	IP          []string `json:"ip"`
	Status      int      `json:"status"`
	Title       string   `json:"title"`
	Port        int      `json:"port"`
	Alive       bool     `json:"alive"`
	Source      string   `json:"source"`
	Time        string   `json:"time"`
	Provider    string   `json:"provider"`
	DNSResolved bool     `json:"dns_resolved"`
	PingAlive   bool     `json:"ping_alive"`
	StatusCode  int      `json:"status_code"` // 新增状态码字段
	StatusText  string   `json:"status_text"` // 新增状态文本字段
}

// NewDomainValidator 创建域名验证器
func NewDomainValidator(cfg *config.Config) *DomainValidator {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second, // 设置60秒超时
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// 创建HTTPS客户端（忽略证书验证）
	httpsClient := &http.Client{
		Timeout: 60 * time.Second, // 设置60秒超时
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 忽略证书验证
			},
		},
	}

	return &DomainValidator{
		config:      cfg,
		client:      client,
		httpsClient: httpsClient,
	}
}

// ValidateDomains 验证域名列表
func (v *DomainValidator) ValidateDomains(domains []string, concurrency int) []ValidationResult {
	if len(domains) == 0 {
		return []ValidationResult{}
	}

	logger.Infof("Starting comprehensive domain validation for %d domains with concurrency %d", len(domains), concurrency)

	// 去重
	uniqueDomains := v.deduplicateDomains(domains)
	logger.Infof("After deduplication: %d unique domains", len(uniqueDomains))

	// 并发验证
	var results []ValidationResult
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errors []error

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, concurrency)

	for _, domain := range uniqueDomains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()

			// 添加异常处理
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic in domain validation for %s: %v", domain, r)
					mutex.Lock()
					errors = append(errors, fmt.Errorf("panic in %s: %v", domain, r))
					mutex.Unlock()
				}
			}()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := v.validateSingleDomain(domain)

			// 添加所有验证结果，不管是否存活
			mutex.Lock()
			results = append(results, result)
			mutex.Unlock()

			if result.Alive {
				logger.Debugf("Domain %s is alive (IP: %v, DNS: %v, Ping: %v, Status: %d, Provider: %s)",
					domain, result.IP, result.DNSResolved, result.PingAlive, result.StatusCode, result.Provider)
			} else {
				logger.Debugf("Domain %s is not alive (DNS: %v, Ping: %v, Status: %d)",
					domain, result.DNSResolved, result.PingAlive, result.StatusCode)
			}
		}(domain)
	}

	wg.Wait()

	if len(errors) > 0 {
		logger.Warnf("Some validation errors occurred: %v", errors)
	}

	logger.Infof("Domain validation completed. Processed %d unique domains", len(results))
	return results
}

// validateSingleDomain 验证单个域名
func (v *DomainValidator) validateSingleDomain(domain string) ValidationResult {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in validateSingleDomain for %s: %v", domain, r)
		}
	}()

	result := ValidationResult{
		Subdomain:   domain,
		IP:          []string{},
		Status:      0,
		Title:       "",
		Port:        0,
		Alive:       false,
		Source:      "validator",
		Time:        time.Now().Format("2006-01-02 15:04:05"),
		Provider:    "",
		DNSResolved: false,
		PingAlive:   false,
		StatusCode:  0,
		StatusText:  "",
	}

	// 1. DNS 解析验证
	ips := v.resolveDomain(domain)
	if len(ips) > 0 {
		result.IP = ips
		result.DNSResolved = true
		logger.Debugf("DNS resolution successful for %s: %v", domain, ips)

		// 2. Ping 验证（TCP连接测试）
		result.PingAlive = v.validatePing(ips[0])
		if result.PingAlive {
			result.Alive = true
			result.StatusCode = 200
			result.StatusText = "Alive"
			logger.Debugf("Ping successful for %s", domain)

			// 3. IP供应商查询
			if len(ips) > 0 {
				result.Provider = v.getIPProvider(ips[0])
			}
		} else {
			result.StatusCode = 0
			result.StatusText = "Ping Failed"
			logger.Debugf("Ping failed for %s", domain)
		}
	} else {
		result.StatusCode = -1
		result.StatusText = "DNS Resolution Failed"
		logger.Debugf("DNS resolution failed for %s", domain)
	}

	logger.Debugf("Validation result for %s: Alive=%v, DNS=%v, Ping=%v, Status=%d, Text=%s",
		domain, result.Alive, result.DNSResolved, result.PingAlive, result.StatusCode, result.StatusText)

	return result
}

// resolveDomain DNS 解析域名
func (v *DomainValidator) resolveDomain(domain string) []string {
	var ips []string

	// 尝试 A 记录解析
	addresses, err := net.LookupHost(domain)
	if err != nil {
		logger.Debugf("Failed to resolve %s: %v", domain, err)
		return ips
	}

	// 过滤有效的 IP 地址
	for _, addr := range addresses {
		if ip := net.ParseIP(addr); ip != nil {
			// 排除私有 IP 地址（可选）
			if !v.config.ExcludePrivateIP || !isPrivateIP(ip) {
				ips = append(ips, addr)
			}
		}
	}

	return ips
}

// validateHTTPRequest 验证HTTP请求
func (v *DomainValidator) validateHTTPRequest(domain string, protocol string) bool {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in validateHTTPRequest for %s: %v", domain, r)
		}
	}()

	url := fmt.Sprintf("%s://%s", protocol, domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Debugf("Failed to create %s request for %s: %v", protocol, domain, err)
		return false
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "OneForAll-Go/1.0")

	resp, err := v.client.Do(req)
	if err != nil {
		logger.Debugf("%s request failed for %s: %v", protocol, domain, err)
		return false
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		logger.Debugf("%s request successful for %s (Status: %d)", protocol, domain, resp.StatusCode)
		return true
	}

	logger.Debugf("%s request failed for %s (Status: %d)", protocol, domain, resp.StatusCode)
	return false
}

// validateHTTPSRequest 验证HTTPS请求（忽略证书验证）
func (v *DomainValidator) validateHTTPSRequest(domain string) bool {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in validateHTTPSRequest for %s: %v", domain, r)
		}
	}()

	url := fmt.Sprintf("https://%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Debugf("Failed to create HTTPS request for %s: %v", domain, err)
		return false
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "OneForAll-Go/1.0")

	resp, err := v.httpsClient.Do(req)
	if err != nil {
		logger.Debugf("HTTPS request failed for %s: %v", domain, err)
		return false
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		logger.Debugf("HTTPS request successful for %s (Status: %d)", domain, resp.StatusCode)
		return true
	}

	logger.Debugf("HTTPS request failed for %s (Status: %d)", domain, resp.StatusCode)
	return false
}

// getHTTPStatus 获取HTTP状态和标题
func (v *DomainValidator) getHTTPStatus(domain string, protocol string) (int, string) {
	url := fmt.Sprintf("%s://%s", protocol, domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, ""
	}

	req.Header.Set("User-Agent", "OneForAll-Go/1.0")

	resp, err := v.client.Do(req)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	// 尝试读取标题（简化实现）
	title := ""
	if resp.StatusCode == 200 {
		title = "OK"
	}

	return resp.StatusCode, title
}

// getHTTPSStatus 获取HTTPS状态和标题
func (v *DomainValidator) getHTTPSStatus(domain string) (int, string) {
	url := fmt.Sprintf("https://%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, ""
	}

	req.Header.Set("User-Agent", "OneForAll-Go/1.0")

	resp, err := v.httpsClient.Do(req)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	// 尝试读取标题（简化实现）
	title := ""
	if resp.StatusCode == 200 {
		title = "OK"
	}

	return resp.StatusCode, title
}

// getIPProvider 获取 IP 提供商信息
func (v *DomainValidator) getIPProvider(ip string) string {
	// 使用开源数据查询IP供应商
	// 这里可以实现多种IP供应商查询服务
	providers := []string{
		"ipinfo.io",
		"ip-api.com",
		"ipapi.co",
		"ipgeolocation.io",
	}

	for _, provider := range providers {
		if providerName := v.queryIPProvider(ip, provider); providerName != "" {
			return providerName
		}
	}

	return "Unknown"
}

// queryIPProvider 查询IP供应商
func (v *DomainValidator) queryIPProvider(ip string, provider string) string {
	// 这里实现具体的IP供应商查询逻辑
	// 暂时返回简化结果
	switch provider {
	case "ipinfo.io":
		return "IPInfo"
	case "ip-api.com":
		return "IP-API"
	case "ipapi.co":
		return "IPAPI"
	case "ipgeolocation.io":
		return "IPGeolocation"
	default:
		return ""
	}
}

// deduplicateDomains 域名去重
func (v *DomainValidator) deduplicateDomains(domains []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, domain := range domains {
		// 标准化域名（小写，去除空格）
		normalized := strings.ToLower(strings.TrimSpace(domain))
		if normalized != "" && !seen[normalized] {
			seen[normalized] = true
			unique = append(unique, normalized)
		}
	}

	return unique
}

// isPrivateIP 检查是否为私有 IP
func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return false
}

// FilterAliveDomains 过滤存活域名
func (v *DomainValidator) FilterAliveDomains(results []ValidationResult) []ValidationResult {
	var aliveResults []ValidationResult
	for _, result := range results {
		if result.Alive {
			aliveResults = append(aliveResults, result)
		}
	}
	return aliveResults
}

// GetValidationStats 获取验证统计信息
func (v *DomainValidator) GetValidationStats(results []ValidationResult) map[string]interface{} {
	stats := make(map[string]interface{})

	total := len(results)
	alive := 0
	dead := 0
	dnsResolved := 0
	pingAlive := 0
	uniqueIPs := make(map[string]bool)
	providers := make(map[string]int)

	for _, result := range results {
		if result.Alive {
			alive++
		} else {
			dead++
		}

		if result.DNSResolved {
			dnsResolved++
		}

		if result.PingAlive {
			pingAlive++
		}

		for _, ip := range result.IP {
			uniqueIPs[ip] = true
		}

		if result.Provider != "" {
			providers[result.Provider]++
		}
	}

	stats["total_domains"] = total
	stats["alive_domains"] = alive
	stats["dead_domains"] = dead
	stats["dns_resolved"] = dnsResolved
	stats["ping_alive"] = pingAlive
	stats["unique_ips"] = len(uniqueIPs)
	stats["providers"] = providers

	if total > 0 {
		stats["alive_percentage"] = float64(alive) / float64(total) * 100
		stats["dns_percentage"] = float64(dnsResolved) / float64(total) * 100
		stats["ping_percentage"] = float64(pingAlive) / float64(total) * 100
	} else {
		stats["alive_percentage"] = 0.0
		stats["dns_percentage"] = 0.0
		stats["ping_percentage"] = 0.0
	}

	return stats
}

// validatePing 验证IP是否可以ping通
func (v *DomainValidator) validatePing(ip string) bool {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in validatePing for %s: %v", ip, r)
		}
	}()

	// 使用TCP连接验证IP是否可达
	conn, err := net.DialTimeout("tcp", ip+":80", 5*time.Second)
	if err != nil {
		// 尝试443端口
		conn, err = net.DialTimeout("tcp", ip+":443", 5*time.Second)
		if err != nil {
			logger.Debugf("Ping failed for %s: %v", ip, err)
			return false
		}
	}
	defer conn.Close()

	logger.Debugf("Ping successful for %s", ip)
	return true
}
