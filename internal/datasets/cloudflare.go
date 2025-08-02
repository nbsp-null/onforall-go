package datasets

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Cloudflare Cloudflare API 数据集模块
type Cloudflare struct {
	*core.Query
	baseURL string
	token   string
}

// CloudflareResponse Cloudflare API 响应结构
type CloudflareResponse struct {
	Success bool `json:"success"`
	Result  []struct {
		ID string `json:"id"`
	} `json:"result"`
	ResultInfo struct {
		TotalPages int `json:"total_pages"`
	} `json:"result_info"`
}

// CloudflareZoneResponse Cloudflare Zone 响应结构
type CloudflareZoneResponse struct {
	Success bool `json:"success"`
	Result  []struct {
		ID string `json:"id"`
	} `json:"result"`
}

// CloudflareCreateZoneResponse Cloudflare 创建 Zone 响应结构
type CloudflareCreateZoneResponse struct {
	Success bool `json:"success"`
	Result  struct {
		ID string `json:"id"`
	} `json:"result"`
}

// NewCloudflare 创建 Cloudflare API 数据集模块
func NewCloudflare(cfg *config.Config) *Cloudflare {
	return &Cloudflare{
		Query:   core.NewQuery("CloudFlareAPIQuery", cfg),
		baseURL: "https://api.cloudflare.com/client/v4/",
		token:   cfg.APIKeys["cloudflare_api_token"],
	}
}

// Run 执行查询
func (c *Cloudflare) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 检查 API 密钥
	if !c.HaveAPI("cloudflare_api_token") {
		return nil, fmt.Errorf("cloudflare_api_token not configured")
	}

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *Cloudflare) query(domain string) error {
	// 设置请求头
	c.SetHeader("Authorization", "Bearer "+c.token)
	c.SetHeader("Content-Type", "application/json")
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 获取账户 ID
	accountID, err := c.getAccountID()
	if err != nil {
		return fmt.Errorf("failed to get account ID: %v", err)
	}

	// 查询域名区域
	zoneID, err := c.getZoneID(domain, accountID)
	if err != nil {
		return fmt.Errorf("failed to get zone ID: %v", err)
	}

	if zoneID != "" {
		// 列出 DNS 记录
		if err := c.listDNS(zoneID); err != nil {
			return fmt.Errorf("failed to list DNS records: %v", err)
		}
	}

	return nil
}

// getAccountID 获取账户 ID
func (c *Cloudflare) getAccountID() (string, error) {
	resp, err := c.HTTPGet(c.baseURL+"accounts", c.GetHeader())
	if err != nil {
		return "", err
	}

	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return "", err
	}

	var response CloudflareResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return "", err
	}

	if !response.Success || len(response.Result) == 0 {
		return "", fmt.Errorf("no account found")
	}

	return response.Result[0].ID, nil
}

// getZoneID 获取区域 ID
func (c *Cloudflare) getZoneID(domain, accountID string) (string, error) {
	params := url.Values{}
	params.Set("name", domain)

	resp, err := c.HTTPGet(c.baseURL+"zones?"+params.Encode(), c.GetHeader())
	if err != nil {
		return "", err
	}

	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return "", err
	}

	var response CloudflareZoneResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("failed to get zone")
	}

	if len(response.Result) == 0 {
		// 创建区域
		return c.createZone(domain, accountID)
	}

	return response.Result[0].ID, nil
}

// createZone 创建区域
func (c *Cloudflare) createZone(domain, accountID string) (string, error) {
	data := map[string]interface{}{
		"name":       domain,
		"account":    map[string]string{"id": accountID},
		"jump_start": true,
		"type":       "full",
	}

	resp, err := c.HTTPPostJSON(c.baseURL+"zones", data, c.GetHeader())
	if err != nil {
		return "", err
	}

	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return "", err
	}

	var response CloudflareCreateZoneResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("failed to create zone")
	}

	return response.Result.ID, nil
}

// listDNS 列出 DNS 记录
func (c *Cloudflare) listDNS(zoneID string) error {
	page := 1
	for {
		params := url.Values{}
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("per_page", "10")

		resp, err := c.HTTPGet(c.baseURL+fmt.Sprintf("zones/%s/dns_records?%s", zoneID, params.Encode()), c.GetHeader())
		if err != nil {
			return err
		}

		body, err := c.ReadResponseBody(resp)
		if err != nil {
			return err
		}

		// 提取子域名
		subdomains := c.ExtractSubdomains(body, c.GetDomain())
		for _, subdomain := range subdomains {
			c.AddSubdomain(subdomain)
		}

		// 解析响应获取总页数
		var response CloudflareResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			break
		}

		if page >= response.ResultInfo.TotalPages {
			break
		}

		page++
		time.Sleep(1 * time.Second) // 避免请求过快
	}

	return nil
}
