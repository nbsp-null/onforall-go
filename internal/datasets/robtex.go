package datasets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Robtex Robtex 数据集模块
type Robtex struct {
	*core.Query
	baseURL string
}

// RobtexRecord Robtex 记录结构
type RobtexRecord struct {
	Rrtype string `json:"rrtype"`
	Rrdata string `json:"rrdata"`
}

// NewRobtex 创建 Robtex 数据集模块
func NewRobtex(cfg *config.Config) *Robtex {
	return &Robtex{
		Query:   core.NewQuery("RobtexQuery", cfg),
		baseURL: "https://freeapi.robtex.com/pdns",
	}
}

// Run 执行查询
func (r *Robtex) Run(domain string) ([]string, error) {
	r.SetDomain(domain)
	r.Begin()
	defer r.Finish()

	// 执行查询
	if err := r.query(domain); err != nil {
		return nil, err
	}

	return r.GetSubdomains(), nil
}

// query 执行查询
func (r *Robtex) query(domain string) error {
	// 设置请求头
	r.SetHeader("User-Agent", r.GetRandomUserAgent())

	// 查询正向 DNS 记录
	forwardURL := fmt.Sprintf("%s/forward/%s", r.baseURL, domain)
	resp, err := r.HTTPGet(forwardURL, r.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Robtex forward: %v", err)
	}

	// 读取响应
	body, err := r.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read forward response: %v", err)
	}

	// 解析 JSON 记录
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var record RobtexRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		// 只处理 A 和 AAAA 记录
		if record.Rrtype == "A" || record.Rrtype == "AAAA" {
			time.Sleep(1 * time.Second) // Robtex有查询频率限制

			// 查询反向 DNS 记录
			reverseURL := fmt.Sprintf("%s/reverse/%s", r.baseURL, record.Rrdata)
			resp, err := r.HTTPGet(reverseURL, r.GetHeader())
			if err != nil {
				continue
			}

			// 读取响应
			reverseBody, err := r.ReadResponseBody(resp)
			if err != nil {
				continue
			}

			// 提取子域名
			subdomains := r.ExtractSubdomains(reverseBody, domain)
			for _, subdomain := range subdomains {
				r.AddSubdomain(subdomain)
			}
		}
	}

	return nil
}
