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

// Fofa Fofa API 搜索模块
type Fofa struct {
	*core.Search
	searchURL string
	email     string
	apiKey    string
	delay     time.Duration
}

// FofaResponse Fofa API 响应结构
type FofaResponse struct {
	Size    int      `json:"size"`
	Results []string `json:"results"`
}

// NewFofa 创建 Fofa API 搜索模块
func NewFofa(cfg *config.Config) *Fofa {
	return &Fofa{
		Search:    core.NewSearch("FoFaAPISearch", cfg),
		searchURL: "https://fofa.info/api/v1/search/all",
		email:     cfg.APIKeys["fofa_api_email"],
		apiKey:    cfg.APIKeys["fofa_api_key"],
		delay:     1 * time.Second,
	}
}

// Run 执行搜索
func (f *Fofa) Run(domain string) ([]string, error) {
	f.SetDomain(domain)
	f.Begin()
	defer f.Finish()

	// 检查 API 密钥
	if !f.HaveAPI("fofa_api_email", "fofa_api_key") {
		return nil, fmt.Errorf("fofa API keys not configured")
	}

	// 执行搜索
	if err := f.search(domain); err != nil {
		return nil, err
	}

	return f.GetSubdomains(), nil
}

// search 执行搜索
func (f *Fofa) search(domain string) error {
	pageNum := 1
	maxRecords := 1000 // 最大记录数

	// 构建查询数据
	queryStr := fmt.Sprintf(`domain="%s"`, domain)
	queryData := base64.StdEncoding.EncodeToString([]byte(queryStr))

	for {
		// 延迟
		time.Sleep(f.delay)

		// 设置请求头
		f.SetHeader("User-Agent", f.GetRandomUserAgent())

		// 构建查询参数
		params := url.Values{}
		params.Set("email", f.email)
		params.Set("key", f.apiKey)
		params.Set("qbase64", queryData)
		params.Set("page", fmt.Sprintf("%d", pageNum))
		params.Set("full", "true")
		params.Set("size", fmt.Sprintf("%d", maxRecords))

		// 发送搜索请求
		searchURL := fmt.Sprintf("%s?%s", f.searchURL, params.Encode())
		resp, err := f.HTTPGet(searchURL, f.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query Fofa API: %v", err)
		}

		// 读取响应
		body, err := f.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response FofaResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return fmt.Errorf("failed to parse JSON response: %v", err)
		}

		// 提取子域名
		subdomains := f.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			f.AddSubdomain(subdomain)
		}

		// 检查是否还有更多结果
		size := response.Size
		if size < maxRecords {
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
