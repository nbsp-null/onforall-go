package certificates

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Racent Racent 证书模块
type Racent struct {
	*core.Query
	baseURL string
	apiKey  string
}

// NewRacent 创建 Racent 证书模块
func NewRacent(cfg *config.Config) *Racent {
	return &Racent{
		Query:   core.NewQuery("RacentQuery", cfg),
		baseURL: "https://face.racent.com/tool/query_ctlog",
		apiKey:  cfg.APIKeys["racent_api_token"],
	}
}

// Run 执行查询
func (r *Racent) Run(domain string) ([]string, error) {
	r.SetDomain(domain)
	r.Begin()
	defer r.Finish()

	// 检查 API 密钥
	if !r.HaveAPI("racent_api_token") {
		return nil, fmt.Errorf("racent_api_token not configured")
	}

	// 执行查询
	if err := r.query(domain); err != nil {
		return nil, err
	}

	return r.GetSubdomains(), nil
}

// query 执行查询
func (r *Racent) query(domain string) error {
	// 设置请求头
	r.SetHeader("User-Agent", r.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("token", r.apiKey)
	params.Set("keyword", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", r.baseURL, params.Encode())
	resp, err := r.HTTPGet(queryURL, r.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Racent: %v", err)
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
