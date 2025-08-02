package certificates

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// MySSL MySSL 证书模块
type MySSL struct {
	*core.Query
	baseURL string
}

// NewMySSL 创建 MySSL 证书模块
func NewMySSL(cfg *config.Config) *MySSL {
	return &MySSL{
		Query:   core.NewQuery("MySSLQuery", cfg),
		baseURL: "https://myssl.com/api/v1/discover_sub_domain",
	}
}

// Run 执行查询
func (m *MySSL) Run(domain string) ([]string, error) {
	m.SetDomain(domain)
	m.Begin()
	defer m.Finish()

	// 执行查询
	if err := m.query(domain); err != nil {
		return nil, err
	}

	return m.GetSubdomains(), nil
}

// query 执行查询
func (m *MySSL) query(domain string) error {
	// 设置请求头
	m.SetHeader("User-Agent", m.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("domain", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", m.baseURL, params.Encode())
	resp, err := m.HTTPGet(queryURL, m.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query MySSL: %v", err)
	}

	// 读取响应
	body, err := m.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := m.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		m.AddSubdomain(subdomain)
	}

	return nil
}
