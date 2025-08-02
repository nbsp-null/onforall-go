package core

import (
	"fmt"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// Check 检查大类基础类（对应 Python 的 Check 基类）
type Check struct {
	*BaseModule
	requestStatus int
}

// NewCheck 创建检查基础类
func NewCheck(name string, cfg *config.Config) *Check {
	return &Check{
		BaseModule:    NewBaseModule(name, ModuleTypeCheck, cfg),
		requestStatus: 1,
	}
}

// ToCheck 检查文件
func (c *Check) ToCheck(filenames []string) {
	urls := make(map[string]bool)
	urlsWWW := make(map[string]bool)

	for _, filename := range filenames {
		urls[fmt.Sprintf("http://%s/%s", c.domain, filename)] = true
		urls[fmt.Sprintf("https://%s/%s", c.domain, filename)] = true
		urlsWWW[fmt.Sprintf("http://www.%s/%s", c.domain, filename)] = true
		urlsWWW[fmt.Sprintf("https://www.%s/%s", c.domain, filename)] = true
	}

	c.checkLoop(urls)
	c.checkLoop(urlsWWW)
}

// checkLoop 检查循环
func (c *Check) checkLoop(urls map[string]bool) {
	for urlStr := range urls {
		c.SetHeader("User-Agent", c.GetRandomUserAgent())

		resp, err := c.HTTPGet(urlStr, c.GetHeader())
		if err != nil {
			logger.Debugf("Connection to %s failed: %v", urlStr, err)
			continue
		}

		// 读取响应
		body, err := c.ReadResponseBody(resp)
		if err != nil {
			continue
		}

		// 提取子域名
		subdomains := c.ExtractSubdomains(body, c.domain)
		for _, subdomain := range subdomains {
			c.AddSubdomain(subdomain)
		}

		// 如果找到子域名，停止检查
		if len(subdomains) > 0 {
			break
		}
	}
}
