package check

import (
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// CDX CDX 检查模块
type CDX struct {
	*core.Check
}

// NewCDX 创建 CDX 检查模块
func NewCDX(cfg *config.Config) *CDX {
	return &CDX{
		Check: core.NewCheck("CrossDomainCheck", cfg),
	}
}

// Run 执行检查
func (c *CDX) Run(domain string) ([]string, error) {
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
func (c *CDX) check(domain string) error {
	// 检查 crossdomain.xml 文件
	filenames := []string{"crossdomain.xml"}
	c.ToCheck(filenames)
	return nil
}
