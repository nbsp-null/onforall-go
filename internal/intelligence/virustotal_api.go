package intelligence

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// VirusTotalAPI VirusTotal API 情报模块
type VirusTotalAPI struct {
	*core.Query
	baseURL string
	key     string
}

// VirusTotalAPIResponse VirusTotal API 响应结构
type VirusTotalAPIResponse struct {
	Meta struct {
		Cursor string `json:"cursor"`
	} `json:"meta"`
}

// NewVirusTotalAPI 创建 VirusTotal API 情报模块
func NewVirusTotalAPI(cfg *config.Config) *VirusTotalAPI {
	return &VirusTotalAPI{
		Query:   core.NewQuery("VirusTotalAPIQuery", cfg),
		baseURL: "https://www.virustotal.com/api/v3/domains/",
		key:     cfg.APIKeys["virustotal_api_key"],
	}
}

// Run 执行查询
func (v *VirusTotalAPI) Run(domain string) ([]string, error) {
	v.SetDomain(domain)
	v.Begin()
	defer v.Finish()

	// 检查 API 密钥
	if !v.HaveAPI("virustotal_api_key") {
		return nil, fmt.Errorf("virustotal_api_key not configured")
	}

	// 执行查询
	if err := v.query(domain); err != nil {
		return nil, err
	}

	return v.GetSubdomains(), nil
}

// query 执行查询
func (v *VirusTotalAPI) query(domain string) error {
	nextCursor := ""

	for {
		// 设置请求头
		v.SetHeader("User-Agent", v.GetRandomUserAgent())
		v.SetHeader("x-apikey", v.key)

		// 构建查询参数
		params := url.Values{}
		params.Set("limit", "40")
		if nextCursor != "" {
			params.Set("cursor", nextCursor)
		}

		// 构建查询 URL
		queryURL := fmt.Sprintf("%s%s/subdomains?%s", v.baseURL, domain, params.Encode())
		resp, err := v.HTTPGet(queryURL, v.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query VirusTotal API: %v", err)
		}

		// 读取响应
		body, err := v.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := v.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break // 没有发现子域名则停止查询
		}

		for _, subdomain := range subdomains {
			v.AddSubdomain(subdomain)
		}

		// 解析 JSON 响应获取下一页游标
		var response VirusTotalAPIResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			break
		}

		nextCursor = response.Meta.Cursor
		if nextCursor == "" {
			break
		}
	}

	return nil
}
