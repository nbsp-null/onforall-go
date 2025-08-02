package datasets

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Spyse Spyse API 数据集模块
type Spyse struct {
	*core.Query
	baseURL string
	token   string
}

// SpyseResponse Spyse API 响应结构
type SpyseResponse struct {
	Data struct {
		Items []interface{} `json:"items"`
	} `json:"data"`
}

// NewSpyse 创建 Spyse API 数据集模块
func NewSpyse(cfg *config.Config) *Spyse {
	return &Spyse{
		Query:   core.NewQuery("SpyseAPIQuery", cfg),
		baseURL: "https://api.spyse.com/v3/data/domain/subdomain",
		token:   cfg.APIKeys["spyse_api_token"],
	}
}

// Run 执行查询
func (s *Spyse) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 检查 API 密钥
	if !s.HaveAPI("spyse_api_token") {
		return nil, fmt.Errorf("spyse_api_token not configured")
	}

	// 执行查询
	if err := s.query(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// query 执行查询
func (s *Spyse) query(domain string) error {
	limit := 100
	offset := 0

	for {
		// 设置请求头
		s.SetHeader("User-Agent", s.GetRandomUserAgent())
		s.SetHeader("Authorization", "Bearer "+s.token)

		// 构建查询参数
		params := url.Values{}
		params.Set("domain", domain)
		params.Set("offset", fmt.Sprintf("%d", offset))
		params.Set("limit", fmt.Sprintf("%d", limit))

		// 发送 GET 请求
		queryURL := fmt.Sprintf("%s?%s", s.baseURL, params.Encode())
		resp, err := s.HTTPGet(queryURL, s.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query Spyse: %v", err)
		}

		// 读取响应
		body, err := s.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response SpyseResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			break
		}

		// 提取子域名
		subdomains := s.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break // 没有发现子域名则停止查询
		}

		for _, subdomain := range subdomains {
			s.AddSubdomain(subdomain)
		}

		// 检查是否还有更多数据
		if len(response.Data.Items) < limit {
			break
		}

		offset += limit
	}

	return nil
}
