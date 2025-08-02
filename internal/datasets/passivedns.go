package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// PassiveDNS PassiveDNS API 数据集模块
type PassiveDNS struct {
	*core.Query
	baseURL string
	token   string
}

// NewPassiveDNS 创建 PassiveDNS API 数据集模块
func NewPassiveDNS(cfg *config.Config) *PassiveDNS {
	baseURL := cfg.APIKeys["passivedns_api_addr"]
	if baseURL == "" {
		baseURL = "http://api.passivedns.cn"
	}

	return &PassiveDNS{
		Query:   core.NewQuery("PassiveDnsQuery", cfg),
		baseURL: baseURL,
		token:   cfg.APIKeys["passivedns_api_token"],
	}
}

// Run 执行查询
func (p *PassiveDNS) Run(domain string) ([]string, error) {
	p.SetDomain(domain)
	p.Begin()
	defer p.Finish()

	// 检查 API 密钥（仅对 passivedns.cn 检查）
	if p.baseURL == "http://api.passivedns.cn" {
		if !p.HaveAPI("passivedns_api_token") {
			return nil, fmt.Errorf("passivedns_api_token not configured")
		}
	}

	// 执行查询
	if err := p.query(domain); err != nil {
		return nil, err
	}

	return p.GetSubdomains(), nil
}

// query 执行查询
func (p *PassiveDNS) query(domain string) error {
	// 设置请求头
	p.SetHeader("User-Agent", p.GetRandomUserAgent())
	p.SetHeader("X-AuthToken", p.token)

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/flint/rrset/*.%s", p.baseURL, domain)

	// 发送 GET 请求
	resp, err := p.HTTPGet(queryURL, p.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query PassiveDNS: %v", err)
	}

	// 读取响应
	body, err := p.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := p.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		p.AddSubdomain(subdomain)
	}

	return nil
}
