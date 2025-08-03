package api

import (
	"fmt"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
	"github.com/oneforall-go/pkg/logger"
)

// SubdomainResult 子域名结果结构
type SubdomainResult struct {
	Subdomain   string   `json:"subdomain"`
	Source      string   `json:"source"`
	Time        string   `json:"time"`
	Alive       bool     `json:"alive"`
	IP          []string `json:"ip,omitempty"`
	DNSResolved bool     `json:"dns_resolved"`
	PingAlive   bool     `json:"ping_alive"`
	StatusCode  int      `json:"status_code"`
	StatusText  string   `json:"status_text"`
	Provider    string   `json:"provider,omitempty"`
}

// Options 配置选项
type Options struct {
	// 基本配置
	Target string `json:"target"` // 目标域名（必需）

	// 功能开关
	EnableValidation bool `json:"enable_validation"`  // 是否启用域名验证
	EnableBruteForce bool `json:"enable_brute_force"` // 是否启用爆破攻击

	// 性能配置
	Concurrency int           `json:"concurrency"` // 并发数
	Timeout     time.Duration `json:"timeout"`     // 超时时间

	// 模块开关
	EnableSearchModules       bool `json:"enable_search_modules"`       // 搜索模块
	EnableDatasetModules      bool `json:"enable_dataset_modules"`      // 数据集模块
	EnableCertificateModules  bool `json:"enable_certificate_modules"`  // 证书模块
	EnableCrawlModules        bool `json:"enable_crawl_modules"`        // 爬虫模块
	EnableCheckModules        bool `json:"enable_check_modules"`        // 检查模块
	EnableIntelligenceModules bool `json:"enable_intelligence_modules"` // 智能模块
	EnableEnrichModules       bool `json:"enable_enrich_modules"`       // 丰富模块

	// 爆破模块配置
	BruteDictionaryURL string `json:"brute_dictionary_url"` // 爆破字典URL
	BruteDNSServerURL  string `json:"brute_dns_server_url"` // 爆破DNS服务器URL

	// 日志配置
	Debug   bool `json:"debug"`   // 调试模式
	Verbose bool `json:"verbose"` // 详细日志
}

// Result 执行结果
type Result struct {
	Domain          string            `json:"domain"`           // 目标域名
	TotalSubdomains int               `json:"total_subdomains"` // 总子域名数
	AliveSubdomains int               `json:"alive_subdomains"` // 存活子域名数
	AlivePercentage float64           `json:"alive_percentage"` // 存活百分比
	Results         []SubdomainResult `json:"results"`          // 详细结果
	ExecutionTime   time.Duration     `json:"execution_time"`   // 执行时间
	Error           string            `json:"error,omitempty"`  // 错误信息
}

// OneForAllAPI OneForAll API接口
type OneForAllAPI struct {
	config     *config.Config
	dispatcher *core.Dispatcher
}

// NewOneForAllAPI 创建新的API实例
func NewOneForAllAPI() *OneForAllAPI {
	cfg := config.GetConfig()
	return &OneForAllAPI{
		config:     cfg,
		dispatcher: core.NewDispatcher(cfg),
	}
}

// RunSubdomainEnumeration 运行子域名枚举（仅返回数据结构，不保存到本地）
func (api *OneForAllAPI) RunSubdomainEnumeration(options Options) (*Result, error) {
	startTime := time.Now()

	// 验证必需参数
	if options.Target == "" {
		return &Result{
			Domain:        options.Target,
			ExecutionTime: time.Since(startTime),
			Error:         "target domain is required",
		}, fmt.Errorf("target domain is required")
	}

	// 设置默认值
	if options.Concurrency <= 0 {
		options.Concurrency = 10
	}
	if options.Timeout <= 0 {
		options.Timeout = 60 * time.Second
	}

	// 配置日志
	if options.Debug {
		logger.Init("debug", "")
	} else if options.Verbose {
		logger.Init("info", "")
	} else {
		logger.Init("warn", "")
	}

	// 注册模块
	api.registerModules(options)

	// 准备库调用选项
	libOptions := map[string]interface{}{
		"enable_validation":    options.EnableValidation,
		"enable_brute_force":   options.EnableBruteForce,
		"concurrency":          options.Concurrency,
		"timeout":              options.Timeout,
		"brute_dictionary_url": options.BruteDictionaryURL,
		"brute_dns_server_url": options.BruteDNSServerURL,
	}

	// 执行子域名枚举
	logger.Infof("Starting subdomain enumeration for domain: %s", options.Target)
	results, err := api.dispatcher.RunLib(options.Target, libOptions)
	if err != nil {
		return &Result{
			Domain:        options.Target,
			ExecutionTime: time.Since(startTime),
			Error:         err.Error(),
		}, err
	}

	// 转换结果格式
	var apiResults []SubdomainResult
	if results != nil {
		apiResults = make([]SubdomainResult, len(results))
		for i, result := range results {
			apiResults[i] = SubdomainResult{
				Subdomain:   result.Subdomain,
				Source:      result.Source,
				Time:        result.Time,
				Alive:       result.Alive,
				IP:          result.IP,
				DNSResolved: result.DNSResolved,
				PingAlive:   result.PingAlive,
				StatusCode:  result.StatusCode,
				StatusText:  result.StatusText,
				Provider:    result.Provider,
			}
		}
	}

	// 计算统计信息
	aliveCount := 0
	for _, result := range apiResults {
		if result.Alive {
			aliveCount++
		}
	}

	alivePercentage := 0.0
	if len(apiResults) > 0 {
		alivePercentage = float64(aliveCount) / float64(len(apiResults)) * 100
	}

	executionTime := time.Since(startTime)

	logger.Infof("Enumeration completed for %s: %d total, %d alive (%.1f%%) in %v",
		options.Target, len(apiResults), aliveCount, alivePercentage, executionTime)

	return &Result{
		Domain:          options.Target,
		TotalSubdomains: len(apiResults),
		AliveSubdomains: aliveCount,
		AlivePercentage: alivePercentage,
		Results:         apiResults,
		ExecutionTime:   executionTime,
	}, nil
}

