package datasets

import (
	"fmt"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// SiteDossier SiteDossier 数据集模块
type SiteDossier struct {
	*core.Query
	baseURL    string
	pageNum    int
	perPageNum int
	delay      time.Duration
}

// NewSiteDossier 创建 SiteDossier 数据集模块
func NewSiteDossier(cfg *config.Config) *SiteDossier {
	return &SiteDossier{
		Query:      core.NewQuery("SiteDossierQuery", cfg),
		baseURL:    "http://www.sitedossier.com/parentdomain/",
		pageNum:    1,
		perPageNum: 100,
		delay:      2 * time.Second,
	}
}

// Run 执行查询
func (s *SiteDossier) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 执行查询
	if err := s.query(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// query 执行查询
func (s *SiteDossier) query(domain string) error {
	for {
		// 设置请求头
		s.SetHeader("User-Agent", s.GetRandomUserAgent())

		// 构建查询 URL
		queryURL := fmt.Sprintf("%s%s/%d", s.baseURL, domain, s.pageNum)

		// 发送 GET 请求
		resp, err := s.HTTPGet(queryURL, s.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query SiteDossier: %v", err)
		}

		// 读取响应
		body, err := s.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := s.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break // 没有发现子域名则停止查询
		}

		for _, subdomain := range subdomains {
			s.AddSubdomain(subdomain)
		}

		// 检查是否有下一页
		if !strings.Contains(body, "Show next 100 items") {
			break // 搜索页面没有出现下一页时停止搜索
		}

		s.pageNum += s.perPageNum
		time.Sleep(s.delay) // 避免请求过快
	}

	return nil
}
