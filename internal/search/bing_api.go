package search

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// BingAPI Bing API 搜索模块
type BingAPI struct {
	*core.Search
	searchURL string
	apiID     string
	apiKey    string
	limitNum  int
	delay     time.Duration
}

// BingAPIResponse Bing API 响应结构
type BingAPIResponse struct {
	WebPages struct {
		Value []struct {
			URL string `json:"url"`
		} `json:"value"`
	} `json:"webPages"`
}

// NewBingAPI 创建 Bing API 搜索模块
func NewBingAPI(cfg *config.Config) *BingAPI {
	return &BingAPI{
		Search:    core.NewSearch("BingAPISearch", cfg),
		searchURL: "https://api.bing.microsoft.com/v7.0/search",
		apiID:     cfg.APIKeys["bing_api_id"],
		apiKey:    cfg.APIKeys["bing_api_key"],
		limitNum:  1000,            // 必应同一个搜索关键词限制搜索条数
		delay:     1 * time.Second, // 必应自定义搜索限制时延1秒
	}
}

// Run 执行搜索
func (b *BingAPI) Run(domain string) ([]string, error) {
	b.SetDomain(domain)
	b.Begin()
	defer b.Finish()

	// 检查 API 密钥
	if !b.HaveAPI("bing_api_id", "bing_api_key") {
		return nil, fmt.Errorf("bing API keys not configured")
	}

	// 执行搜索
	if err := b.search(domain, ""); err != nil {
		return nil, err
	}

	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range b.Filter(domain, b.GetSubdomains()) {
		if err := b.search(domain, statement); err != nil {
			b.LogError("Failed to search with filter %s: %v", statement, err)
		}
	}

	// 递归搜索下一层的子域
	if b.IsRecursiveSearch() {
		for _, subdomain := range b.RecursiveSubdomain() {
			if err := b.search(subdomain, ""); err != nil {
				b.LogError("Failed to search subdomain %s: %v", subdomain, err)
			}
		}
	}

	return b.GetSubdomains(), nil
}

// search 执行搜索
func (b *BingAPI) search(domain, filteredSubdomain string) error {
	pageNum := 0 // 二次搜索重新置0

	for {
		// 延迟
		time.Sleep(b.delay)

		// 设置请求头
		b.SetHeader("Ocp-Apim-Subscription-Key", b.apiKey)
		b.SetHeader("User-Agent", b.GetRandomUserAgent())

		// 构建搜索查询
		query := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("q", query)
		params.Set("safesearch", "Off")
		params.Set("count", fmt.Sprintf("%d", b.GetPerPageNum()))
		params.Set("offset", fmt.Sprintf("%d", pageNum))

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", b.searchURL, params.Encode())
		resp, err := b.HTTPGet(searchURL, b.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to search: %v", err)
		}

		// 读取响应
		body, err := b.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response BingAPIResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return fmt.Errorf("failed to parse JSON response: %v", err)
		}

		// 提取子域名
		subdomains := b.ExtractSubdomains(body, domain)

		// 检查是否继续搜索
		if !b.CheckSubdomains(subdomains) {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			b.AddSubdomain(subdomain)
		}

		pageNum += b.GetPerPageNum()

		// 检查搜索条数限制
		if pageNum >= b.limitNum {
			break
		}
	}

	return nil
}
