package check

import (
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Sitemap Sitemap 检查模块
type Sitemap struct {
	*core.Check
}

// NewSitemap 创建 Sitemap 检查模块
func NewSitemap(cfg *config.Config) *Sitemap {
	return &Sitemap{
		Check: core.NewCheck("SitemapCheck", cfg),
	}
}

// Run 执行检查
func (s *Sitemap) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 执行检查
	if err := s.check(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// check 执行检查
func (s *Sitemap) check(domain string) error {
	// 设置要检查的文件名
	filenames := []string{
		"sitemap.xml",
		"sitemap.txt",
		"sitemap.html",
		"sitemapindex.xml",
	}

	// 调用基类的 ToCheck 方法
	s.ToCheck(filenames)

	return nil
}
