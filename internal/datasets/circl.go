package datasets

import (
	"encoding/base64"
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Circl Circl API 数据集模块
type Circl struct {
	*core.Query
	baseURL  string
	username string
	password string
}

// NewCircl 创建 Circl API 数据集模块
func NewCircl(cfg *config.Config) *Circl {
	return &Circl{
		Query:    core.NewQuery("CirclAPIQuery", cfg),
		baseURL:  "https://www.circl.lu/pdns/query/",
		username: cfg.APIKeys["circl_api_username"],
		password: cfg.APIKeys["circl_api_password"],
	}
}

// Run 执行查询
func (c *Circl) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 检查 API 密钥
	if !c.HaveAPI("circl_api_username", "circl_api_password") {
		return nil, fmt.Errorf("circl API credentials not configured")
	}

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *Circl) query(domain string) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 设置基本认证
	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
	c.SetHeader("Authorization", "Basic "+auth)

	// 构建查询 URL
	queryURL := c.baseURL + domain

	// 发送 GET 请求
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Circl: %v", err)
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
