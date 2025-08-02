package brute

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
	"github.com/oneforall-go/pkg/logger"
)

// Brute 爆破模块
type Brute struct {
	*core.BaseModule
	domain         string
	wordlist       string
	nextlist       string
	concurrent     int
	recursive      bool
	depth          int
	enableWildcard bool
	wildcardIPs    []string
	wildcardTTL    int
	nameservers    []string
	results        map[string]*BruteResult
	mu             sync.RWMutex

	// 进度跟踪
	totalCount     int
	processedCount int
	successCount   int
	startTime      time.Time
}

// BruteResult 爆破结果
type BruteResult struct {
	Subdomain string   `json:"subdomain"`
	IPs       []string `json:"ips"`
	CNAMEs    []string `json:"cnames"`
	TTLs      []int    `json:"ttls"`
	Valid     bool     `json:"valid"`
}

// WildcardDetectionResult 泛解析检测结果
type WildcardDetectionResult struct {
	IsWildcard     bool                    `json:"is_wildcard"`
	SuccessRate    float64                 `json:"success_rate"`
	IPRepeatRate   float64                 `json:"ip_repeat_rate"`
	TestCount      int                     `json:"test_count"`
	SuccessCount   int                     `json:"success_count"`
	UniqueIPs      int                     `json:"unique_ips"`
	TotalIPs       int                     `json:"total_ips"`
	TestSubdomains []string                `json:"test_subdomains"`
	TestResults    map[string]*BruteResult `json:"test_results"`
}

// NewBrute 创建爆破模块
func NewBrute(cfg *config.Config) *Brute {
	brute := &Brute{
		BaseModule: core.NewBaseModule("Brute", core.ModuleTypeBrute, cfg),
		domain:     "",
		wordlist:   "",
		nextlist:   "",
		concurrent: 20, // 默认设置为20个线程
		recursive:  false,
		depth:      1,
		results:    make(map[string]*BruteResult),
	}

	// 如果配置中有设置，则使用配置值
	if cfg.MultiThreading.BruteForceConcurrency > 0 {
		brute.concurrent = cfg.MultiThreading.BruteForceConcurrency
	}

	// 初始化字典路径
	brute.initDictPaths()

	logger.Infof("Brute force module initialized with concurrency: %d", brute.concurrent)
	return brute
}

