package search

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Sogou Sogou 搜索引擎模块
type Sogou struct {
	*core.Search
	searchURL  string
	limitNum   int
	perPageNum int
}

// NewSogou 创建 Sogou 搜索模块
func NewSogou(cfg *config.Config) *Sogou {
	return &Sogou{
		Search:     core.NewSearch("SogouSearch", cfg),
		searchURL:  "https://www.sogou.com/web",
		limitNum:   1000, // 限制搜索条数
		perPageNum: 10,   // 每页显示条数
	}
}

// Run 执行搜索
func (s *Sogou) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 执行搜索
	if err := s.search(domain, ""); err != nil {
		return nil, err
	}

	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range s.Filter(domain, s.GetSubdomains()) {
		if err := s.search(domain, statement); err != nil {
			s.LogError("Failed to search with filter %s: %v", statement, err)
		}
	}

	// 递归搜索下一层的子域
	if s.IsRecursiveSearch() {
		for _, subdomain := range s.RecursiveSubdomain() {
			if err := s.search(subdomain, ""); err != nil {
				s.LogError("Failed to search subdomain %s: %v", subdomain, err)
			}
		}
	}

	return s.GetSubdomains(), nil
}

// search 执行搜索
func (s *Sogou) search(domain, filteredSubdomain string) error {
	pageNum := 1

	for {
		// 设置请求头
		s.SetHeader("User-Agent", s.GetRandomUserAgent())

		// 构建搜索查询
		word := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("query", word)
		params.Set("page", fmt.Sprintf("%d", pageNum))
		params.Set("num", fmt.Sprintf("%d", s.perPageNum))

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", s.searchURL, params.Encode())
		resp, err := s.HTTPGet(searchURL, s.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to search: %v", err)
		}

		// 读取响应
		body, err := s.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := s.ExtractSubdomains(body, domain)

		// 检查是否继续搜索
		if !s.CheckSubdomains(subdomains) {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			s.AddSubdomain(subdomain)
		}

		pageNum++

		// 检查是否有下一页
		if !strings.Contains(body, "<a id=\"sogou_next\"") {
			break
		}

		// 检查搜索条数限制
		if pageNum*s.perPageNum >= s.limitNum {
			break
		}
	}

	return nil
}
