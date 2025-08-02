package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// BeVigil BeVigil 数据集模块
type BeVigil struct {
	*core.Query
	baseURL string
	apiKey  string
}

// NewBeVigil 创建 BeVigil 数据集模块
func NewBeVigil(cfg *config.Config) *BeVigil {
	return &BeVigil{
		Query:   core.NewQuery("BeVigilOsintApi", cfg),
		baseURL: "http://osint.bevigil.com/api/{}/subdomains/",
		apiKey:  cfg.APIKeys["bevigil_api"],
	}
}

// Run 执行查询
func (b *BeVigil) Run(domain string) ([]string, error) {
	b.SetDomain(domain)
	b.Begin()
	defer b.Finish()

	// 检查 API 密钥
	if !b.HaveAPI("bevigil_api") {
		return nil, fmt.Errorf("bevigil_api not configured")
	}

	// 执行查询
	if err := b.query(domain); err != nil {
		return nil, err
	}

	return b.GetSubdomains(), nil
}

// query 执行查询
func (b *BeVigil) query(domain string) error {
	// 设置请求头
	b.SetHeader("User-Agent", b.GetRandomUserAgent())
	b.SetHeader("X-Access-Token", b.apiKey)

	// 构建查询 URL
	queryURL := fmt.Sprintf("http://osint.bevigil.com/api/%s/subdomains/", domain)

	// 发送 GET 请求
	resp, err := b.HTTPGet(queryURL, b.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query BeVigil: %v", err)
	}

	// 读取响应
	body, err := b.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := b.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		b.AddSubdomain(subdomain)
	}

	return nil
}