// Run 运行爆破模块
func (b *Brute) Run(domain string) ([]string, error) {
	logger.Infof("=== Starting brute force attack for domain: %s ===", domain)
	logger.Debugf("Brute module configuration:")
	logger.Debugf("  - Domain: %s", domain)
	logger.Debugf("  - Concurrent: %d", b.concurrent)
	logger.Debugf("  - Wordlist: %s", b.wordlist)
	logger.Debugf("  - Nextlist: %s", b.nextlist)
	logger.Debugf("  - Recursive: %t", b.recursive)
	logger.Debugf("  - Depth: %d", b.depth)

	// 初始化字典路径
	logger.Debugf("Initializing dictionary paths...")
	b.initDictPaths()
	logger.Debugf("Dictionary paths initialized: wordlist=%s, nextlist=%s", b.wordlist, b.nextlist)

	// 获取权威 DNS 服务器
	logger.Debugf("Getting nameservers for domain: %s", domain)
	if err := b.getNameservers(domain); err != nil {
		logger.Errorf("Failed to get nameservers: %v", err)
		return nil, err
	}
	logger.Debugf("Nameservers loaded: %v", b.nameservers)

	// 高级泛解析检测
	logger.Debugf("Starting advanced wildcard detection for domain: %s", domain)
	wildcardResult, err := b.detectWildcardAdvanced(domain)
	if err != nil {
		logger.Errorf("Failed to detect wildcard: %v", err)
		return nil, err
	}

	// 如果检测到泛解析，跳过爆破
	if wildcardResult.IsWildcard {
		logger.Warnf("Wildcard DNS detected for domain %s. Skipping brute force attack.", domain)
		logger.Infof("Wildcard detection details:")
		logger.Infof("  - Success rate: %.2f%% (threshold: 90%%)", wildcardResult.SuccessRate)
		logger.Infof("  - IP repeat rate: %.2f%% (threshold: 50%%)", wildcardResult.IPRepeatRate)
		logger.Infof("  - Test subdomains: %v", wildcardResult.TestSubdomains)
		logger.Infof("  - Reason: Domain has wildcard DNS enabled, brute force would be ineffective")
		return []string{}, nil
	}

	// 没有检测到泛解析，继续爆破
	logger.Infof("No wildcard DNS detected for domain %s. Proceeding with brute force attack.", domain)
	logger.Infof("Wildcard detection details:")
	logger.Infof("  - Success rate: %.2f%% (threshold: 90%%)", wildcardResult.SuccessRate)
	logger.Infof("  - IP repeat rate: %.2f%% (threshold: 50%%)", wildcardResult.IPRepeatRate)
	logger.Infof("  - Test subdomains: %v", wildcardResult.TestSubdomains)
	logger.Infof("  - Reason: Domain does not have wildcard DNS, brute force will be effective")

	// 检测泛解析（保留原有逻辑作为备用）
	logger.Debugf("Running fallback wildcard detection for domain: %s", domain)
	if err := b.detectWildcard(domain); err != nil {
		logger.Errorf("Failed to detect wildcard: %v", err)
		return nil, err
	}
	logger.Debugf("Fallback wildcard detection completed: enableWildcard=%t, wildcardIPs=%v", b.enableWildcard, b.wildcardIPs)

	// 生成爆破字典
	logger.Debugf("Generating dictionary for domain: %s", domain)
	subdomains, err := b.generateDict(domain)
	if err != nil {
		logger.Errorf("Failed to generate dictionary: %v", err)
		return nil, err
	}
	logger.Infof("Dictionary generated: %d subdomains", len(subdomains))

	// 初始化统计信息
	b.totalCount = len(subdomains)
	b.processedCount = 0
	b.successCount = 0
	b.startTime = time.Now()

	logger.Infof("Brute force configuration: concurrency=%d, wordlist_size=%d",
		b.concurrent, b.totalCount)
	logger.Infof("Brute force attack initialized: %d subdomains to test", b.totalCount)

	// 执行爆破
	logger.Debugf("Starting brute force subdomain testing...")
	if err := b.bruteSubdomains(domain, subdomains); err != nil {
		logger.Errorf("Brute force subdomain testing failed: %v", err)
		return []string{}, err
	}

	// 计算最终统计信息
	elapsed := time.Since(b.startTime)
	successRate := float64(b.successCount) / float64(b.totalCount) * 100

	logger.Infof("=== Brute force attack completed! ===")
	logger.Infof("Final statistics:")
	logger.Infof("  - Total subdomains tested: %d", b.totalCount)
	logger.Infof("  - Successfully resolved: %d", b.successCount)
	logger.Infof("  - Success rate: %.2f%%", successRate)
	logger.Infof("  - Total time elapsed: %v", elapsed)
	logger.Infof("  - Average speed: %.2f subdomains/second", float64(b.totalCount)/elapsed.Seconds())

	// 返回结果
	var results []string
	for subdomain, result := range b.results {
		if result.Valid {
			results = append(results, subdomain)
		}
	}

	logger.Infof("Brute force attack found %d valid subdomains", len(results))
	if len(results) > 0 {
		logger.Debugf("Valid subdomains found: %v", results)
	}
	return results, nil
}

// initDictPaths 初始化字典文件路径
func (b *Brute) initDictPaths() {
	logger.Debugf("Initializing dictionary paths...")

	dataDir := "data"
	b.wordlist = filepath.Join(dataDir, "subnames.txt")
	b.nextlist = filepath.Join(dataDir, "subnames_next.txt")

	logger.Debugf("Dictionary paths set:")
	logger.Debugf("  - Wordlist: %s", b.wordlist)
	logger.Debugf("  - Nextlist: %s", b.nextlist)

	// 检查文件是否存在
	if _, err := os.Stat(b.wordlist); os.IsNotExist(err) {
		logger.Errorf("Wordlist file does not exist: %s", b.wordlist)
	} else {
		logger.Debugf("Wordlist file exists: %s", b.wordlist)
	}

	if _, err := os.Stat(b.nextlist); os.IsNotExist(err) {
		logger.Errorf("Nextlist file does not exist: %s", b.nextlist)
	} else {
		logger.Debugf("Nextlist file exists: %s", b.nextlist)
	}
}

