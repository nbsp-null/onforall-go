package certificates

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Censys Censys 证书模块
type Censys struct {
	*core.Query
	baseURL string
	apiID   string
	secret  string
}

// CensysResponse Censys API 响应结构
type CensysResponse struct {
	Status string `json:"status"`
	Result struct {
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
		Hits []struct {
			Names []string `json:"names"`
		} `json:"hits"`
	} `json:"result"`
}

// NewCensys 创建 Censys 模块
func NewCensys(cfg *config.Config) *Censys {
	return &Censys{
		Query:   core.NewQuery("CensysAPIQuery", cfg),
		baseURL: "https://search.censys.io/api/v2/certificates/search",
		apiID:   cfg.APIKeys["censys_api_id"],
		secret:  cfg.APIKeys["censys_api_secret"],
	}
}

// Run 执行查询
func (c *Censys) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 检查 API 密钥
	if !c.HaveAPI("censys_api_id", "censys_api_secret") {
		return nil, fmt.Errorf("censys API keys not configured")
	}

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *Censys) query(domain string) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("q", fmt.Sprintf("names: %s", domain))
	params.Set("per_page", "100")

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Censys API: %v", err)
	}

	// 读取响应
	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var response CensysResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 检查状态
	if response.Status != "OK" {
		return fmt.Errorf("Censys API returned status: %s", response.Status)
	}

	// 提取子域名
	c.extractSubdomains(response, domain)

	// 处理分页
	nextCursor := response.Result.Links.Next
	for nextCursor != "" {
		if err := c.queryNextPage(domain, nextCursor); err != nil {
			c.LogError("Failed to query next page: %v", err)
			break
		}
	}

	return nil
}

// queryNextPage 查询下一页
func (c *Censys) queryNextPage(domain, cursor string) error {
	// 构建查询参数
	params := url.Values{}
	params.Set("q", fmt.Sprintf("names: %s", domain))
	params.Set("per_page", "100")
	params.Set("cursor", cursor)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query next page: %v", err)
	}

	// 读取响应
	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析 JSON 响应
	var response CensysResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 提取子域名
	c.extractSubdomains(response, domain)

	return nil
}

// extractSubdomains 提取子域名
func (c *Censys) extractSubdomains(response CensysResponse, domain string) {
	for _, hit := range response.Result.Hits {
		for _, name := range hit.Names {
			if c.IsValidSubdomain(name, domain) {
				c.AddSubdomain(name)
			}
		}
	}
}
