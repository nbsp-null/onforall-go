package datasets

import (
	"encoding/json"
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// SecurityTrails SecurityTrails 数据集模块
type SecurityTrails struct {
	*core.Query
	baseURL string
	apiKey  string
}

// SecurityTrailsResponse SecurityTrails API 响应结构
type SecurityTrailsResponse struct {
	Subdomains []string `json:"subdomains"`
}

// NewSecurityTrails 创建 SecurityTrails 模块
func NewSecurityTrails(cfg *config.Config) *SecurityTrails {
	return &SecurityTrails{
		Query:   core.NewQuery("SecurityTrailsAPIQuery", cfg),
		baseURL: "https://api.securitytrails.com/v1/domain/",
		apiKey:  cfg.APIKeys["securitytrails_api"],
	}
}

// Run 执行查询
func (s *SecurityTrails) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 检查 API 密钥
	if !s.HaveAPI("securitytrails_api") {
		return nil, fmt.Errorf("securitytrails_api key not configured")
	}

	// 执行查询
	if err := s.query(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// query 执行查询
func (s *SecurityTrails) query(domain string) error {
	// 设置请求头
	s.SetHeader("APIKEY", s.apiKey)
	s.SetHeader("Content-Type", "application/json")
	s.SetHeader("User-Agent", s.GetRandomUserAgent())

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s%s/subdomains", s.baseURL, domain)

	// 发送 GET 请求
	resp, err := s.HTTPGet(queryURL, s.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query SecurityTrails API: %v", err)
	}

	// 读取响应
	body, err := s.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var response SecurityTrailsResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 添加子域名
	for _, subdomain := range response.Subdomains {
		fullSubdomain := fmt.Sprintf("%s.%s", subdomain, domain)
		if s.IsValidSubdomain(fullSubdomain, domain) {
			s.AddSubdomain(fullSubdomain)
		}
	}

	return nil
}
