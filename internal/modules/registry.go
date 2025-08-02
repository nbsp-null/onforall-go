package modules

import (
	"github.com/oneforall-go/internal/config"
)

// Registry 模块注册器
type Registry struct {
	manager *Manager
}

// NewRegistry 创建模块注册器
func NewRegistry(cfg *config.Config) *Registry {
	manager := NewManager(cfg)
	return &Registry{
		manager: manager,
	}
}

// RegisterAllModules 注册所有模块
func (r *Registry) RegisterAllModules() {
	// 这里可以注册所有模块
	// 为了避免循环导入，我们将在主程序中注册模块
	r.manager.ListModules()
}

// GetManager 获取模块管理器
func (r *Registry) GetManager() *Manager {
	return r.manager
}

// ListRegisteredModules 列出已注册的模块
func (r *Registry) ListRegisteredModules() {
	r.manager.ListModules()
}
