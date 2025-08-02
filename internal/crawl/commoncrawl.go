package crawl

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// CommonCrawl CommonCrawl 爬虫模块
type CommonCrawl struct {
	*core.Crawl
	baseURL string
}

// CommonCrawlResponse CommonCrawl API 响应结构
type CommonCrawlResponse struct {
	URL    string `json:"urlkey"`
	Status string `json:"status"`
	Text   string `json:"text"`
}

// NewCommonCrawl 创建 CommonCrawl 爬虫模块
func NewCommonCrawl(cfg *config.Config) *CommonCrawl {
	return &CommonCrawl{
		Crawl:   core.NewCrawl("CommonCrawl", cfg),
		baseURL: "https://index.commoncrawl.org/CC-MAIN-2023-50-index",
	}
}

// Run 执行爬虫
func (c *CommonCrawl) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 执行爬虫
	if err := c.crawl(domain, 50); err != nil {
		return nil, err
	}

	// 爬取已发现的子域以发现新的子域
	for _, subdomain := range c.GetSubdomains() {
		if subdomain != domain {
			if err := c.crawl(subdomain, 10); err != nil {
				c.LogError("Failed to crawl subdomain %s: %v", subdomain, err)
			}
		}
	}

	return c.GetSubdomains(), nil
}

// crawl 执行爬虫
func (c *CommonCrawl) crawl(domain string, limit int) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("url", fmt.Sprintf("*.%s/*", domain))
	params.Set("output", "json")
	params.Set("fl", "urlkey,status,text")
	params.Set("limit", strconv.Itoa(limit))

	// 发送查询请求
	queryURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query CommonCrawl: %v", err)
	}

	// 读取响应
	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应（每行一个 JSON 对象）
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var response CommonCrawlResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			continue
		}

		// 检查状态码
		if response.Status != "301" && response.Status != "302" {
			// 从文本中提取子域名
			subdomains := c.ExtractSubdomains(response.Text, domain)
			for _, subdomain := range subdomains {
				c.AddSubdomain(subdomain)
			}
		}
	}

	return nil
}
