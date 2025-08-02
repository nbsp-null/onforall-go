package datasets

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// IPv4Info IPv4Info API 数据集模块
type IPv4Info struct {
	*core.Query
	baseURL string
	apiKey  string
}

// IPv4InfoResponse IPv4Info API 响应结构
type IPv4InfoResponse struct {
	Subdomains []string `json:"Subdomains"`
}

// NewIPv4Info 创建 IPv4Info API 数据集模块
func NewIPv4Info(cfg *config.Config) *IPv4Info {
	return &IPv4Info{
		Query:   core.NewQuery("IPv4InfoAPIQuery", cfg),
		baseURL: "http://ipv4info.com/api_v1/",
		apiKey:  cfg.APIKeys["ipv4info_api_key"],
	}
}

// Run 执行查询
func (i *IPv4Info) Run(domain string) ([]string, error) {
	i.SetDomain(domain)
	i.Begin()
	defer i.Finish()

	// 检查 API 密钥
	if !i.HaveAPI("ipv4info_api_key") {
		return nil, fmt.Errorf("ipv4info_api_key not configured")
	}

	// 执行查询
	if err := i.query(domain); err != nil {
		return nil, err
	}

	return i.GetSubdomains(), nil
}

// query 执行查询
func (i *IPv4Info) query(domain string) error {
	page := 0
	for page < 50 { // ipv4info子域查询接口最多允许查询50页
		// 设置请求头
		i.SetHeader("User-Agent", i.GetRandomUserAgent())

		// 构建查询参数
		params := url.Values{}
		params.Set("type", "SUBDOMAINS")
		params.Set("key", i.apiKey)
		params.Set("value", domain)
		params.Set("page", strconv.Itoa(page))

		// 发送 GET 请求
		queryURL := fmt.Sprintf("%s?%s", i.baseURL, params.Encode())
		resp, err := i.HTTPGet(queryURL, i.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query IPv4Info: %v", err)
		}

		// 读取响应
		body, err := i.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 解析 JSON 响应
		var response IPv4InfoResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			break
		}

		// 提取子域名
		subdomains := i.ExtractSubdomains(body, domain)
		for _, subdomain := range subdomains {
			i.AddSubdomain(subdomain)
		}

		// 检查是否还有更多数据
		if len(response.Subdomains) < 300 {
			break
		}

		page++
		time.Sleep(1 * time.Second) // 避免请求过快
	}

	return nil
}
