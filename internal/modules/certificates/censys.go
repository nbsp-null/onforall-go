package certificates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/modules"
)

// Censys Censys 证书模块
type Censys struct {
	*modules.BaseModule
	baseURL string
}

// CensysResponse Censys API 响应结构
type CensysResponse struct {
	Status string `json:"status"`
	Result struct {
		Hits []struct {
			Names []string `json:"names"`
		} `json:"hits"`
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
	} `json:"result"`
}

// NewCensys 创建 Censys 模块
func NewCensys(cfg *config.Config) *Censys {
	return &Censys{
		BaseModule: modules.NewBaseModule("Censys", modules.ModuleTypeCertificate, cfg),
		baseURL:    "https://search.censys.io/api/v2/certificates/search",
	}
}

// Run 执行查询
func (c *Censys) Run(domain string) ([]string, error) {
	c.LogInfo("Starting Censys query for domain: %s", domain)

	// 检查 API 密钥
	apiID := c.GetAPIKey("censys_api_id")
	apiSecret := c.GetAPIKey("censys_api_secret")
	if apiID == "" || apiSecret == "" {
		return nil, fmt.Errorf("Censys API credentials not configured")
	}

	var allSubdomains []string
	nextCursor := ""

	for {
		// 构建查询参数
		params := url.Values{}
		params.Set("q", fmt.Sprintf("names: %s", domain))
		params.Set("per_page", "100")
		if nextCursor != "" {
			params.Set("cursor", nextCursor)
		}

		// 构建请求
		req, err := http.NewRequest("GET", c.baseURL+"?"+params.Encode(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// 设置认证头
		req.SetBasicAuth(apiID, apiSecret)
		req.Header.Set("Accept", "application/json")

		// 执行请求
		resp, err := c.GetHTTPClient().Do(req)
		if err != nil {
			return nil, fmt.Errorf("query request failed: %v", err)
		}

		// 读取响应
		body, err := c.ReadResponseBody(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var result CensysResponse
		if err := json.Unmarshal([]byte(body), &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %v", err)
		}

		// 检查状态
		if result.Status != "OK" {
			c.LogError("Censys API returned status: %s", result.Status)
			break
		}

		// 提取子域名
		for _, hit := range result.Result.Hits {
			for _, name := range hit.Names {
				if c.IsValidSubdomain(name, domain) {
					allSubdomains = append(allSubdomains, name)
				}
			}
		}

		// 检查是否有下一页
		nextCursor = result.Result.Links.Next
		if nextCursor == "" {
			break
		}

		// 延迟以避免速率限制
		c.Sleep()
	}

	// 去重
	allSubdomains = c.Deduplicate(allSubdomains)

	c.LogInfo("Censys query completed, found %d subdomains", len(allSubdomains))
	return allSubdomains, nil
}
