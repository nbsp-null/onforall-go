package brute

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oneforall-go/internal/dns"
	"github.com/oneforall-go/pkg/logger"
	"golang.org/x/sync/semaphore"
)

// Subdomain 子域信息结构（避免循环导入）
type Subdomain struct {
	Subdomain string   `json:"subdomain" csv:"subdomain"`
	IP        []string `json:"ip" csv:"ip"`
	Status    int      `json:"status" csv:"status"`
	Title     string   `json:"title" csv:"title"`
	Port      int      `json:"port" csv:"port"`
	Alive     bool     `json:"alive" csv:"alive"`
	Source    string   `json:"source" csv:"source"`
	Time      string   `json:"time" csv:"time"`
}

// Client 暴力破解客户端
type Client struct {
	concurrency int64
	timeout     time.Duration
	semaphore   *semaphore.Weighted
	dnsClient   *dns.Client
	wordlists   []string
}

// NewClient 创建新的暴力破解客户端
func NewClient(concurrency int, timeout int) *Client {
	return &Client{
		concurrency: int64(concurrency),
		timeout:     time.Duration(timeout) * time.Second,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		dnsClient:   dns.NewClient(timeout, concurrency),
		wordlists:   getDefaultWordlists(),
	}
}

// Brute 执行暴力破解
func (c *Client) Brute(domain string) ([]Subdomain, error) {
	logger.Infof("Starting brute force for domain: %s", domain)

	// 加载字典
	words, err := c.loadWordlists()
	if err != nil {
		return nil, fmt.Errorf("failed to load wordlists: %v", err)
	}

	logger.Infof("Loaded %d words for brute force", len(words))

	// 创建结果通道
	results := make(chan Subdomain, len(words))
	var wg sync.WaitGroup

	// 启动工作协程
	for _, word := range words {
		wg.Add(1)
		go func(subdomain string) {
			defer wg.Done()
			c.bruteSubdomain(domain, subdomain, results)
		}(word)
	}

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	subdomains := make([]Subdomain, 0)
	for subdomain := range results {
		subdomains = append(subdomains, subdomain)
	}

	logger.Infof("Brute force completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}

// bruteSubdomain 暴力破解单个子域
func (c *Client) bruteSubdomain(domain, subdomain string, results chan<- Subdomain) {
	// 获取信号量
	if err := c.semaphore.Acquire(context.Background(), 1); err != nil {
		logger.Debugf("Failed to acquire semaphore for %s: %v", subdomain, err)
		return
	}
	defer c.semaphore.Release(1)

	// 构造完整域名
	fullDomain := subdomain + "." + domain

	// DNS 解析
	ips, err := c.dnsClient.Resolve(fullDomain)
	if err != nil {
		logger.Debugf("Failed to resolve %s: %v", fullDomain, err)
		return
	}

	// 如果解析成功，添加到结果
	if len(ips) > 0 {
		result := Subdomain{
			Subdomain: fullDomain,
			IP:        ips,
			Source:    "brute",
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
		results <- result
		logger.Debugf("Found subdomain: %s -> %v", fullDomain, ips)
	}
}

// loadWordlists 加载字典文件
func (c *Client) loadWordlists() ([]string, error) {
	words := make([]string, 0)
	seen := make(map[string]bool)

	for _, wordlist := range c.wordlists {
		fileWords, err := c.loadWordlist(wordlist)
		if err != nil {
			logger.Warnf("Failed to load wordlist %s: %v", wordlist, err)
			continue
		}

		// 去重
		for _, word := range fileWords {
			word = strings.TrimSpace(word)
			if word != "" && !seen[word] {
				seen[word] = true
				words = append(words, word)
			}
		}
	}

	return words, nil
}

// loadWordlist 加载单个字典文件
func (c *Client) loadWordlist(filename string) ([]string, error) {
	// 尝试多个路径
	paths := []string{
		filename,
		filepath.Join("data", "wordlists", filename),
		filepath.Join("wordlists", filename),
	}

	var file *os.File
	var err error

	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			break
		}
	}

	if file == nil {
		return nil, fmt.Errorf("failed to open wordlist file: %s", filename)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			words = append(words, word)
		}
	}

	return words, scanner.Err()
}

// getDefaultWordlists 获取默认字典文件列表
func getDefaultWordlists() []string {
	return []string{
		"subnames.txt",
		"subnames_medium.txt",
		"subnames_next.txt",
	}
}

// SetWordlists 设置字典文件列表
func (c *Client) SetWordlists(wordlists []string) {
	c.wordlists = wordlists
}

// GetWordlists 获取字典文件列表
func (c *Client) GetWordlists() []string {
	return c.wordlists
}

// SetConcurrency 设置并发数
func (c *Client) SetConcurrency(concurrency int) {
	c.concurrency = int64(concurrency)
	c.semaphore = semaphore.NewWeighted(int64(concurrency))
}

// GetConcurrency 获取并发数
func (c *Client) GetConcurrency() int {
	return int(c.concurrency)
}

// SetTimeout 设置超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取超时时间
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}
