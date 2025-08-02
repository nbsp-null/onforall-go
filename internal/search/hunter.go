package search

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Hunter Hunter API 搜索模块
type Hunter struct {
	*core.Search
	searchURL string
	apiKey    string
	delay     time.Duration
}

// HunterResponse Hunter API 响应结构
type HunterResponse struct {
	Data struct {
		Total int `json:"total"`
	} `json:"data"`
}

// NewHunter 创建 Hunter API 搜索模块
func NewHunter(cfg *config.Config) *Hunter {
	return &Hunter{
		Search:    core.NewSearch("HunterAPISearch", cfg),
		searchURL: "https://hunter.qianxin.com/openApi/search",
		apiKey:    cfg.APIKeys["hunter_api_key"],
		delay:     1 * time.Second,
	}
}

// Run 执行搜索
func (h *Hunter) Run(domain string) ([]string, error) {
	h.SetDomain(domain)
	h.Begin()
	defer h.Finish()

	// 检查 API 密钥
	if !h.HaveAPI("hunter_api_key") {
		return nil, fmt.Errorf("hunter_api_key not configured")
	}

	// 执行搜索
	if err := h.search(domain); err != nil {
		return nil, err
	}

	return h.GetSubdomains(), nil
}

// search 执行搜索
func (h *Hunter) search(domain string) error {
	pageNum := 1
	maxRecords := 1000 // 最大记录数

	// 构建查询数据
	queryStr := fmt.Sprintf(`domain_suffix="%s"`, domain)
	queryData := base64.StdEncoding.EncodeToString([]byte(queryStr))

	for {
		// 延迟
		time.Sleep(h.delay)

		// 设置请求头
		h.SetHeader("User-Agent", h.GetRandomUserAgent())

		// 构建查询参数
		params := url.Values{}
		params.Set("api-key", h.apiKey)
		params.Set("search", queryData)
		params.Set("page", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", "100")
		params.Set("is_web", "1")

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", h.searchURL, params.Encode())
		resp, err := h.HTTPGet(searchURL, h.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query Hunter API: %v", err)
		}

		// 读取响应
		body, err := h.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response HunterResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return fmt.Errorf("failed to parse JSON response: %v", err)
		}

		// 提取子域名
		subdomains := h.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			h.AddSubdomain(subdomain)
		}

		// 检查是否还有更多结果
		total := response.Data.Total
		if pageNum*100 >= total {
			break
		}

		pageNum++

		// 检查页数限制
		if 100*pageNum >= maxRecords {
			break
		}
	}

	return nil
}
