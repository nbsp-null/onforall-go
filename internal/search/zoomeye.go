package search

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// ZoomEye ZoomEye API 搜索模块
type ZoomEye struct {
	*core.Search
	searchURL  string
	apiKey     string
	delay      time.Duration
	perPageNum int
}

// ZoomEyeResponse ZoomEye API 响应结构
type ZoomEyeResponse struct {
	Total int `json:"total"`
}

// NewZoomEye 创建 ZoomEye API 搜索模块
func NewZoomEye(cfg *config.Config) *ZoomEye {
	return &ZoomEye{
		Search:     core.NewSearch("ZoomEyeAPISearch", cfg),
		searchURL:  "https://api.zoomeye.org/domain/search",
		apiKey:     cfg.APIKeys["zoomeye_api_key"],
		delay:      2 * time.Second,
		perPageNum: 30,
	}
}

// Run 执行搜索
func (z *ZoomEye) Run(domain string) ([]string, error) {
	z.SetDomain(domain)
	z.Begin()
	defer z.Finish()

	// 检查 API 密钥
	if !z.HaveAPI("zoomeye_api_key") {
		return nil, fmt.Errorf("zoomeye_api_key not configured")
	}

	// 执行搜索
	if err := z.search(domain); err != nil {
		return nil, err
	}

	return z.GetSubdomains(), nil
}

// search 执行搜索
func (z *ZoomEye) search(domain string) error {
	pageNum := 1
	maxRecords := 1000 // 最大记录数

	for {
		// 延迟
		time.Sleep(z.delay)

		// 设置请求头
		z.SetHeader("API-KEY", z.apiKey)
		z.SetHeader("User-Agent", z.GetRandomUserAgent())

		// 构建查询参数
		params := url.Values{}
		params.Set("q", domain)
		params.Set("page", fmt.Sprintf("%d", pageNum))
		params.Set("type", "1")

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", z.searchURL, params.Encode())
		resp, err := z.HTTPGet(searchURL, z.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query ZoomEye API: %v", err)
		}

		// 检查响应状态
		if resp.StatusCode == 403 {
			break
		}

		// 读取响应
		body, err := z.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response ZoomEyeResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return fmt.Errorf("failed to parse JSON response: %v", err)
		}

		// 提取子域名
		subdomains := z.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			z.AddSubdomain(subdomain)
		}

		// 检查是否还有更多结果
		total := response.Total
		if pageNum*z.perPageNum >= total {
			break
		}

		pageNum++

		// 检查页数限制
		if pageNum > 400 {
			break
		}

		// 检查记录数限制
		if z.perPageNum*pageNum >= maxRecords {
			break
		}
	}

	return nil
}
