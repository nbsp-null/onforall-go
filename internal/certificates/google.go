package certificates

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Google Google 证书模块
type Google struct {
	*core.Query
	baseURL string
}

// NewGoogle 创建 Google 证书模块
func NewGoogle(cfg *config.Config) *Google {
	return &Google{
		Query:   core.NewQuery("GoogleQuery", cfg),
		baseURL: "https://transparencyreport.google.com/transparencyreport/api/v3/httpsreport/ct/certsearch",
	}
}

// Run 执行查询
func (g *Google) Run(domain string) ([]string, error) {
	g.SetDomain(domain)
	g.Begin()
	defer g.Finish()

	// 执行查询
	if err := g.query(domain); err != nil {
		return nil, err
	}

	return g.GetSubdomains(), nil
}

// query 执行查询
func (g *Google) query(domain string) error {
	// 设置请求头
	g.SetHeader("User-Agent", g.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("include_expired", "true")
	params.Set("include_subdomains", "true")
	params.Set("domain", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())
	resp, err := g.HTTPGet(queryURL, g.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Google: %v", err)
	}

	// 读取响应
	body, err := g.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := g.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		g.AddSubdomain(subdomain)
	}

	return nil
}