// getNameservers 获取权威 DNS 服务器
func (b *Brute) getNameservers(domain string) error {
	logger.Debugf("Getting nameservers for domain: %s", domain)

	// 查询 NS 记录
	logger.Debugf("Querying NS records for domain: %s", domain)
	nsRecords, err := b.queryNS(domain)
	if err != nil {
		logger.Errorf("Failed to query NS records: %v", err)
		return err
	}
	logger.Debugf("NS records found: %v", nsRecords)

	// 查询 NS 服务器的 A 记录
	logger.Debugf("Querying A records for NS servers...")
	for _, ns := range nsRecords {
		logger.Debugf("Querying A record for NS server: %s", ns)
		ips, err := b.queryA(ns)
		if err != nil {
			logger.Debugf("Failed to get A record for NS server %s: %v", ns, err)
			continue
		}
		logger.Debugf("A records for NS server %s: %v", ns, ips)
		b.nameservers = append(b.nameservers, ips...)
	}

	// 如果没有获取到权威服务器，使用公共 DNS
	if len(b.nameservers) == 0 {
		logger.Warnf("No authoritative nameservers found, using public DNS servers")
		b.nameservers = b.getPublicNameservers()
	}

	logger.Infof("Nameservers loaded: %d servers", len(b.nameservers))
	logger.Debugf("Nameservers list: %v", b.nameservers)
	return nil
}

// queryNS 查询 NS 记录
func (b *Brute) queryNS(domain string) ([]string, error) {
	logger.Debugf("Querying NS records for domain: %s", domain)

	client := new(dns.Client)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	msg.RecursionDesired = true

	logger.Debugf("Sending NS query to 8.8.8.8:53")
	resp, _, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		logger.Errorf("NS query failed: %v", err)
		return nil, err
	}

	var nsRecords []string
	for _, answer := range resp.Answer {
		if ns, ok := answer.(*dns.NS); ok {
			nsRecord := strings.TrimSuffix(ns.Ns, ".")
			nsRecords = append(nsRecords, nsRecord)
			logger.Debugf("Found NS record: %s", nsRecord)
		}
	}

	logger.Debugf("NS query completed, found %d records", len(nsRecords))
	return nsRecords, nil
}

// queryA 查询 A 记录
func (b *Brute) queryA(domain string) ([]string, error) {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in queryA for %s: %v", domain, r)
		}
	}()

	client := new(dns.Client)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	msg.RecursionDesired = true

	// 遍历多个DNS服务器
	for _, nameserver := range b.nameservers {
		resp, _, err := client.Exchange(msg, nameserver+":53")
		if err != nil {
			continue
		}

		var ips []string
		for _, answer := range resp.Answer {
			if a, ok := answer.(*dns.A); ok {
				ips = append(ips, a.A.String())
			}
		}

		if len(ips) > 0 {
			return ips, nil
		}
	}

	return nil, fmt.Errorf("no A record found for %s", domain)
}

// getPublicNameservers 获取公共 DNS 服务器
func (b *Brute) getPublicNameservers() []string {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in getPublicNameservers: %v", r)
		}
	}()

	// 读取nameservers.txt文件
	nameservers := []string{}

	// 尝试读取nameservers.txt
	data, err := os.ReadFile("data/nameservers.txt")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				nameservers = append(nameservers, line)
			}
		}
		logger.Infof("Loaded %d nameservers from data/nameservers.txt", len(nameservers))
	} else {
		logger.Warnf("Failed to load nameservers.txt: %v", err)
	}

	// 如果文件读取失败，使用默认的公共DNS服务器
	if len(nameservers) == 0 {
		nameservers = []string{
			"8.8.8.8",
			"8.8.4.4",
			"1.1.1.1",
			"1.0.0.1",
			"114.114.114.114",
			"114.114.115.115",
		}
		logger.Infof("Using default nameservers: %v", nameservers)
	}

	return nameservers
}

// detectWildcard 检测泛解析
func (b *Brute) detectWildcard(domain string) error {
	// 生成随机子域名进行测试
	randomSubdomains := []string{
		fmt.Sprintf("test%d.%s", time.Now().Unix(), domain),
		fmt.Sprintf("random%d.%s", time.Now().Unix(), domain),
	}

	var testIPs []string
	for _, subdomain := range randomSubdomains {
		ips, err := b.queryA(subdomain)
		if err != nil {
			continue
		}
		testIPs = append(testIPs, ips...)
	}

	// 如果随机子域名都解析到相同的 IP，可能存在泛解析
	if len(testIPs) > 0 {
		b.enableWildcard = true
		b.wildcardIPs = testIPs
	}

	return nil
}

