package certificates

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// CertSpotter CertSpotter 证书模块
type CertSpotter struct {
	*core.Query
	baseURL string
}

// CertSpotterResponse CertSpotter API 响应结构
type CertSpotterResponse []struct {
	DNSNames []string `json:"dns_names"`
}

// NewCertSpotter 创建 CertSpotter 证书模块
func NewCertSpotter(cfg *config.Config) *CertSpotter {
	return &CertSpotter{
		Query:   core.NewQuery("CertSpotterQuery", cfg),
		baseURL: "https://api.certspotter.com/v1/issuances",
	}
}

// Run 执行查询
func (c *CertSpotter) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *CertSpotter) query(domain string) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("domain", domain)
	params.Set("include_subdomains", "true")
	params.Set("expand", "dns_names")

	// 发送查询请求
	queryURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query CertSpotter: %v", err)
	}

	// 读取响应
	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var response CertSpotterResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 提取子域名
	for _, cert := range response {
		for _, dnsName := range cert.DNSNames {
			if c.IsValidSubdomain(dnsName, domain) {
				c.AddSubdomain(dnsName)
			}
		}
	}

	return nil
}
