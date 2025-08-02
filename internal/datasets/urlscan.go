package datasets

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// URLScan URLScan 数据集模块
type URLScan struct {
	*core.Query
	baseURL string
}

// NewURLScan 创建 URLScan 数据集模块
func NewURLScan(cfg *config.Config) *URLScan {
	return &URLScan{
		Query:   core.NewQuery("UrlscanQuery", cfg),
		baseURL: "https://urlscan.io/api/v1/search/",
	}
}

// Run 执行查询
func (u *URLScan) Run(domain string) ([]string, error) {
	u.SetDomain(domain)
	u.Begin()
	defer u.Finish()

	// 执行查询
	if err := u.query(domain); err != nil {
		return nil, err
	}

	return u.GetSubdomains(), nil
}

// query 执行查询
func (u *URLScan) query(domain string) error {
	// 设置请求头
	u.SetHeader("User-Agent", u.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("q", "domain:"+domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", u.baseURL, params.Encode())
	resp, err := u.HTTPGet(queryURL, u.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query URLScan: %v", err)
	}

	// 读取响应
	body, err := u.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := u.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		u.AddSubdomain(subdomain)
	}

	return nil
}
