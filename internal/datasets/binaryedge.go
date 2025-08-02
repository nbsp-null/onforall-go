package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// BinaryEdge BinaryEdge API 数据集模块
type BinaryEdge struct {
	*core.Query
	baseURL string
	apiKey  string
}

// NewBinaryEdge 创建 BinaryEdge API 数据集模块
func NewBinaryEdge(cfg *config.Config) *BinaryEdge {
	return &BinaryEdge{
		Query:   core.NewQuery("BinaryEdgeAPIQuery", cfg),
		baseURL: "https://api.binaryedge.io/v2/query/domains/subdomain/",
		apiKey:  cfg.APIKeys["binaryedge_api"],
	}
}

// Run 执行查询
func (b *BinaryEdge) Run(domain string) ([]string, error) {
	b.SetDomain(domain)
	b.Begin()
	defer b.Finish()

	// 检查 API 密钥
	if !b.HaveAPI("binaryedge_api") {
		return nil, fmt.Errorf("binaryedge_api not configured")
	}

	// 执行查询
	if err := b.query(domain); err != nil {
		return nil, err
	}

	return b.GetSubdomains(), nil
}

// query 执行查询
func (b *BinaryEdge) query(domain string) error {
	// 设置请求头
	b.SetHeader("User-Agent", b.GetRandomUserAgent())
	b.SetHeader("X-Key", b.apiKey)

	// 构建查询 URL
	queryURL := b.baseURL + domain

	// 发送 GET 请求
	resp, err := b.HTTPGet(queryURL, b.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query BinaryEdge: %v", err)
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
