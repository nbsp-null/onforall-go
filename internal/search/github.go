package search

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// GitHub GitHub API 搜索模块
type GitHub struct {
	*core.Search
	searchURL string
	apiToken  string
	delay     time.Duration
}

// GitHubResponse GitHub API 响应结构
type GitHubResponse struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	} `json:"items"`
}

// NewGitHub 创建 GitHub API 搜索模块
func NewGitHub(cfg *config.Config) *GitHub {
	return &GitHub{
		Search:    core.NewSearch("GithubAPISearch", cfg),
		searchURL: "https://api.github.com/search/code",
		apiToken:  cfg.APIKeys["github_api_token"],
		delay:     5 * time.Second,
	}
}

// Run 执行搜索
func (g *GitHub) Run(domain string) ([]string, error) {
	g.SetDomain(domain)
	g.Begin()
	defer g.Finish()

	// 检查 API 密钥
	if !g.HaveAPI("github_api_token") {
		return nil, fmt.Errorf("github_api_token not configured")
	}

	// 执行搜索
	if err := g.search(domain); err != nil {
		return nil, err
	}

	return g.GetSubdomains(), nil
}

// search 执行搜索
func (g *GitHub) search(domain string) error {
	page := 1

	// 设置请求头
	g.SetHeader("Accept", "application/vnd.github.v3.text-match+json")
	g.SetHeader("Authorization", "token "+g.apiToken)
	g.SetHeader("User-Agent", g.GetRandomUserAgent())

	for {
		// 延迟
		time.Sleep(g.delay)

		// 构建查询参数
		params := url.Values{}
		params.Set("q", domain)
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("sort", "indexed")
		params.Set("access_token", g.apiToken)

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", g.searchURL, params.Encode())
		resp, err := g.HTTPGet(searchURL, g.GetHeader())
		if err != nil {
			g.LogError("Failed to query GitHub API: %v", err)
			break
		}

		// 检查响应状态
		if resp == nil || resp.StatusCode != 200 {
			g.LogError("GitHub API query failed with status: %d", resp.StatusCode)
			break
		}

		// 读取响应
		body, err := g.ReadResponseBody(resp)
		if err != nil {
			g.LogError("Failed to read response: %v", err)
			break
		}

		// 解析 JSON 响应
		var response GitHubResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			g.LogError("Failed to parse JSON response: %v", err)
			break
		}

		// 提取子域名
		subdomains := g.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			g.AddSubdomain(subdomain)
		}

		page++

		// 检查是否还有更多结果
		totalCount := response.TotalCount
		if page*100 > totalCount {
			break
		}

		// 限制最大查询页数
		if page*100 > 1000 {
			break
		}
	}

	return nil
}