// registerModules 注册模块
func (api *OneForAllAPI) registerModules(options Options) {
	// 注册搜索模块
	if options.EnableSearchModules {
		api.registerSearchModules()
	}

	// 注册数据集模块
	if options.EnableDatasetModules {
		api.registerDatasetModules()
	}

	// 注册证书模块
	if options.EnableCertificateModules {
		api.registerCertificateModules()
	}

	// 注册爬虫模块
	if options.EnableCrawlModules {
		api.registerCrawlModules()
	}

	// 注册检查模块
	if options.EnableCheckModules {
		api.registerCheckModules()
	}

	// 注册智能模块
	if options.EnableIntelligenceModules {
		api.registerIntelligenceModules()
	}

	// 注册爆破模块（如果启用）
	if options.EnableBruteForce {
		api.registerBruteModules()
	}

	// 注册丰富模块
	if options.EnableEnrichModules {
		api.registerEnrichModules()
	}
}

// registerSearchModules 注册搜索模块
func (api *OneForAllAPI) registerSearchModules() {
	logger.Debug("Registering search modules...")
	// 这里可以添加具体的搜索模块注册逻辑
}

// registerDatasetModules 注册数据集模块
func (api *OneForAllAPI) registerDatasetModules() {
	logger.Debug("Registering dataset modules...")
	// 这里可以添加具体的数据集模块注册逻辑
}

// registerCertificateModules 注册证书模块
func (api *OneForAllAPI) registerCertificateModules() {
	logger.Debug("Registering certificate modules...")
	// 这里可以添加具体的证书模块注册逻辑
}

// registerCrawlModules 注册爬虫模块
func (api *OneForAllAPI) registerCrawlModules() {
	logger.Debug("Registering crawl modules...")
	// 这里可以添加具体的爬虫模块注册逻辑
}

// registerCheckModules 注册检查模块
func (api *OneForAllAPI) registerCheckModules() {
	logger.Debug("Registering check modules...")
	// 这里可以添加具体的检查模块注册逻辑
}

// registerIntelligenceModules 注册智能模块
func (api *OneForAllAPI) registerIntelligenceModules() {
	logger.Debug("Registering intelligence modules...")
	// 这里可以添加具体的智能模块注册逻辑
}

// registerBruteModules 注册爆破模块
func (api *OneForAllAPI) registerBruteModules() {
	logger.Debug("Registering brute force modules...")
	// 这里可以添加具体的爆破模块注册逻辑
}

// registerEnrichModules 注册丰富模块
func (api *OneForAllAPI) registerEnrichModules() {
	logger.Debug("Registering enrich modules...")
	// 这里可以添加具体的丰富模块注册逻辑
}

// GetDefaultOptions 获取默认配置选项
func GetDefaultOptions() Options {
	return Options{
		EnableValidation:          true,
		EnableBruteForce:          false,
		Concurrency:               10,
		Timeout:                   60 * time.Second,
		EnableSearchModules:       true,
		EnableDatasetModules:      true,
		EnableCertificateModules:  true,
		EnableCrawlModules:        true,
		EnableCheckModules:        true,
		EnableIntelligenceModules: true,
		EnableEnrichModules:       true,
		BruteDictionaryURL:        "", // 默认使用本地字典
		BruteDNSServerURL:         "", // 默认使用本地DNS服务器
		Debug:                     false,
		Verbose:                   false,
	}
}
