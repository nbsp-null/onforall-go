package core

import (
	"github.com/oneforall-go/internal/config"
)

// Query 查询大类基础类（对应 Python 的 Query 基类）
type Query struct {
	*BaseModule
}

// NewQuery 创建查询基础类
func NewQuery(name string, cfg *config.Config) *Query {
	return &Query{
		BaseModule: NewBaseModule(name, ModuleTypeSearch, cfg),
	}
}
