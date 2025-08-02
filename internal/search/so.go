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

// SO SO 搜索模块
type SO struct {
	*core.Search
	searchURL  string
	limitNum   int
	perPageNum int
}

// NewSO 创建 SO 搜索模块
func NewSO(cfg *config.Config) *SO {
	return &SO{
		Search:     core.NewSearch("SoSearch", cfg),
		searchURL:  "https://www.so.com/s",
		limitNum:   640, // 限制搜索条数
		perPageNum: 10,  // 默认每页显示10页
	}
}

// Run 执行搜索
func (s *SO) Run(domain string) ([]string, error) {
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
func (s *SO) search(domain, filteredSubdomain string) error {
	pageNum := 1

	for {
		// 延迟
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// 设置请求头
		s.SetHeader("User-Agent", s.GetRandomUserAgent())

		// 构建搜索查询
		word := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		params := url.Values{}
		params.Set("q", word)
		params.Set("pn", fmt.Sprintf("%d", pageNum))

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
		if !strings.Contains(body, `<a id="snext"`) {
			break
		}

		// 检查搜索条数限制
		if pageNum*s.perPageNum >= s.limitNum {
			break
		}
	}

	return nil
}
