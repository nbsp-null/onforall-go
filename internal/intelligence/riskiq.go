package intelligence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// RiskIQ RiskIQ API 情报模块
type RiskIQ struct {
	*core.Query
	baseURL  string
	username string
	key      string
}

// RiskIQResponse RiskIQ API 响应结构
type RiskIQResponse struct {
	Subdomains []string `json:"subdomains"`
}

// NewRiskIQ 创建 RiskIQ API 情报模块
func NewRiskIQ(cfg *config.Config) *RiskIQ {
	return &RiskIQ{
		Query:    core.NewQuery("RiskIQAPIQuery", cfg),
		baseURL:  "https://api.riskiq.net/pt/v2/enrichment/subdomains",
		username: cfg.APIKeys["riskiq_api_username"],
		key:      cfg.APIKeys["riskiq_api_key"],
	}
}

// Run 执行查询
func (r *RiskIQ) Run(domain string) ([]string, error) {
	r.SetDomain(domain)
	r.Begin()
	defer r.Finish()

	// 检查 API 密钥
	if !r.HaveAPI("riskiq_api_username", "riskiq_api_key") {
		return nil, fmt.Errorf("riskiq API credentials not configured")
	}

	// 执行查询
	if err := r.query(domain); err != nil {
		return nil, err
	}

	return r.GetSubdomains(), nil
}

// query 执行查询
func (r *RiskIQ) query(domain string) error {
	// 设置请求头
	r.SetHeader("User-Agent", r.GetRandomUserAgent())
	r.SetHeader("Accept", "application/json")

	// 设置基本认证
	auth := base64.StdEncoding.EncodeToString([]byte(r.username + ":" + r.key))
	r.SetHeader("Authorization", "Basic "+auth)

	// 构建查询参数
	params := url.Values{}
	params.Set("query", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", r.baseURL, params.Encode())
	resp, err := r.HTTPGet(queryURL, r.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query RiskIQ: %v", err)
	}

	// 读取响应
	body, err := r.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var response RiskIQResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 构建子域名字符串
	subdomainStr := ""
	for _, name := range response.Subdomains {
		subdomainStr += fmt.Sprintf("%s.%s ", name, domain)
	}

	// 提取子域名
	subdomains := r.ExtractSubdomains(subdomainStr, domain)
	for _, subdomain := range subdomains {
		r.AddSubdomain(subdomain)
	}

	return nil
}
