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

// Baidu 百度搜索引擎模块
type Baidu struct {
	*core.Search
	searchURL string
}

// NewBaidu 创建百度搜索模块
func NewBaidu(cfg *config.Config) *Baidu {
	return &Baidu{
		Search:    core.NewSearch("BaiduSearch", cfg),
		searchURL: "https://www.baidu.com/s",
	}
}

// Run 执行搜索
func (b *Baidu) Run(domain string) ([]string, error) {
	b.SetDomain(domain)
	b.Begin()
	defer b.Finish()

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
func (b *Baidu) search(domain, filteredSubdomain string) error {
	pageNum := 0
	perPageNum := 50

	// 设置请求头
	b.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	b.SetHeader("Referer", "https://www.baidu.com")

	for {
		// 随机延迟
		delay := time.Duration(rand.Intn(3)+1) * time.Second
		time.Sleep(delay)

		// 构建搜索查询
		word := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("wd", word)
		params.Set("pn", fmt.Sprintf("%d", pageNum))
		params.Set("rn", fmt.Sprintf("%d", perPageNum))
		params.Set("ie", "utf-8")
		params.Set("tn", "baiduhome_pg")

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

		pageNum += perPageNum

		// 检查是否有下一页
		if !b.hasNextPage(body, pageNum) {
			break
		}
	}

	return nil
}

// hasNextPage 检查是否有下一页
func (b *Baidu) hasNextPage(body string, pageNum int) bool {
	return strings.Contains(body, fmt.Sprintf("pn=%d", pageNum))
}
