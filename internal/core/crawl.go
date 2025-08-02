package core

import (
	"github.com/oneforall-go/internal/config"
)

// Crawl 爬虫大类基础类（对应 Python 的 Crawl 基类）
type Crawl struct {
	*BaseModule
}

// NewCrawl 创建爬虫基础类
func NewCrawl(name string, cfg *config.Config) *Crawl {
	return &Crawl{
		BaseModule: NewBaseModule(name, ModuleTypeCrawl, cfg),
	}
}
