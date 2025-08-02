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

// WZSearch WZSearch 搜索模块
type WZSearch struct {
	*core.Search
	searchURL string
}

// NewWZSearch 创建 WZSearch 搜索模块
func NewWZSearch(cfg *config.Config) *WZSearch {
	return &WZSearch{
		Search:    core.NewSearch("WzSearch", cfg),
		searchURL: "https://www.wuzhuiso.com/s",
	}
}

// Run 执行搜索
func (w *WZSearch) Run(domain string) ([]string, error) {
	w.SetDomain(domain)
	w.Begin()
	defer w.Finish()

	// 执行搜索
	if err := w.search(domain, ""); err != nil {
		return nil, err
	}

	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range w.Filter(domain, w.GetSubdomains()) {
		if err := w.search(domain, statement); err != nil {
			w.LogError("Failed to search with filter %s: %v", statement, err)
		}
	}

	// 递归搜索下一层的子域
	if w.IsRecursiveSearch() {
		for _, subdomain := range w.RecursiveSubdomain() {
			if err := w.search(subdomain, ""); err != nil {
				w.LogError("Failed to search subdomain %s: %v", subdomain, err)
			}
		}
	}

	return w.GetSubdomains(), nil
}

// search 执行搜索
func (w *WZSearch) search(domain, filteredSubdomain string) error {
	pageNum := 1

	for {
		// 延迟
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// 设置请求头
		w.SetHeader("User-Agent", w.GetRandomUserAgent())

		// 构建搜索查询
		query := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("q", query)
		params.Set("pn", fmt.Sprintf("%d", pageNum))
		params.Set("src", "page_www")
		params.Set("fr", "none")

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", w.searchURL, params.Encode())
		resp, err := w.HTTPGet(searchURL, w.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to search: %v", err)
		}

		// 读取响应
		body, err := w.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := w.ExtractSubdomains(body, domain)

		// 检查是否继续搜索
		if !w.CheckSubdomains(subdomains) {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			w.AddSubdomain(subdomain)
		}

		pageNum++

		// 检查是否有下一页
		if !strings.Contains(body, `next" href`) {
			break
		}
	}

	return nil
}
