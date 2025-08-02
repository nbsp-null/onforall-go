package intelligence

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// AlienVault AlienVault 情报模块
type AlienVault struct {
	*core.Query
	baseURL string
}

// NewAlienVault 创建 AlienVault 情报模块
func NewAlienVault(cfg *config.Config) *AlienVault {
	return &AlienVault{
		Query:   core.NewQuery("AlienVaultQuery", cfg),
		baseURL: "https://otx.alienvault.com/api/v1/indicators/domain",
	}
}

// Run 执行查询
func (a *AlienVault) Run(domain string) ([]string, error) {
	a.SetDomain(domain)
	a.Begin()
	defer a.Finish()

	// 执行查询
	if err := a.query(domain); err != nil {
		return nil, err
	}

	return a.GetSubdomains(), nil
}

// query 执行查询
func (a *AlienVault) query(domain string) error {
	// 设置请求头
	a.SetHeader("User-Agent", a.GetRandomUserAgent())

	// 查询被动 DNS
	dnsURL := fmt.Sprintf("%s/%s/passive_dns", a.baseURL, domain)
	resp, err := a.HTTPGet(dnsURL, a.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query AlienVault DNS: %v", err)
	}

	// 读取响应
	body, err := a.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read DNS response: %v", err)
	}

	// 提取子域名
	subdomains := a.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		a.AddSubdomain(subdomain)
	}

	// 查询 URL 列表
	urlListURL := fmt.Sprintf("%s/%s/url_list", a.baseURL, domain)
	resp, err = a.HTTPGet(urlListURL, a.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query AlienVault URL list: %v", err)
	}

	// 读取响应
	body, err = a.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read URL list response: %v", err)
	}

	// 提取子域名
	subdomains = a.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		a.AddSubdomain(subdomain)
	}

	return nil
}
