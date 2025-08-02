package datasets

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// HackerTarget HackerTarget 数据集模块
type HackerTarget struct {
	*core.Query
	baseURL string
}

// NewHackerTarget 创建 HackerTarget 数据集模块
func NewHackerTarget(cfg *config.Config) *HackerTarget {
	return &HackerTarget{
		Query:   core.NewQuery("HackerTargetQuery", cfg),
		baseURL: "https://api.hackertarget.com/hostsearch/",
	}
}

// Run 执行查询
func (h *HackerTarget) Run(domain string) ([]string, error) {
	h.SetDomain(domain)
	h.Begin()
	defer h.Finish()

	// 执行查询
	if err := h.query(domain); err != nil {
		return nil, err
	}

	return h.GetSubdomains(), nil
}

// query 执行查询
func (h *HackerTarget) query(domain string) error {
	// 设置请求头
	h.SetHeader("User-Agent", h.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("q", domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", h.baseURL, params.Encode())
	resp, err := h.HTTPGet(queryURL, h.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query HackerTarget: %v", err)
	}

	// 读取响应
	body, err := h.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := h.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		h.AddSubdomain(subdomain)
	}

	return nil
}
