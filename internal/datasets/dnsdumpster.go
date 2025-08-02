package datasets

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// DNSDumpster DNSDumpster 数据集模块
type DNSDumpster struct {
	*core.Query
	baseURL string
}

// NewDNSDumpster 创建 DNSDumpster 模块
func NewDNSDumpster(cfg *config.Config) *DNSDumpster {
	return &DNSDumpster{
		Query:   core.NewQuery("DNSDumpsterQuery", cfg),
		baseURL: "https://dnsdumpster.com/",
	}
}

// Run 执行查询
func (d *DNSDumpster) Run(domain string) ([]string, error) {
	d.SetDomain(domain)
	d.Begin()
	defer d.Finish()

	// 执行查询
	if err := d.query(domain); err != nil {
		return nil, err
	}

	return d.GetSubdomains(), nil
}

// query 执行查询
func (d *DNSDumpster) query(domain string) error {
	// 设置请求头
	d.SetHeader("Referer", "https://dnsdumpster.com")
	d.SetHeader("User-Agent", d.GetRandomUserAgent())

	// 获取初始页面和 CSRF token
	resp, err := d.HTTPGet(d.baseURL, d.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to get initial page: %v", err)
	}

	// 提取 CSRF token
	csrfToken := d.extractCSRFToken(resp)
	if csrfToken == "" {
		return fmt.Errorf("failed to extract CSRF token")
	}

	// 构建 POST 数据
	data := url.Values{}
	data.Set("csrfmiddlewaretoken", csrfToken)
	data.Set("targetip", domain)
	data.Set("user", "free")

	// 发送 POST 请求
	resp, err = d.HTTPPost(d.baseURL, data, d.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to post query: %v", err)
	}

	// 读取响应
	body, err := d.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := d.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		d.AddSubdomain(subdomain)
	}

	return nil
}

// extractCSRFToken 提取 CSRF token
func (d *DNSDumpster) extractCSRFToken(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	// 从响应头中提取 CSRF token
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			return cookie.Value
		}
	}

	return ""
}
