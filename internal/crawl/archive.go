package crawl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Archive Archive.org 爬虫模块
type Archive struct {
	*core.Crawl
	baseURL string
}

// NewArchive 创建 Archive 模块
func NewArchive(cfg *config.Config) *Archive {
	return &Archive{
		Crawl:   core.NewCrawl("ArchiveCrawl", cfg),
		baseURL: "https://web.archive.org/cdx/search/cdx",
	}
}

// Run 执行爬取
func (a *Archive) Run(domain string) ([]string, error) {
	a.SetDomain(domain)
	a.Begin()
	defer a.Finish()

	// 执行爬取
	if err := a.crawl(domain); err != nil {
		return nil, err
	}

	return a.GetSubdomains(), nil
}

// crawl 执行爬取
func (a *Archive) crawl(domain string) error {
	// 设置请求头
	a.SetHeader("User-Agent", a.GetRandomUserAgent())

	// 构建查询参数
	params := map[string]string{
		"url":      fmt.Sprintf("*.%s/*", domain),
		"output":   "json",
		"fl":       "original",
		"collapse": "urlkey",
		"limit":    "10000",
	}

	// 构建查询 URL
	queryURL := a.buildQueryURL(params)

	// 发送 GET 请求
	resp, err := a.HTTPGet(queryURL, a.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to crawl Archive.org: %v", err)
	}

	// 读取响应
	body, err := a.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := a.extractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		a.AddSubdomain(subdomain)
	}

	return nil
}

// buildQueryURL 构建查询 URL
func (a *Archive) buildQueryURL(params map[string]string) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
	}
	return fmt.Sprintf("%s?%s", a.baseURL, strings.Join(queryParams, "&"))
}

// extractSubdomains 从响应中提取子域名
func (a *Archive) extractSubdomains(body, domain string) []string {
	var subdomains []string

	// 提取 URL
	urls := a.extractURLs(body)
	for _, urlStr := range urls {
		subdomain := a.extractSubdomainFromURL(urlStr, domain)
		if subdomain != "" && a.IsValidSubdomain(subdomain, domain) {
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains
}

// extractURLs 从文本中提取 URL
func (a *Archive) extractURLs(text string) []string {
	var urls []string
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlRegex.FindAllString(text, -1)
	for _, match := range matches {
		urls = append(urls, match)
	}
	return urls
}

// extractSubdomainFromURL 从 URL 中提取子域名
func (a *Archive) extractSubdomainFromURL(urlStr, domain string) string {
	// 移除协议前缀
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	// 提取域名部分
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}

	// 检查是否是目标域名的子域名
	if strings.HasSuffix(urlStr, "."+domain) {
		return urlStr
	}

	return ""
}
