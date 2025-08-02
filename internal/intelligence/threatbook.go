package intelligence

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// ThreatBook ThreatBook API 情报模块
type ThreatBook struct {
	*core.Query
	baseURL string
	key     string
}

// NewThreatBook 创建 ThreatBook API 情报模块
func NewThreatBook(cfg *config.Config) *ThreatBook {
	return &ThreatBook{
		Query:   core.NewQuery("ThreatBookAPIQuery", cfg),
		baseURL: "https://api.threatbook.cn/v3/domain/sub_domains",
		key:     cfg.APIKeys["threatbook_api_key"],
	}
}

// Run 执行查询
func (t *ThreatBook) Run(domain string) ([]string, error) {
	t.SetDomain(domain)
	t.Begin()
	defer t.Finish()

	// 检查 API 密钥
	if !t.HaveAPI("threatbook_api_key") {
		return nil, fmt.Errorf("threatbook_api_key not configured")
	}

	// 执行查询
	if err := t.query(domain); err != nil {
		return nil, err
	}

	return t.GetSubdomains(), nil
}

// query 执行查询
func (t *ThreatBook) query(domain string) error {
	// 设置请求头
	t.SetHeader("User-Agent", t.GetRandomUserAgent())

	// 构建 POST 数据
	data := url.Values{}
	data.Set("apikey", t.key)
	data.Set("resource", domain)

	// 发送 POST 请求
	resp, err := t.HTTPPost(t.baseURL, data, t.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query ThreatBook: %v", err)
	}

	// 读取响应
	body, err := t.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := t.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		t.AddSubdomain(subdomain)
	}

	return nil
}
