package core

import (
	"github.com/oneforall-go/internal/config"
)

// Registry 模块注册器
type Registry struct {
	dispatcher *Dispatcher
}

// NewRegistry 创建模块注册器
func NewRegistry(cfg *config.Config) *Registry {
	return &Registry{
		dispatcher: NewDispatcher(cfg),
	}
}

// RegisterAllModules 注册所有模块
func (r *Registry) RegisterAllModules() {
	// 这里将在主程序中注册模块
	// 为了避免循环导入，具体的模块注册将在 main.go 中处理
	r.dispatcher.ListModules()
}

// GetDispatcher 获取调度器
func (r *Registry) GetDispatcher() *Dispatcher {
	return r.dispatcher
}
