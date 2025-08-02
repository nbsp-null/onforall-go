package datasets

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// RapidDNS RapidDNS 数据集模块
type RapidDNS struct {
	*core.Query
	baseURL string
}

// NewRapidDNS 创建 RapidDNS 数据集模块
func NewRapidDNS(cfg *config.Config) *RapidDNS {
	return &RapidDNS{
		Query:   core.NewQuery("RapidDNSQuery", cfg),
		baseURL: "http://rapiddns.io/subdomain/",
	}
}

// Run 执行查询
func (r *RapidDNS) Run(domain string) ([]string, error) {
	r.SetDomain(domain)
	r.Begin()
	defer r.Finish()

	// 执行查询
	if err := r.query(domain); err != nil {
		return nil, err
	}

	return r.GetSubdomains(), nil
}

// query 执行查询
func (r *RapidDNS) query(domain string) error {
	// 设置请求头
	r.SetHeader("User-Agent", r.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("full", "1")

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s%s?%s", r.baseURL, domain, params.Encode())

	// 发送 GET 请求
	resp, err := r.HTTPGet(queryURL, r.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query RapidDNS: %v", err)
	}

	// 读取响应
	body, err := r.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := r.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		r.AddSubdomain(subdomain)
	}

	return nil
}
