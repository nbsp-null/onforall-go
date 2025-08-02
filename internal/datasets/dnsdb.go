package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// DNSDB DNSDB API 数据集模块
type DNSDB struct {
	*core.Query
	baseURL string
	apiKey  string
}

// NewDNSDB 创建 DNSDB API 数据集模块
func NewDNSDB(cfg *config.Config) *DNSDB {
	return &DNSDB{
		Query:   core.NewQuery("DNSdbAPIQuery", cfg),
		baseURL: "https://api.dnsdb.info/lookup/rrset/name/",
		apiKey:  cfg.APIKeys["dnsdb_api_key"],
	}
}

// Run 执行查询
func (d *DNSDB) Run(domain string) ([]string, error) {
	d.SetDomain(domain)
	d.Begin()
	defer d.Finish()

	// 检查 API 密钥
	if !d.HaveAPI("dnsdb_api_key") {
		return nil, fmt.Errorf("dnsdb_api_key not configured")
	}

	// 执行查询
	if err := d.query(domain); err != nil {
		return nil, err
	}

	return d.GetSubdomains(), nil
}

// query 执行查询
func (d *DNSDB) query(domain string) error {
	// 设置请求头
	d.SetHeader("User-Agent", d.GetRandomUserAgent())
	d.SetHeader("X-API-Key", d.apiKey)

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s*.%s", d.baseURL, domain)

	// 发送 GET 请求
	resp, err := d.HTTPGet(queryURL, d.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query DNSDB: %v", err)
	}

	// 读取响应
	body, err := d.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := d.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		d.AddSubdomain(subdomain)
	}

	return nil
}
