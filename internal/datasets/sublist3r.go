package datasets

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Sublist3r Sublist3r 数据集模块
type Sublist3r struct {
	*core.Query
	baseURL string
}

// NewSublist3r 创建 Sublist3r 数据集模块
func NewSublist3r(cfg *config.Config) *Sublist3r {
	return &Sublist3r{
		Query:   core.NewQuery("Sublist3rQuery", cfg),
		baseURL: "https://api.sublist3r.com/search.php",
	}
}

// Run 执行查询
func (s *Sublist3r) Run(domain string) ([]string, error) {
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
func (s *Sublist3r) query(domain string) error {
	// 设置请求头
	s.SetHeader("User-Agent", s.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("domain", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", s.baseURL, params.Encode())
	resp, err := s.HTTPGet(queryURL, s.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Sublist3r: %v", err)
	}

	// 读取响应
	body, err := s.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := s.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		s.AddSubdomain(subdomain)
	}

	return nil
}
