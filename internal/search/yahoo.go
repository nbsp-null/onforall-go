package search

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Yahoo Yahoo 搜索引擎模块
type Yahoo struct {
	*core.Search
	initURL   string
	searchURL string
	limitNum  int
	delay     time.Duration
}

// NewYahoo 创建 Yahoo 搜索模块
func NewYahoo(cfg *config.Config) *Yahoo {
	return &Yahoo{
		Search:    core.NewSearch("YahooSearch", cfg),
		initURL:   "https://search.yahoo.com/",
		searchURL: "https://search.yahoo.com/search",
		limitNum:  1000, // Yahoo限制搜索条数
		delay:     2 * time.Second,
	}
}

// Run 执行搜索
func (y *Yahoo) Run(domain string) ([]string, error) {
	y.SetDomain(domain)
	y.Begin()
	defer y.Finish()

	// 执行搜索
	if err := y.search(domain, ""); err != nil {
		return nil, err
	}

	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range y.Filter(domain, y.GetSubdomains()) {
		if err := y.search(domain, statement); err != nil {
			y.LogError("Failed to search with filter %s: %v", statement, err)
		}
	}

	// 递归搜索下一层的子域
	if y.IsRecursiveSearch() {
		for _, subdomain := range y.RecursiveSubdomain() {
			if err := y.search(subdomain, ""); err != nil {
				y.LogError("Failed to search subdomain %s: %v", subdomain, err)
			}
		}
	}

	return y.GetSubdomains(), nil
}

// search 执行搜索
func (y *Yahoo) search(domain, filteredSubdomain string) error {
	pageNum := 1
	perPageNum := 30 // Yahoo每次搜索最大条数

	// 设置请求头
	y.SetHeader("User-Agent", y.GetRandomUserAgent())

	// 获取初始页面和 Cookie
	resp, err := y.HTTPGet(y.initURL, y.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to get initial page: %v", err)
	}

	// 提取 Cookie
	if resp != nil && resp.Cookies() != nil {
		for _, cookie := range resp.Cookies() {
			y.SetHeader("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
	}

	for {
		// 延迟
		time.Sleep(y.delay)

		// 构建搜索查询
		query := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("p", query)
		params.Set("b", fmt.Sprintf("%d", pageNum))
		params.Set("pz", fmt.Sprintf("%d", perPageNum))

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", y.searchURL, params.Encode())
		resp, err := y.HTTPGet(searchURL, y.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to search: %v", err)
		}

		// 读取响应
		body, err := y.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 清理 HTML 标签
		body = strings.ReplaceAll(body, "<b>", "")
		body = strings.ReplaceAll(body, "</b>", "")

		// 提取子域名
		subdomains := y.ExtractSubdomains(body, domain)

		// 检查是否继续搜索
		if !y.CheckSubdomains(subdomains) {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			y.AddSubdomain(subdomain)
		}

		// 检查是否有下一页
		if !strings.Contains(body, ">Next</a>") {
			break
		}

		pageNum += perPageNum

		// 检查搜索条数限制
		if pageNum >= y.limitNum {
			break
		}
	}

	return nil
}
