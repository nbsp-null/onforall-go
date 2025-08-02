package crawl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/modules"
)

// Archive Archive.org 爬虫模块
type Archive struct {
	*modules.BaseModule
	baseURL string
}

// NewArchive 创建 Archive 模块
func NewArchive(cfg *config.Config) *Archive {
	return &Archive{
		BaseModule: modules.NewBaseModule("Archive", modules.ModuleTypeCrawl, cfg),
		baseURL:    "https://web.archive.org/cdx/search/cdx",
	}
}

// Run 执行爬取
func (a *Archive) Run(domain string) ([]string, error) {
	a.LogInfo("Starting Archive.org crawl for domain: %s", domain)

	var allSubdomains []string

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

	// 执行请求
	resp, err := a.HTTPGet(queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("crawl request failed: %v", err)
	}

	// 读取响应
	body, err := a.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := a.extractSubdomains(body, domain)
	allSubdomains = append(allSubdomains, subdomains...)

	// 对发现的子域名进行递归爬取
	for _, subdomain := range subdomains {
		if subdomain != domain {
			a.LogDebug("Recursively crawling subdomain: %s", subdomain)
			recursiveSubdomains, err := a.crawlSubdomain(subdomain)
			if err != nil {
				a.LogDebug("Recursive crawl failed for %s: %v", subdomain, err)
				continue
			}
			allSubdomains = append(allSubdomains, recursiveSubdomains...)
		}
	}

	// 去重
	allSubdomains = a.Deduplicate(allSubdomains)

	a.LogInfo("Archive.org crawl completed, found %d subdomains", len(allSubdomains))
	return allSubdomains, nil
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

	// 解析 JSON 响应（简化实现）
	// 实际实现需要解析 JSON 数组
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "[]" {
			continue
		}

		// 提取 URL 中的子域名
		urls := a.extractURLs(line)
		for _, url := range urls {
			subdomain := a.extractSubdomainFromURL(url, domain)
			if subdomain != "" {
				subdomains = append(subdomains, subdomain)
			}
		}
	}

	return subdomains
}

// extractURLs 从文本中提取 URL
func (a *Archive) extractURLs(text string) []string {
	urlPattern := regexp.MustCompile(`https?://[^\s"']+`)
	return urlPattern.FindAllString(text, -1)
}

// extractSubdomainFromURL 从 URL 中提取子域名
func (a *Archive) extractSubdomainFromURL(urlStr, domain string) string {
	// 简单的子域名提取逻辑
	if strings.Contains(urlStr, domain) {
		// 提取子域名
		re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9.-]+\.%s)`, regexp.QuoteMeta(domain)))
		matches := re.FindStringSubmatch(urlStr)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// crawlSubdomain 爬取子域名
func (a *Archive) crawlSubdomain(subdomain string) ([]string, error) {
	// 限制递归深度
	params := map[string]string{
		"url":      fmt.Sprintf("*.%s/*", subdomain),
		"output":   "json",
		"fl":       "original",
		"collapse": "urlkey",
		"limit":    "1000", // 减少限制
	}

	queryURL := a.buildQueryURL(params)
	resp, err := a.HTTPGet(queryURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := a.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}

	return a.extractSubdomains(body, subdomain), nil
}
