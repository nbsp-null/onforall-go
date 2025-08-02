package datasets

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/modules"
)

// SecurityTrails SecurityTrails 数据集模块
type SecurityTrails struct {
	*modules.BaseModule
	baseURL string
}

// SecurityTrailsResponse SecurityTrails API 响应结构
type SecurityTrailsResponse struct {
	Subdomains []string `json:"subdomains"`
}

// NewSecurityTrails 创建 SecurityTrails 模块
func NewSecurityTrails(cfg *config.Config) *SecurityTrails {
	return &SecurityTrails{
		BaseModule: modules.NewBaseModule("SecurityTrails", modules.ModuleTypeDataset, cfg),
		baseURL:    "https://api.securitytrails.com/v1/domain/",
	}
}

// Run 执行查询
func (s *SecurityTrails) Run(domain string) ([]string, error) {
	s.LogInfo("Starting SecurityTrails query for domain: %s", domain)

	// 检查 API 密钥
	apiKey := s.GetAPIKey("securitytrails_api")
	if apiKey == "" {
		return nil, fmt.Errorf("SecurityTrails API key not configured")
	}

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s%s/subdomains", s.baseURL, domain)

	// 设置查询参数
	params := url.Values{}
	params.Set("apikey", apiKey)

	// 执行查询
	resp, err := s.HTTPGet(queryURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %v", err)
	}

	// 读取响应
	body, err := s.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var result SecurityTrailsResponse
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 构建完整子域名
	var subdomains []string
	for _, prefix := range result.Subdomains {
		subdomain := fmt.Sprintf("%s.%s", prefix, domain)
		subdomains = append(subdomains, subdomain)
	}

	s.LogInfo("SecurityTrails query completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}
