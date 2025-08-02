package enrich

import (
	"encoding/json"
	//"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
	"github.com/oneforall-go/pkg/logger"
)

// EnrichResult 反查结果
type EnrichResult struct {
	IP           string   `json:"ip"`
	IsCDN        bool     `json:"is_cdn"`
	ReverseNames []string `json:"reverse_names"`
	Provider     string   `json:"provider"`
}

// Enrich 域名反查模块
type Enrich struct {
	*core.BaseModule
	cdnIPs      map[string]bool
	nameservers []string
	concurrent  int
	timeout     time.Duration
}

// NewEnrich 创建反查模块
func NewEnrich(cfg *config.Config) *Enrich {
	enrich := &Enrich{
		BaseModule: core.NewBaseModule("enrich", "Domain IP enrichment and reverse DNS lookup", cfg),
		cdnIPs:     make(map[string]bool),
		concurrent: cfg.MultiThreading.EnrichConcurrency,
		timeout:    time.Duration(cfg.MultiThreading.EnrichTimeout) * time.Second,
	}

	// 加载CDN IP列表
	enrich.loadCDNIPs()

	// 加载DNS服务器列表
	enrich.loadNameservers()

	return enrich
}

// Run 运行反查模块
func (e *Enrich) Run(domain string) ([]string, error) {
	logger.Infof("Starting domain enrichment for: %s", domain)

	// 获取域名的IP列表
	ips, err := e.getDomainIPs(domain)
	if err != nil {
		logger.Errorf("Failed to get IPs for domain %s: %v", domain, err)
		return []string{}, err
	}

	if len(ips) == 0 {
		logger.Warnf("No IPs found for domain: %s", domain)
		return []string{}, nil
	}

	// 并发处理IP反查
	results := e.enrichIPs(ips)

	// 转换为子域名列表
	var subdomains []string
	for _, result := range results {
		if !result.IsCDN && len(result.ReverseNames) > 0 {
			subdomains = append(subdomains, result.ReverseNames...)
		}
	}

	logger.Infof("Enrichment completed for %s, found %d non-CDN IPs with %d reverse names",
		domain, len(results), len(subdomains))

	return subdomains, nil
}

// loadCDNIPs 加载CDN IP列表
func (e *Enrich) loadCDNIPs() {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in loadCDNIPs: %v", r)
		}
	}()

	data, err := os.ReadFile("data/cdn_ip_cidr.json")
	if err != nil {
		logger.Errorf("Failed to load CDN IP file: %v", err)
		return
	}

	var cidrs []string
	if err := json.Unmarshal(data, &cidrs); err != nil {
		logger.Errorf("Failed to parse CDN IP file: %v", err)
		return
	}

	// 将CIDR转换为IP范围
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Debugf("Invalid CIDR: %s", cidr)
			continue
		}

		// 将网络地址转换为字符串键
		e.cdnIPs[ipNet.String()] = true
	}

	logger.Infof("Loaded %d CDN IP ranges", len(e.cdnIPs))
}

// loadNameservers 加载DNS服务器列表
func (e *Enrich) loadNameservers() {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in loadNameservers: %v", r)
		}
	}()

	// 读取nameservers.txt
	data, err := os.ReadFile("data/nameservers.txt")
	if err != nil {
		logger.Errorf("Failed to load nameservers.txt: %v", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			e.nameservers = append(e.nameservers, line)
		}
	}

	// 读取nameservers_cn.txt
	data, err = os.ReadFile("data/nameservers_cn.txt")
	if err == nil {
		lines = strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				e.nameservers = append(e.nameservers, line)
			}
		}
	}

	logger.Infof("Loaded %d nameservers", len(e.nameservers))
}

