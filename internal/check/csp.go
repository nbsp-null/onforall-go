package check

import (
	"fmt"
	"net/http"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// CSP CSP 检查模块
type CSP struct {
	*core.Check
	domain    string
	cspHeader http.Header
}

// NewCSP 创建 CSP 检查模块
func NewCSP(cfg *config.Config) *CSP {
	return &CSP{
		Check: core.NewCheck("CSPCheck", cfg),
	}
}

// Run 执行检查
func (c *CSP) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 执行检查
	if err := c.check(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// check 执行检查
func (c *CSP) check(domain string) error {
	// 获取 CSP 头
	cspHeader := c.grabHeader(domain)
	if cspHeader == nil {
		return fmt.Errorf("failed to get CSP header for domain: %s", domain)
	}

	// 检查 Content-Security-Policy 头
	csp := cspHeader.Get("Content-Security-Policy")
	if csp == "" {
		return fmt.Errorf("no Content-Security-Policy header found for domain: %s", domain)
	}

	// 提取子域名
	subdomains := c.ExtractSubdomains(csp, domain)
	for _, subdomain := range subdomains {
		c.AddSubdomain(subdomain)
	}

	return nil
}

// grabHeader 获取响应头
func (c *CSP) grabHeader(domain string) http.Header {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 尝试不同的 URL
	urls := []string{
		fmt.Sprintf("http://%s", domain),
		fmt.Sprintf("https://%s", domain),
		fmt.Sprintf("http://www.%s", domain),
		fmt.Sprintf("https://www.%s", domain),
	}

	for _, url := range urls {
		// 发送 GET 请求
		resp, err := c.HTTPGet(url, c.GetHeader())
		if err != nil {
			continue
		}

		// 检查响应状态
		if resp.StatusCode == 200 {
			return resp.Header
		}

		// 关闭响应体
		resp.Body.Close()
	}

	return nil
}
