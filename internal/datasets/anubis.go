package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Anubis Anubis 数据集模块
type Anubis struct {
	*core.Query
	baseURL string
}

// NewAnubis 创建 Anubis 数据集模块
func NewAnubis(cfg *config.Config) *Anubis {
	return &Anubis{
		Query:   core.NewQuery("AnubisQuery", cfg),
		baseURL: "https://jldc.me/anubis/subdomains/",
	}
}

// Run 执行查询
func (a *Anubis) Run(domain string) ([]string, error) {
	a.SetDomain(domain)
	a.Begin()
	defer a.Finish()

	// 执行查询
	if err := a.query(domain); err != nil {
		return nil, err
	}

	return a.GetSubdomains(), nil
}

// query 执行查询
func (a *Anubis) query(domain string) error {
	// 设置请求头
	a.SetHeader("User-Agent", a.GetRandomUserAgent())

	// 构建查询 URL
	queryURL := a.baseURL + domain

	// 发送 GET 请求
	resp, err := a.HTTPGet(queryURL, a.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Anubis: %v", err)
	}

	// 读取响应
	body, err := a.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := a.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		a.AddSubdomain(subdomain)
	}

	return nil
}
