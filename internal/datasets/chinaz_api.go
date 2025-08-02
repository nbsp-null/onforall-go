package datasets

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// ChinazAPI Chinaz API 数据集模块
type ChinazAPI struct {
	*core.Query
	baseURL string
	apiKey  string
}

// NewChinazAPI 创建 Chinaz API 数据集模块
func NewChinazAPI(cfg *config.Config) *ChinazAPI {
	return &ChinazAPI{
		Query:   core.NewQuery("ChinazAPIQuery", cfg),
		baseURL: "https://apidata.chinaz.com/CallAPI/Alexa",
		apiKey:  cfg.APIKeys["chinaz_api"],
	}
}

// Run 执行查询
func (c *ChinazAPI) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 检查 API 密钥
	if !c.HaveAPI("chinaz_api") {
		return nil, fmt.Errorf("chinaz_api not configured")
	}

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *ChinazAPI) query(domain string) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("domainName", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Chinaz API: %v", err)
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