// detectWildcardAdvanced 高级泛解析检测
func (b *Brute) detectWildcardAdvanced(domain string) (*WildcardDetectionResult, error) {
	logger.Infof("=== Starting advanced wildcard detection for domain: %s ===", domain)

	result := &WildcardDetectionResult{
		IsWildcard:     false,
		SuccessRate:    0.0,
		IPRepeatRate:   0.0,
		TestCount:      20,
		SuccessCount:   0,
		UniqueIPs:      0,
		TotalIPs:       0,
		TestSubdomains: make([]string, 0),
		TestResults:    make(map[string]*BruteResult),
	}

	// 生成随机测试子域名
	testSubdomains, err := b.generateRandomTestSubdomains(domain, result.TestCount)
	if err != nil {
		logger.Errorf("Failed to generate random test subdomains: %v", err)
		return result, err
	}

	result.TestSubdomains = testSubdomains
	logger.Infof("Generated %d random test subdomains for wildcard detection", len(testSubdomains))

	// 并发测试子域名
	var wg sync.WaitGroup
	var mu sync.Mutex
	allIPs := make(map[string]int) // IP -> 出现次数

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, b.concurrent)

	logger.Debugf("Starting concurrent wildcard detection with %d test subdomains", len(testSubdomains))

	for _, subdomain := range testSubdomains {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(subdomain string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// 添加异常处理
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic in wildcard detection for %s: %v", subdomain, r)
				}
			}()

			// 查询子域名
			bruteResult := b.querySubdomain(subdomain)

			mu.Lock()
			result.TestResults[subdomain] = bruteResult

			if bruteResult.Valid {
				result.SuccessCount++
				// 统计IP出现次数
				for _, ip := range bruteResult.IPs {
					allIPs[ip]++
				}
			}
			mu.Unlock()

			logger.Debugf("Wildcard test: %s -> Valid: %t, IPs: %v", subdomain, bruteResult.Valid, bruteResult.IPs)
		}(subdomain)
	}

	logger.Debugf("Waiting for all wildcard detection goroutines to complete...")
	wg.Wait()
	logger.Debugf("All wildcard detection goroutines completed")

	// 计算统计信息
	result.calculateStatistics(allIPs)

	// 判断是否为泛解析
	result.IsWildcard = result.SuccessRate > 90.0 && result.IPRepeatRate > 50.0

	logger.Infof("=== Wildcard detection completed ===")
	logger.Infof("Test results:")
	logger.Infof("  - Test count: %d", result.TestCount)
	logger.Infof("  - Success count: %d", result.SuccessCount)
	logger.Infof("  - Success rate: %.2f%%", result.SuccessRate)
	logger.Infof("  - Total IPs: %d", result.TotalIPs)
	logger.Infof("  - Unique IPs: %d", result.UniqueIPs)
	logger.Infof("  - IP repeat rate: %.2f%%", result.IPRepeatRate)
	logger.Infof("  - Is wildcard: %t", result.IsWildcard)

	if result.IsWildcard {
		logger.Warnf("Domain %s has wildcard DNS enabled! Skipping brute force attack.", domain)
	} else {
		logger.Infof("Domain %s does not have wildcard DNS. Proceeding with brute force attack.", domain)
	}

	return result, nil
}

