package search

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// GoogleAPI Google API 搜索模块
type GoogleAPI struct {
	*core.Search
	searchURL  string
	apiKey     string
	apiID      string
	delay      time.Duration
	perPageNum int
}

// GoogleAPIResponse Google API 响应结构
type GoogleAPIResponse struct {
	Items []struct {
		Link string `json:"link"`
	} `json:"items"`
}

// NewGoogleAPI 创建 Google API 搜索模块
func NewGoogleAPI(cfg *config.Config) *GoogleAPI {
	return &GoogleAPI{
		Search:     core.NewSearch("GoogleAPISearch", cfg),
		searchURL:  "https://www.googleapis.com/customsearch/v1",
		apiKey:     cfg.APIKeys["google_api_key"],
		apiID:      cfg.APIKeys["google_api_id"],
		delay:      1 * time.Second,
		perPageNum: 10, // 每次只能请求10个结果
	}
}

// Run 执行搜索
func (g *GoogleAPI) Run(domain string) ([]string, error) {
	g.SetDomain(domain)
	g.Begin()
	defer g.Finish()

	// 检查 API 密钥
	if !g.HaveAPI("google_api_id", "google_api_key") {
		return nil, fmt.Errorf("google API keys not configured")
	}

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
func (g *GoogleAPI) search(domain, filteredSubdomain string) error {
	pageNum := 1

	for {
		// 延迟
		time.Sleep(g.delay)

		// 设置请求头
		g.SetHeader("User-Agent", g.GetRandomUserAgent())

		// 构建搜索查询
		word := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("key", g.apiKey)
		params.Set("cx", g.apiID)
		params.Set("q", word)
		params.Set("fields", "items/link")
		params.Set("start", fmt.Sprintf("%d", pageNum))
		params.Set("num", fmt.Sprintf("%d", g.perPageNum))

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

		// 解析 JSON 响应
		var response GoogleAPIResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return fmt.Errorf("failed to parse JSON response: %v", err)
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

		pageNum += g.perPageNum

		// 检查页数限制（免费的API只能查询前100条结果）
		if pageNum > 100 {
			break
		}
	}

	return nil
}
