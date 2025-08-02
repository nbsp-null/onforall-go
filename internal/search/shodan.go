package search

import (
	"encoding/json"
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Shodan Shodan API 搜索模块
type Shodan struct {
	*core.Search
	searchURL string
	apiKey    string
}

// ShodanResponse Shodan API 响应结构
type ShodanResponse struct {
	Subdomains []string `json:"subdomains"`
}

// NewShodan 创建 Shodan API 搜索模块
func NewShodan(cfg *config.Config) *Shodan {
	return &Shodan{
		Search:    core.NewSearch("ShodanAPISearch", cfg),
		searchURL: "https://api.shodan.io/dns/domain/",
		apiKey:    cfg.APIKeys["shodan_api_key"],
	}
}

// Run 执行搜索
func (s *Shodan) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 检查 API 密钥
	if !s.HaveAPI("shodan_api_key") {
		return nil, fmt.Errorf("shodan_api_key not configured")
	}

	// 执行搜索
	if err := s.search(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// search 执行搜索
func (s *Shodan) search(domain string) error {
	// 设置请求头
	s.SetHeader("User-Agent", s.GetRandomUserAgent())

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s%s?key=%s", s.searchURL, domain, s.apiKey)

	// 发送 GET 请求
	resp, err := s.HTTPGet(queryURL, s.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Shodan API: %v", err)
	}

	// 读取响应
	body, err := s.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var response ShodanResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 处理子域名
	for _, subdomain := range response.Subdomains {
		fullSubdomain := fmt.Sprintf("%s.%s", subdomain, domain)
		if s.IsValidSubdomain(fullSubdomain, domain) {
			s.AddSubdomain(fullSubdomain)
		}
	}

	return nil
}
