package search

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Quake Quake API 搜索模块
type Quake struct {
	*core.Search
	searchURL string
	apiKey    string
	delay     time.Duration
}

// QuakeResponse Quake API 响应结构
type QuakeResponse struct {
	Meta struct {
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	} `json:"meta"`
	Data []struct {
		Service struct {
			HTTP struct {
				Host string `json:"host"`
			} `json:"http"`
		} `json:"service"`
	} `json:"data"`
}

// NewQuake 创建 Quake API 搜索模块
func NewQuake(cfg *config.Config) *Quake {
	return &Quake{
		Search:    core.NewSearch("QuakeAPISearch", cfg),
		searchURL: "https://quake.360.net/api/v3/search/quake_service",
		apiKey:    cfg.APIKeys["quake_api_key"],
		delay:     1 * time.Second,
	}
}

// Run 执行搜索
func (q *Quake) Run(domain string) ([]string, error) {
	q.SetDomain(domain)
	q.Begin()
	defer q.Finish()

	// 检查 API 密钥
	if !q.HaveAPI("quake_api_key") {
		return nil, fmt.Errorf("quake_api_key not configured")
	}

	// 执行搜索
	if err := q.search(domain); err != nil {
		return nil, err
	}

	return q.GetSubdomains(), nil
}

// search 执行搜索
func (q *Quake) search(domain string) error {
	pageNum := 0
	perPageNum := 100
	maxRecords := 1000 // 最大记录数

	for {
		// 延迟
		time.Sleep(q.delay)

		// 设置请求头
		q.SetHeader("Content-Type", "application/json")
		q.SetHeader("X-QuakeToken", q.apiKey)
		q.SetHeader("User-Agent", q.GetRandomUserAgent())

		// 构建查询数据
		query := map[string]interface{}{
			"query":   fmt.Sprintf(`domain:"%s"`, domain),
			"start":   pageNum * perPageNum,
			"size":    perPageNum,
			"include": []string{"service.http.host"},
		}

		// 发送 POST 请求
		resp, err := q.HTTPPostJSON(q.searchURL, query, q.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query Quake API: %v", err)
		}

		// 读取响应
		body, err := q.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response QuakeResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return fmt.Errorf("failed to parse JSON response: %v", err)
		}

		// 提取子域名
		subdomains := q.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			q.AddSubdomain(subdomain)
		}

		// 检查是否还有更多结果
		total := response.Meta.Pagination.Total
		if pageNum*perPageNum >= total {
			break
		}

		pageNum++

		// 检查页数限制
		if perPageNum*pageNum >= maxRecords {
			break
		}
	}

	return nil
}
