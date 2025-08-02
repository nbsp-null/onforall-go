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

// Gitee Gitee 搜索模块
type Gitee struct {
	*core.Search
	searchURL string
}

// NewGitee 创建 Gitee 搜索模块
func NewGitee(cfg *config.Config) *Gitee {
	return &Gitee{
		Search:    core.NewSearch("GiteeSearch", cfg),
		searchURL: "https://search.gitee.com/",
	}
}

// Run 执行搜索
func (g *Gitee) Run(domain string) ([]string, error) {
	g.SetDomain(domain)
	g.Begin()
	defer g.Finish()

	// 执行搜索
	if err := g.search(domain); err != nil {
		return nil, err
	}

	return g.GetSubdomains(), nil
}

// search 执行搜索
func (g *Gitee) search(domain string) error {
	pageNum := 1

	for {
		// 延迟
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// 设置请求头
		g.SetHeader("User-Agent", g.GetRandomUserAgent())

		// 构建查询参数
		params := url.Values{}
		params.Set("pageno", fmt.Sprintf("%d", pageNum))
		params.Set("q", domain)
		params.Set("type", "code")

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", g.searchURL, params.Encode())
		resp, err := g.HTTPGet(searchURL, g.GetHeader())
		if err != nil {
			g.LogError("Failed to query Gitee: %v", err)
			break
		}

		// 检查响应状态
		if resp.StatusCode != 200 {
			g.LogError("Gitee query failed with status: %d", resp.StatusCode)
			break
		}

		// 读取响应
		body, err := g.ReadResponseBody(resp)
		if err != nil {
			g.LogError("Failed to read response: %v", err)
			break
		}

		// 检查是否为空结果
		if strings.Contains(body, `class="empty-box"`) {
			break
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

		// 检查是否有下一页
		if strings.Contains(body, `<li class="disabled"><a href="###">`) {
			break
		}

		pageNum++

		// 检查页数限制
		if pageNum >= 100 {
			break
		}
	}

	return nil
}
