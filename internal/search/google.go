package search

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Google Google 搜索引擎模块
type Google struct {
	*core.Search
	initURL   string
	searchURL string
}

// NewGoogle 创建 Google 搜索模块
func NewGoogle(cfg *config.Config) *Google {
	return &Google{
		Search:    core.NewSearch("GoogleSearch", cfg),
		initURL:   "https://www.google.com/",
		searchURL: "https://www.google.com/search",
	}
}

// Run 执行搜索
func (g *Google) Run(domain string) ([]string, error) {
	g.SetDomain(domain)
	g.Begin()
	defer g.Finish()

	// 执行搜索
	if err := g.search(domain, ""); err != nil {
		return nil, err
	}

	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range g.Filter(domain, g.GetSubdomains()) {
		if err := g.search(domain, statement); err != nil {
			g.LogError("Failed to search with filter %s: %v", statement, err)
		}
	}

	// 递归搜索下一层的子域
	if g.IsRecursiveSearch() {
		for _, subdomain := range g.RecursiveSubdomain() {
			if err := g.search(subdomain, ""); err != nil {
				g.LogError("Failed to search subdomain %s: %v", subdomain, err)
			}
		}
	}

	return g.GetSubdomains(), nil
}

// search 执行搜索
func (g *Google) search(domain, filteredSubdomain string) error {
	pageNum := 1
	perPageNum := 50

	// 设置请求头
	g.SetHeader("User-Agent", "Googlebot")
	g.SetHeader("Referer", "https://www.google.com")

	// 获取初始页面和 Cookie
	resp, err := g.HTTPGet(g.initURL, g.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to get initial page: %v", err)
	}

	// 提取 Cookie
	if resp != nil && resp.Cookies() != nil {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "csrftoken" {
				g.SetHeader("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
				break
			}
		}
	}

	for {
		// 随机延迟
		delay := time.Duration(rand.Intn(5)+1) * time.Second
		time.Sleep(delay)

		// 构建搜索查询
		word := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("q", word)
		params.Set("start", fmt.Sprintf("%d", pageNum))
		params.Set("num", fmt.Sprintf("%d", perPageNum))
		params.Set("filter", "0")
		params.Set("btnG", "Search")
		params.Set("gbv", "1")
		params.Set("hl", "en")

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", g.searchURL, params.Encode())
		resp, err := g.HTTPGet(searchURL, g.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to search: %v", err)
		}

		// 读取响应
		body, err := g.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := g.ExtractSubdomains(body, domain)

		// 检查是否继续搜索
		if !g.CheckSubdomains(subdomains) {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			g.AddSubdomain(subdomain)
		}

		pageNum += perPageNum

		// 检查是否有下一页
		if !g.hasNextPage(body, pageNum) {
			break
		}

		// 检查是否被重定向
		if strings.Contains(body, "302 Moved") {
			break
		}
	}

	return nil
}

// hasNextPage 检查是否有下一页
func (g *Google) hasNextPage(body string, pageNum int) bool {
	return strings.Contains(body, fmt.Sprintf("start=%d", pageNum))
}
