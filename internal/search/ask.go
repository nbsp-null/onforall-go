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

// Ask Ask 搜索引擎模块
type Ask struct {
	*core.Search
	searchURL  string
	limitNum   int
	perPageNum int
}

// NewAsk 创建 Ask 搜索模块
func NewAsk(cfg *config.Config) *Ask {
	return &Ask{
		Search:     core.NewSearch("AskSearch", cfg),
		searchURL:  "https://www.search.ask.com/web",
		limitNum:   200, // 限制搜索条数
		perPageNum: 10,  // 默认每页显示10页
	}
}

// Run 执行搜索
func (a *Ask) Run(domain string) ([]string, error) {
	a.SetDomain(domain)
	a.Begin()
	defer a.Finish()

	// 执行搜索
	if err := a.search(domain, ""); err != nil {
		return nil, err
	}

	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range a.Filter(domain, a.GetSubdomains()) {
		if err := a.search(domain, statement); err != nil {
			a.LogError("Failed to search with filter %s: %v", statement, err)
		}
	}

	// 递归搜索下一层的子域
	if a.IsRecursiveSearch() {
		for _, subdomain := range a.RecursiveSubdomain() {
			if err := a.search(subdomain, ""); err != nil {
				a.LogError("Failed to search subdomain %s: %v", subdomain, err)
			}
		}
	}

	return a.GetSubdomains(), nil
}

// search 执行搜索
func (a *Ask) search(domain, filteredSubdomain string) error {
	pageNum := 1

	for {
		// 延迟
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// 设置请求头
		a.SetHeader("User-Agent", a.GetRandomUserAgent())

		// 构建搜索查询
		query := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("q", query)
		params.Set("page", fmt.Sprintf("%d", pageNum))

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", a.searchURL, params.Encode())
		resp, err := a.HTTPGet(searchURL, a.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to search: %v", err)
		}

		// 读取响应
		body, err := a.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := a.ExtractSubdomains(body, domain)

		// 检查是否继续搜索
		if !a.CheckSubdomains(subdomains) {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			a.AddSubdomain(subdomain)
		}

		pageNum++

		// 检查是否有下一页
		if !strings.Contains(body, ">Next<") {
			break
		}
	}

	return nil
}