// generateDict 生成爆破字典
func (b *Brute) generateDict(domain string) ([]string, error) {
	var subdomains []string

	// 读取字典文件
	wordlist := b.wordlist
	if b.recursive {
		wordlist = b.nextlist
	}

	logger.Infof("Loading wordlist from: %s", wordlist)

	file, err := os.Open(wordlist)
	if err != nil {
		return nil, fmt.Errorf("failed to open wordlist %s: %v", wordlist, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		word := strings.TrimSpace(scanner.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}

		// 生成子域名
		subdomain := fmt.Sprintf("%s.%s", word, domain)
		subdomains = append(subdomains, subdomain)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading wordlist: %v", err)
	}

	logger.Infof("Generated %d subdomains from wordlist (read %d lines)", len(subdomains), lineCount)

	// 显示前几个子域名作为示例
	if len(subdomains) > 0 {
		sampleCount := 5
		if len(subdomains) < sampleCount {
			sampleCount = len(subdomains)
		}
		logger.Debugf("Sample subdomains: %v", subdomains[:sampleCount])
	}

	return subdomains, nil
}

// generateRandomTestSubdomains 生成随机测试子域名
func (b *Brute) generateRandomTestSubdomains(domain string, count int) ([]string, error) {
	// 读取字典文件
	words, err := b.loadWordlist()
	if err != nil {
		return nil, fmt.Errorf("failed to load wordlist: %v", err)
	}

	// 随机选择测试词
	testWords := b.selectRandomWords(words, count)

	// 生成测试子域名
	var testSubdomains []string
	for _, word := range testWords {
		testSubdomain := word + "." + domain
		testSubdomains = append(testSubdomains, testSubdomain)
	}

	return testSubdomains, nil
}

// loadWordlist 加载字典文件
func (b *Brute) loadWordlist() ([]string, error) {
	var words []string

	// 尝试加载主字典
	if b.wordlist != "" {
		file, err := os.Open(b.wordlist)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				word := strings.TrimSpace(scanner.Text())
				if word != "" {
					words = append(words, word)
				}
			}
		}
	}

	// 如果主字典为空，使用一些默认测试词
	if len(words) == 0 {
		words = []string{
			"test", "www", "mail", "ftp", "admin", "blog", "api", "dev", "stage", "prod",
			"app", "web", "cdn", "static", "img", "css", "js", "docs", "help", "support",
		}
	}

	return words, nil
}

// selectRandomWords 随机选择单词
func (b *Brute) selectRandomWords(words []string, count int) []string {
	if len(words) <= count {
		return words
	}

	// 使用crypto/rand进行真正的随机选择
	var selected []string
	used := make(map[int]bool)

	for len(selected) < count {
		// 生成随机索引
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
		if err != nil {
			logger.Errorf("Failed to generate random index: %v", err)
			continue
		}
		randomIndex := int(index.Int64())

		if !used[randomIndex] {
			selected = append(selected, words[randomIndex])
			used[randomIndex] = true
		}
	}

	return selected
}

// calculateStatistics 计算统计信息
func (r *WildcardDetectionResult) calculateStatistics(allIPs map[string]int) {
	// 计算成功率
	if r.TestCount > 0 {
		r.SuccessRate = float64(r.SuccessCount) / float64(r.TestCount) * 100
	}

	// 计算IP统计信息
	r.UniqueIPs = len(allIPs)
	r.TotalIPs = 0
	for _, count := range allIPs {
		r.TotalIPs += count
	}

	// 计算IP重复率
	if r.TotalIPs > 0 {
		r.IPRepeatRate = float64(r.TotalIPs-r.UniqueIPs) / float64(r.TotalIPs) * 100
	}
}

// bruteSubdomains 爆破子域名
func (b *Brute) bruteSubdomains(domain string, subdomains []string) error {
	logger.Infof("Starting brute force with %d subdomains, concurrency: %d", len(subdomains), b.concurrent)
	logger.Debugf("Brute force parameters:")
	logger.Debugf("  - Domain: %s", domain)
	logger.Debugf("  - Subdomains count: %d", len(subdomains))
	logger.Debugf("  - Concurrency: %d", b.concurrent)
	logger.Debugf("  - Nameservers: %v", b.nameservers)

	// 创建并发控制
	semaphore := make(chan struct{}, b.concurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 进度报告定时器
	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

	// 启动进度报告协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Panic in progress reporting: %v", r)
			}
		}()

		for range progressTicker.C {
			b.reportProgress()
		}
	}()

	logger.Debugf("Starting concurrent subdomain testing...")
	for i, subdomain := range subdomains {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(subdomain string, index int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// 添加异常处理
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic in brute force for %s: %v", subdomain, r)
					mu.Lock()
					b.processedCount++
					mu.Unlock()
				}
			}()

			// 每1000个显示一次进度
			if index%1000 == 0 {
				logger.Debugf("Processing subdomain %d/%d: %s", index+1, len(subdomains), subdomain)
			}

			// 查询子域名
			result := b.querySubdomain(subdomain)

			// 检查是否为有效子域名
			if b.isValidSubdomain(result) {
				mu.Lock()
				b.results[subdomain] = result
				b.processedCount++
				if result.Valid {
					b.successCount++
					logger.Infof("Found valid subdomain: %s (IPs: %v, CNAMEs: %v)",
						subdomain, result.IPs, result.CNAMEs)
				}
				mu.Unlock()
			} else {
				mu.Lock()
				b.processedCount++
				mu.Unlock()
			}
		}(subdomain, i)
	}

	logger.Debugf("Waiting for all goroutines to complete...")
	wg.Wait()

	// 最终进度报告
	b.reportProgress()

	logger.Debugf("Brute force subdomain testing completed")
	return nil
}

