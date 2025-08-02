package modules

import (
	"fmt"
	"sync"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// ModuleType 模块类型
type ModuleType string

const (
	ModuleTypeSearch       ModuleType = "search"
	ModuleTypeDataset      ModuleType = "dataset"
	ModuleTypeCertificate  ModuleType = "certificate"
	ModuleTypeIntelligence ModuleType = "intelligence"
	ModuleTypeCheck        ModuleType = "check"
	ModuleTypeCrawl        ModuleType = "crawl"
	ModuleTypeDNSQuery     ModuleType = "dnsquery"
)

// Module 模块接口
type Module interface {
	Name() string
	Type() ModuleType
	Run(domain string) ([]string, error)
	IsEnabled() bool
}

// Manager 模块管理器
type Manager struct {
	modules map[ModuleType][]Module
	config  *config.Config
	mutex   sync.RWMutex
}

// NewManager 创建新的模块管理器
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		modules: make(map[ModuleType][]Module),
		config:  cfg,
	}
}

// RegisterModule 注册模块
func (m *Manager) RegisterModule(module Module) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	moduleType := module.Type()
	if m.modules[moduleType] == nil {
		m.modules[moduleType] = make([]Module, 0)
	}
	m.modules[moduleType] = append(m.modules[moduleType], module)
	logger.Debugf("Registered module: %s (%s)", module.Name(), moduleType)
}

// GetModules 获取指定类型的所有模块
func (m *Manager) GetModules(moduleType ModuleType) []Module {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.modules[moduleType]
}

// GetAllModules 获取所有模块
func (m *Manager) GetAllModules() map[ModuleType][]Module {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[ModuleType][]Module)
	for moduleType, modules := range m.modules {
		result[moduleType] = make([]Module, len(modules))
		copy(result[moduleType], modules)
	}
	return result
}

// RunModules 运行指定类型的所有模块
func (m *Manager) RunModules(moduleType ModuleType, domain string) ([]string, error) {
	modules := m.GetModules(moduleType)
	if len(modules) == 0 {
		return []string{}, nil
	}

	logger.Infof("Running %d %s modules for domain: %s", len(modules), moduleType, domain)

	var allResults []string
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errors []error

	for _, module := range modules {
		if !module.IsEnabled() {
			logger.Debugf("Module %s is disabled, skipping", module.Name())
			continue
		}

		wg.Add(1)
		go func(module Module) {
			defer wg.Done()

			logger.Debugf("Starting module: %s", module.Name())
			results, err := module.Run(domain)
			if err != nil {
				logger.Errorf("Module %s failed: %v", module.Name(), err)
				mutex.Lock()
				errors = append(errors, fmt.Errorf("%s: %v", module.Name(), err))
				mutex.Unlock()
				return
			}

			mutex.Lock()
			allResults = append(allResults, results...)
			mutex.Unlock()

			logger.Infof("Module %s completed, found %d subdomains", module.Name(), len(results))
		}(module)
	}

	wg.Wait()

	if len(errors) > 0 {
		logger.Warnf("Some modules failed: %v", errors)
	}

	return allResults, nil
}

// RunAllModules 运行所有模块
func (m *Manager) RunAllModules(domain string) (map[ModuleType][]string, error) {
	allModules := m.GetAllModules()
	results := make(map[ModuleType][]string)

	for moduleType := range allModules {
		moduleResults, err := m.RunModules(moduleType, domain)
		if err != nil {
			logger.Errorf("Failed to run %s modules: %v", moduleType, err)
			continue
		}
		results[moduleType] = moduleResults
	}

	return results, nil
}

// GetModuleStats 获取模块统计信息
func (m *Manager) GetModuleStats() map[ModuleType]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[ModuleType]int)
	for moduleType, modules := range m.modules {
		enabledCount := 0
		for _, module := range modules {
			if module.IsEnabled() {
				enabledCount++
			}
		}
		stats[moduleType] = enabledCount
	}
	return stats
}

// ListModules 列出所有模块
func (m *Manager) ListModules() {
	stats := m.GetModuleStats()
	logger.Info("Available modules:")
	for moduleType, count := range stats {
		logger.Infof("  %s: %d enabled modules", moduleType, count)
	}
}
