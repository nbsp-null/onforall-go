package intelligence

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// VirusTotal VirusTotal 情报模块
type VirusTotal struct {
	*core.Query
	baseURL string
}

// VirusTotalResponse VirusTotal API 响应结构
type VirusTotalResponse struct {
	Meta struct {
		Cursor string `json:"cursor"`
	} `json:"meta"`
}

// NewVirusTotal 创建 VirusTotal 情报模块
func NewVirusTotal(cfg *config.Config) *VirusTotal {
	return &VirusTotal{
		Query:   core.NewQuery("VirusTotalQuery", cfg),
		baseURL: "https://www.virustotal.com/ui/domains/",
	}
}

// Run 执行查询
func (v *VirusTotal) Run(domain string) ([]string, error) {
	v.SetDomain(domain)
	v.Begin()
	defer v.Finish()

	// 执行查询
	if err := v.query(domain); err != nil {
		return nil, err
	}

	return v.GetSubdomains(), nil
}

// query 执行查询
func (v *VirusTotal) query(domain string) error {
	nextCursor := ""

	for {
		// 设置请求头
		v.SetHeader("User-Agent", v.GetRandomUserAgent())
		v.SetHeader("Referer", "https://www.virustotal.com/")
		v.SetHeader("TE", "Trailers")

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
			return fmt.Errorf("failed to query VirusTotal: %v", err)
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
		var response VirusTotalResponse
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