// reportProgress 报告进度
func (b *Brute) reportProgress() {
	if b.totalCount == 0 {
		return
	}

	elapsed := time.Since(b.startTime)
	progress := float64(b.processedCount) / float64(b.totalCount) * 100
	successRate := float64(b.successCount) / float64(b.processedCount) * 100
	if b.processedCount == 0 {
		successRate = 0
	}

	// 计算剩余时间
	var remainingTime time.Duration
	if b.processedCount > 0 {
		avgTimePerItem := elapsed / time.Duration(b.processedCount)
		remainingItems := b.totalCount - b.processedCount
		remainingTime = avgTimePerItem * time.Duration(remainingItems)
	}

	logger.Infof("Brute force progress: %d/%d (%.1f%%) - Success: %d (%.1f%%) - Elapsed: %v - Remaining: %v",
		b.processedCount, b.totalCount, progress,
		b.successCount, successRate,
		elapsed.Round(time.Second), remainingTime.Round(time.Second))
}

// querySubdomain 查询子域名（优化版本，只进行DNS查询）
func (b *Brute) querySubdomain(subdomain string) *BruteResult {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in querySubdomain for %s: %v", subdomain, r)
		}
	}()

	result := &BruteResult{
		Subdomain: subdomain,
		Valid:     false,
	}

	// 只查询 A 记录，有解析IP就成功
	ips, err := b.queryA(subdomain)
	if err != nil {
		logger.Debugf("No A record for %s: %v", subdomain, err)
		return result
	}

	if len(ips) > 0 {
		result.IPs = ips
		result.Valid = true
		logger.Debugf("Found A record for %s: %v", subdomain, ips)
	}

	return result
}

// queryCNAME 查询 CNAME 记录
func (b *Brute) queryCNAME(domain string) ([]string, error) {
	// 添加异常处理
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic in queryCNAME for %s: %v", domain, r)
		}
	}()

	client := new(dns.Client)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)
	msg.RecursionDesired = true

	// 遍历多个DNS服务器
	for _, nameserver := range b.nameservers {
		resp, _, err := client.Exchange(msg, nameserver+":53")
		if err != nil {
			continue
		}

		var cnames []string
		for _, answer := range resp.Answer {
			if cname, ok := answer.(*dns.CNAME); ok {
				cnames = append(cnames, strings.TrimSuffix(cname.Target, "."))
			}
		}

		if len(cnames) > 0 {
			return cnames, nil
		}
	}

	return nil, fmt.Errorf("no CNAME record found for %s", domain)
}

// isValidSubdomain 检查是否为有效子域名
func (b *Brute) isValidSubdomain(result *BruteResult) bool {
	logger.Debugf("Validating subdomain: %s", result.Subdomain)
	logger.Debugf("  - Valid: %t", result.Valid)
	logger.Debugf("  - IPs: %v", result.IPs)
	logger.Debugf("  - CNAMEs: %v", result.CNAMEs)

	// 检查是否有有效的IP或CNAME记录
	if len(result.IPs) > 0 || len(result.CNAMEs) > 0 {
		logger.Debugf("Subdomain %s is valid (has IPs or CNAMEs)", result.Subdomain)
		return true
	}

	logger.Debugf("Subdomain %s is not valid (no IPs or CNAMEs)", result.Subdomain)
	return false
}

// recursiveBrute 递归爆破
func (b *Brute) recursiveBrute(domain string) error {
	// 获取当前有效的子域名
	var validSubdomains []string
	for subdomain, result := range b.results {
		if result.Valid {
			validSubdomains = append(validSubdomains, subdomain)
		}
	}

	// 对每个有效子域名进行递归爆破
	for _, subdomain := range validSubdomains {
		if b.depth > 0 {
			// 生成下一层子域名的字典
			nextSubdomains, err := b.generateDict(subdomain)
			if err != nil {
				continue
			}

			// 爆破下一层子域名
			if err := b.bruteSubdomains(subdomain, nextSubdomains); err != nil {
				continue
			}
		}
	}

	return nil
}
