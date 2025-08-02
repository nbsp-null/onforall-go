package datasets

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/modules"
)

// DNSDumpster DNSDumpster 数据集模块
type DNSDumpster struct {
	*modules.BaseModule
	baseURL string
}

// NewDNSDumpster 创建 DNSDumpster 模块
func NewDNSDumpster(cfg *config.Config) *DNSDumpster {
	return &DNSDumpster{
		BaseModule: modules.NewBaseModule("DNSDumpster", modules.ModuleTypeDataset, cfg),
		baseURL:    "https://dnsdumpster.com/",
	}
}

// Run 执行查询
func (d *DNSDumpster) Run(domain string) ([]string, error) {
	d.LogInfo("Starting DNSDumpster query for domain: %s", domain)

	// 设置请求头
	headers := map[string]string{
		"Referer": "https://dnsdumpster.com",
	}

	// 获取初始页面
	resp, err := d.HTTPGet(d.baseURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial page: %v", err)
	}

	// 提取 CSRF token
	csrfToken := d.extractCSRFToken(resp)
	if csrfToken == "" {
		return nil, fmt.Errorf("failed to extract CSRF token")
	}

	// 构建查询数据
	data := url.Values{}
	data.Set("csrfmiddlewaretoken", csrfToken)
	data.Set("targetip", domain)
	data.Set("user", "free")

	// 执行查询
	queryResp, err := d.HTTPPost(d.baseURL, data, headers)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %v", err)
	}

	// 读取响应
	body, err := d.ReadResponseBody(queryResp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := d.ExtractSubdomains(body, domain)

	d.LogInfo("DNSDumpster query completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}

// extractCSRFToken 提取 CSRF token
func (d *DNSDumpster) extractCSRFToken(resp *http.Response) string {
	// 从响应头中提取 CSRF token
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			return cookie.Value
		}
	}
	return ""
}
