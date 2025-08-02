package datasets

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Chinaz Chinaz 数据集模块
type Chinaz struct {
	*core.Query
	baseURL string
}

// NewChinaz 创建 Chinaz 数据集模块
func NewChinaz(cfg *config.Config) *Chinaz {
	return &Chinaz{
		Query:   core.NewQuery("ChinazQuery", cfg),
		baseURL: "https://alexa.chinaz.com/",
	}
}

// Run 执行查询
func (c *Chinaz) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *Chinaz) query(domain string) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 构建查询 URL
	queryURL := c.baseURL + domain

	// 发送 GET 请求
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Chinaz: %v", err)
	}

	// 读取响应
	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := c.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		c.AddSubdomain(subdomain)
	}

	return nil
}
