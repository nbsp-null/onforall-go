package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// DNSGrep DNSGrep 数据集模块
type DNSGrep struct {
	*core.Query
	baseURL string
}

// NewDNSGrep 创建 DNSGrep 数据集模块
func NewDNSGrep(cfg *config.Config) *DNSGrep {
	return &DNSGrep{
		Query:   core.NewQuery("DnsgrepQuery", cfg),
		baseURL: "https://www.dnsgrep.cn/subdomain/",
	}
}

// Run 执行查询
func (d *DNSGrep) Run(domain string) ([]string, error) {
	d.SetDomain(domain)
	d.Begin()
	defer d.Finish()

	// 执行查询
	if err := d.query(domain); err != nil {
		return nil, err
	}

	return d.GetSubdomains(), nil
}

// query 执行查询
func (d *DNSGrep) query(domain string) error {
	// 设置请求头
	d.SetHeader("User-Agent", d.GetRandomUserAgent())

	// 构建查询 URL
	queryURL := d.baseURL + domain

	// 发送 GET 请求
	resp, err := d.HTTPGet(queryURL, d.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query DNSGrep: %v", err)
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
