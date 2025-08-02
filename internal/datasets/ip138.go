package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// IP138 IP138 数据集模块
type IP138 struct {
	*core.Query
	baseURL string
}

// NewIP138 创建 IP138 数据集模块
func NewIP138(cfg *config.Config) *IP138 {
	return &IP138{
		Query:   core.NewQuery("IP138Query", cfg),
		baseURL: "https://site.ip138.com/{domain}/domain.htm",
	}
}

// Run 执行查询
func (i *IP138) Run(domain string) ([]string, error) {
	i.SetDomain(domain)
	i.Begin()
	defer i.Finish()

	// 执行查询
	if err := i.query(domain); err != nil {
		return nil, err
	}

	return i.GetSubdomains(), nil
}

// query 执行查询
func (i *IP138) query(domain string) error {
	// 设置请求头
	i.SetHeader("User-Agent", i.GetRandomUserAgent())

	// 构建查询 URL
	queryURL := fmt.Sprintf("https://site.ip138.com/%s/domain.htm", domain)

	// 发送 GET 请求
	resp, err := i.HTTPGet(queryURL, i.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query IP138: %v", err)
	}

	// 读取响应
	body, err := i.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := i.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		i.AddSubdomain(subdomain)
	}

	return nil
}
