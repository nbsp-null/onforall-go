package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// FullHunt FullHunt API 数据集模块
type FullHunt struct {
	*core.Query
	baseURL string
	apiKey  string
}

// NewFullHunt 创建 FullHunt API 数据集模块
func NewFullHunt(cfg *config.Config) *FullHunt {
	return &FullHunt{
		Query:   core.NewQuery("FullHuntAPIQuery", cfg),
		baseURL: "https://fullhunt.io/api/v1/domain/",
		apiKey:  cfg.APIKeys["fullhunt_api_key"],
	}
}

// Run 执行查询
func (f *FullHunt) Run(domain string) ([]string, error) {
	f.SetDomain(domain)
	f.Begin()
	defer f.Finish()

	// 检查 API 密钥
	if !f.HaveAPI("fullhunt_api_key") {
		return nil, fmt.Errorf("fullhunt_api_key not configured")
	}

	// 执行查询
	if err := f.query(domain); err != nil {
		return nil, err
	}

	return f.GetSubdomains(), nil
}

// query 执行查询
func (f *FullHunt) query(domain string) error {
	// 设置请求头
	f.SetHeader("User-Agent", f.GetRandomUserAgent())
	f.SetHeader("X-API-KEY", f.apiKey)

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s%s/subdomains", f.baseURL, domain)

	// 发送 GET 请求
	resp, err := f.HTTPGet(queryURL, f.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query FullHunt: %v", err)
	}

	// 读取响应
	body, err := f.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := f.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		f.AddSubdomain(subdomain)
	}

	return nil
}
