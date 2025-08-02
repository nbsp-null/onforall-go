package search

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/modules"
)

// Google 搜索引擎模块
type Google struct {
	*modules.BaseModule
	searchURL string
	initURL   string
}

// NewGoogle 创建 Google 搜索模块
func NewGoogle(cfg *config.Config) *Google {
	return &Google{
		BaseModule: modules.NewBaseModule("Google", modules.ModuleTypeSearch, cfg),
		searchURL:  "https://www.google.com/search",
		initURL:    "https://www.google.com/",
	}
}

// Run 执行搜索
func (g *Google) Run(domain string) ([]string, error) {
	g.LogInfo("Starting Google search for domain: %s", domain)

	var allSubdomains []string

	// 基本搜索
	subdomains, err := g.search(domain, "")
	if err != nil {
		g.LogError("Basic search failed: %v", err)
	} else {
		allSubdomains = append(allSubdomains, subdomains...)
	}

	// 过滤搜索
	for _, filter := range g.getFilters(domain) {
		subdomains, err := g.search(domain, filter)
		if err != nil {
			g.LogError("Filter search failed for %s: %v", filter, err)
			continue
		}
		allSubdomains = append(allSubdomains, subdomains...)
	}

	// 去重
	allSubdomains = g.Deduplicate(allSubdomains)

	g.LogInfo("Google search completed, found %d subdomains", len(allSubdomains))
	return allSubdomains, nil
}

// search 执行搜索
func (g *Google) search(domain, filteredSubdomain string) ([]string, error) {
	var subdomains []string
	pageNum := 0
	perPage := 50

	// 设置请求头
	headers := map[string]string{
		"User-Agent": "Googlebot",
		"Referer":    "https://www.google.com",
	}

	// 获取初始页面
	_, err := g.HTTPGet(g.initURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial page: %v", err)
	}

	// 搜索循环
	for {
		// 构建搜索查询
		query := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)

		// 构建请求参数
		params := url.Values{}
		params.Set("q", query)
		params.Set("start", fmt.Sprintf("%d", pageNum))
		params.Set("num", fmt.Sprintf("%d", perPage))
		params.Set("filter", "0")
		params.Set("btnG", "Search")
		params.Set("gbv", "1")
		params.Set("hl", "en")

		// 执行搜索请求
		searchURL := fmt.Sprintf("%s?%s", g.searchURL, params.Encode())
		resp, err := g.HTTPGet(searchURL, headers)
		if err != nil {
			return nil, fmt.Errorf("search request failed: %v", err)
		}

		// 读取响应
		body, err := g.ReadResponseBody(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		pageSubdomains := g.ExtractSubdomains(body, domain)
		if len(pageSubdomains) == 0 {
			break
		}

		subdomains = append(subdomains, pageSubdomains...)

		// 检查是否有下一页
		if !g.hasNextPage(body, pageNum) {
			break
		}

		pageNum += perPage
		if pageNum >= 1000 { // Google 限制
			break
		}
	}

	return subdomains, nil
}

// hasNextPage 检查是否有下一页
func (g *Google) hasNextPage(body string, pageNum int) bool {
	nextPattern := fmt.Sprintf(`start=%d`, pageNum+50)
	return strings.Contains(body, nextPattern)
}

// getFilters 获取过滤条件
func (g *Google) getFilters(domain string) []string {
	return []string{
		" -www",
		" -mail",
		" -blog",
		" -dev",
		" -test",
		" -staging",
		" -admin",
		" -api",
		" -cdn",
		" -static",
	}
}