// getDomainIPs 获取域名的IP列表
func (e *Enrich) getDomainIPs(domain string) ([]string, error) {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in getDomainIPs for %s: %v", domain, r)
		}
	}()

	ips, err := net.LookupHost(domain)
	if err != nil {
		return nil, err
	}

	// 过滤私有IP
	var publicIPs []string
	for _, ip := range ips {
		if !e.isPrivateIP(ip) {
			publicIPs = append(publicIPs, ip)
		}
	}

	return publicIPs, nil
}

// enrichIPs 并发处理IP反查
func (e *Enrich) enrichIPs(ips []string) []EnrichResult {
	var results []EnrichResult
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, e.concurrent)

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			// 添加异常处理
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic in enrichIPs for %s: %v", ip, r)
				}
			}()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := e.enrichSingleIP(ip)

			mutex.Lock()
			results = append(results, result)
			mutex.Unlock()
		}(ip)
	}

	wg.Wait()
	return results
}

// enrichSingleIP 处理单个IP的反查
func (e *Enrich) enrichSingleIP(ip string) EnrichResult {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in enrichSingleIP for %s: %v", ip, r)
		}
	}()

	result := EnrichResult{
		IP:           ip,
		IsCDN:        false,
		ReverseNames: []string{},
		Provider:     "",
	}

	// 检查是否为CDN IP
	if e.isCDNIP(ip) {
		result.IsCDN = true
		logger.Debugf("IP %s is identified as CDN", ip)
		return result
	}

	// 执行反向DNS查询
	reverseNames := e.reverseDNSLookup(ip)
	result.ReverseNames = reverseNames

	// 获取IP提供商信息
	result.Provider = e.getIPProvider(ip)

	if len(reverseNames) > 0 {
		logger.Debugf("IP %s reverse lookup found: %v", ip, reverseNames)
	}

	return result
}

// isCDNIP 检查IP是否为CDN
func (e *Enrich) isCDNIP(ip string) bool {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in isCDNIP for %s: %v", ip, r)
		}
	}()

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查IP是否在CDN范围内
	for cidr := range e.cdnIPs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// reverseDNSLookup 反向DNS查询
func (e *Enrich) reverseDNSLookup(ip string) []string {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in reverseDNSLookup for %s: %v", ip, r)
		}
	}()

	var reverseNames []string

	// 遍历多个DNS服务器
	for _, nameserver := range e.nameservers {
		names, err := e.queryReverseDNS(ip, nameserver)
		if err != nil {
			continue
		}

		reverseNames = append(reverseNames, names...)
	}

	// 去重
	return e.deduplicateStrings(reverseNames)
}

// queryReverseDNS 查询反向DNS
func (e *Enrich) queryReverseDNS(ip string, nameserver string) ([]string, error) {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in queryReverseDNS for %s: %v", ip, r)
		}
	}()

	client := new(dns.Client)
	client.Timeout = e.timeout

	// 构建反向查询域名
	reverseIP := e.reverseIP(ip)
	msg := new(dns.Msg)
	msg.SetQuestion(reverseIP, dns.TypePTR)
	msg.RecursionDesired = true

	resp, _, err := client.Exchange(msg, nameserver+":53")
	if err != nil {
		return nil, err
	}

	var names []string
	for _, answer := range resp.Answer {
		if ptr, ok := answer.(*dns.PTR); ok {
			name := strings.TrimSuffix(ptr.Ptr, ".")
			names = append(names, name)
		}
	}

	return names, nil
}

// reverseIP 反转IP地址用于PTR查询
func (e *Enrich) reverseIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ""
	}

	// 反转IP段
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ".") + ".in-addr.arpa."
}

// isPrivateIP 检查是否为私有IP
func (e *Enrich) isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 私有IP范围
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	}

	for _, cidr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// getIPProvider 获取IP提供商信息
func (e *Enrich) getIPProvider(ip string) string {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in getIPProvider for %s: %v", ip, r)
		}
	}()

	// 这里可以集成IP地理位置数据库
	// 目前返回占位符
	return "Unknown"
}

// deduplicateStrings 字符串去重
func (e *Enrich) deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}
