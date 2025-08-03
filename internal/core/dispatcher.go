package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/validator"
	"github.com/oneforall-go/pkg/logger"
)

// ExecutionStep 执行步骤
type ExecutionStep struct {
	Name        string
	Modules     []Module
	Concurrency int
	Timeout     time.Duration
	Enabled     bool
}

// Dispatcher 统一调度器
type Dispatcher struct {
	config *config.Config

	// 各大类模块
	searchModules       []Module
	datasetModules      []Module
	certificateModules  []Module
	bruteModules        []Module
	dnsLookupModules    []Module
	resolveModules      []Module
	checkModules        []Module
	crawlModules        []Module
	intelligenceModules []Module
	enrichModules       []Module

	// 执行步骤
	executionSteps []ExecutionStep

	// 域名验证器
	validator *validator.DomainValidator

	// 线程安全
	mutex sync.RWMutex
}

// NewDispatcher 创建调度器
func NewDispatcher(cfg *config.Config) *Dispatcher {
	d := &Dispatcher{
		config:           cfg,
		searchModules:    make([]Module, 0),
		bruteModules:     make([]Module, 0),
		dnsLookupModules: make([]Module, 0),
		resolveModules:   make([]Module, 0),
		checkModules:     make([]Module, 0),
		crawlModules:     make([]Module, 0),
		enrichModules:    make([]Module, 0),
		executionSteps:   make([]ExecutionStep, 0),
		validator:        validator.NewDomainValidator(cfg),
	}

	// 初始化执行步骤
	d.initExecutionSteps()

	return d
}

// initExecutionSteps 初始化执行步骤
func (d *Dispatcher) initExecutionSteps() {
	d.executionSteps = []ExecutionStep{
		{
			Name:        "Fast Search",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.FastSearchConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.FastSearchTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableFastSearch,
		},
		{
			Name:        "Dataset",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.DatasetConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.DatasetTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableDataset,
		},
		{
			Name:        "Certificate",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.CertificateConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.CertificateTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableCertificate,
		},
		{
			Name:        "Crawl",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.CrawlConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.CrawlTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableCrawl,
		},
		{
			Name:        "DNS Lookup",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.DNSLookupConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.DNSLookupTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableDNSLookup,
		},
		{
			Name:        "Intelligence",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.IntelligenceConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.IntelligenceTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableIntelligence,
		},
		{
			Name:        "Brute Force",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.BruteForceConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.BruteForceTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableBruteForce,
		},
		{
			Name:        "File Check",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.FileCheckConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.FileCheckTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableFileCheck,
		},
		{
			Name:        "Enrich",
			Modules:     []Module{},
			Concurrency: d.config.MultiThreading.EnrichConcurrency,
			Timeout:     time.Duration(d.config.MultiThreading.EnrichTimeout) * time.Second,
			Enabled:     d.config.MultiThreading.EnableEnrich,
		},
		{
			Name:        "Validation",
			Modules:     []Module{},
			Concurrency: d.config.ValidationConcurrency,
			Timeout:     time.Duration(d.config.ValidationTimeout) * time.Second,
			Enabled:     d.config.EnableDomainValidation,
		},
	}
}

// RegisterModule 注册模块
func (d *Dispatcher) RegisterModule(module Module) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	logger.Debugf("Registering module: %s", module.Name())

	moduleType := d.getModuleType(module)
	logger.Debugf("Module %s classified as type: %s", module.Name(), moduleType)

	switch moduleType {
	case ModuleTypeSearch:
		d.searchModules = append(d.searchModules, module)
		logger.Debugf("Added to search modules")
	case ModuleTypeBrute:
		d.bruteModules = append(d.bruteModules, module)
		logger.Debugf("Added to brute modules")
	case ModuleTypeDNSLookup:
		d.dnsLookupModules = append(d.dnsLookupModules, module)
		logger.Debugf("Added to DNS lookup modules")
	case ModuleTypeCheck:
		d.checkModules = append(d.checkModules, module)
		logger.Debugf("Added to check modules")
	case ModuleTypeCrawl:
		d.crawlModules = append(d.crawlModules, module)
		logger.Debugf("Added to crawl modules")
	case ModuleTypeEnrich:
		// 根据模块名称进一步分类
		moduleName := module.Name()
		if isDatasetModule(moduleName) {
			d.datasetModules = append(d.datasetModules, module)
			logger.Debugf("Added to dataset modules")
		} else if isCertificateModule(moduleName) {
			d.certificateModules = append(d.certificateModules, module)
			logger.Debugf("Added to certificate modules")
		} else if isIntelligenceModule(moduleName) {
			d.intelligenceModules = append(d.intelligenceModules, module)
			logger.Debugf("Added to intelligence modules")
		} else {
			d.enrichModules = append(d.enrichModules, module)
			logger.Debugf("Added to enrich modules")
		}
	default:
		logger.Warnf("Unknown module type for %s: %s", module.Name(), moduleType)
	}

	logger.Infof("Successfully registered module: %s (Type: %s)", module.Name(), moduleType)
}

// RunAllModules 运行所有模块（分步执行）
func (d *Dispatcher) RunAllModules(domain string) (map[ModuleType][]string, []validator.ValidationResult, error) {
	results := make(map[ModuleType][]string)
	var allSubdomains []string
	var validationResults []validator.ValidationResult

	logger.Infof("=== Starting subdomain enumeration for domain: %s ===", domain)
	logger.Debugf("Total execution steps: %d", len(d.executionSteps))

	// 执行所有步骤（包括爆破模块）
	logger.Infof("=== Running all modules ===")
	for i, step := range d.executionSteps {
		// 跳过验证模块，它将在最后单独处理
		if step.Name == "Validation" {
			logger.Debugf("Skipping validation step, will be handled separately")
			continue
		}

		logger.Debugf("Processing step %d/%d: %s", i+1, len(d.executionSteps), step.Name)

		if !step.Enabled {
			logger.Debugf("Step %d (%s) is disabled, skipping", i+1, step.Name)
			continue
		}

		logger.Infof("Step %d/%d: %s (Concurrency: %d, Timeout: %v)",
			i+1, len(d.executionSteps), step.Name, step.Concurrency, step.Timeout)

		// 获取当前步骤的模块
		stepModules := d.getModulesForStep(step.Name, false) // 不排除爆破模块
		if len(stepModules) == 0 {
			logger.Debugf("No modules for step %s, skipping", step.Name)
			continue
		}

		logger.Debugf("Step %s has %d modules to execute", step.Name, len(stepModules))

		// 执行当前步骤并等待完全完成
		logger.Infof("Starting step %s execution...", step.Name)
		stepResults, err := d.runModulesWithConcurrency(stepModules, domain, step.Concurrency, step.Timeout, d.getModuleTypeForStep(step.Name) == ModuleTypeBrute)
		if err != nil {
			logger.Errorf("Step %s failed: %v", step.Name, err)
			continue
		}

		// 合并结果
		stepType := d.getModuleTypeForStep(step.Name)
		results[stepType] = stepResults
		allSubdomains = append(allSubdomains, stepResults...)

		logger.Infof("Step %s completed, found %d subdomains (Total: %d)",
			step.Name, len(stepResults), len(allSubdomains))

		// 确保当前步骤完全完成后再继续下一步
		logger.Infof("Step %s fully completed, proceeding to next step", step.Name)
	}

	logger.Infof("=== All collection modules completed ===")
	logger.Infof("Total subdomains collected: %d", len(allSubdomains))

	// 执行验证模块
	logger.Infof("=== Running validation module ===")
	if d.config.EnableDomainValidation && len(allSubdomains) > 0 {
		logger.Info("=== Starting domain validation and deduplication ===")

		// 验证域名
		validationResults = d.validator.ValidateDomains(allSubdomains, d.config.ValidationConcurrency)

		// 保留所有结果，包括验证不通过的域名
		var allValidatedResults []validator.ValidationResult
		for _, result := range validationResults {
			allValidatedResults = append(allValidatedResults, result)
		}

		// 获取验证统计信息
		stats := d.validator.GetValidationStats(validationResults)
		logger.Infof("Validation completed: %d total, %d alive (%.1f%%), DNS: %d (%.1f%%), Ping: %d (%.1f%%)",
			stats["total_domains"], stats["alive_domains"], stats["alive_percentage"],
			stats["dns_resolved"], stats["dns_percentage"],
			stats["ping_alive"], stats["ping_percentage"])

		// 更新所有步骤的结果，保留所有域名但标记存活状态
		for stepType := range results {
			// 为每个步骤的结果添加验证信息
			var validatedResults []string
			for _, subdomain := range results[stepType] {
				// 查找对应的验证结果
				for _, validationResult := range allValidatedResults {
					if validationResult.Subdomain == subdomain {
						// 如果验证通过，保留域名；如果验证不通过，也保留但标记为不存活
						validatedResults = append(validatedResults, subdomain)
						break
					}
				}
			}
			results[stepType] = validatedResults
		}

		// 更新总域名列表，保留所有域名
		allSubdomains = []string{}
		for _, result := range allValidatedResults {
			allSubdomains = append(allSubdomains, result.Subdomain)
		}
	}

	logger.Infof("Final result: %d unique domains", len(allSubdomains))
	return results, validationResults, nil
}

// RunLib 库调用接口，支持参数化调用并返回数据结构数组
func (d *Dispatcher) RunLib(domain string, options map[string]interface{}) ([]SubdomainResult, error) {
	logger.Infof("=== Starting library call for domain: %s ===", domain)

	// 解析选项参数
	enableValidation := true
	if val, ok := options["enable_validation"].(bool); ok {
		enableValidation = val
	}

	enableBruteForce := true
	if val, ok := options["enable_brute_force"].(bool); ok {
		enableBruteForce = val
	}

	concurrency := 10
	if val, ok := options["concurrency"].(int); ok {
		concurrency = val
	}

	timeout := 60 * time.Second
	if val, ok := options["timeout"].(time.Duration); ok {
		timeout = val
	}

	var allResults []SubdomainResult
	var allSubdomains []string

	// 执行所有收集模块
	logger.Infof("=== Running collection modules ===")
	for i, step := range d.executionSteps {
		// 跳过验证模块和爆破模块（如果禁用）
		if step.Name == "Validation" {
			logger.Debugf("Skipping validation step in library call")
			continue
		}

		if step.Name == "Brute Force" && !enableBruteForce {
			logger.Debugf("Skipping brute force step (disabled)")
			continue
		}

		logger.Debugf("Processing step %d/%d: %s", i+1, len(d.executionSteps), step.Name)

		if !step.Enabled {
			logger.Debugf("Step %d (%s) is disabled, skipping", i+1, step.Name)
			continue
		}

		logger.Infof("Step %d/%d: %s (Concurrency: %d, Timeout: %v)",
			i+1, len(d.executionSteps), step.Name, concurrency, timeout)

		// 获取当前步骤的模块
		stepModules := d.getModulesForStep(step.Name, false)
		if len(stepModules) == 0 {
			logger.Debugf("No modules for step %s, skipping", step.Name)
			continue
		}

		logger.Debugf("Step %s has %d modules to execute", step.Name, len(stepModules))

		// 执行当前步骤
		stepResults, err := d.runModulesWithConcurrency(stepModules, domain, concurrency, timeout, d.getModuleTypeForStep(step.Name) == ModuleTypeBrute)
		if err != nil {
			logger.Errorf("Step %s failed: %v", step.Name, err)
			continue
		}

		// 转换为SubdomainResult结构
		stepType := d.getModuleTypeForStep(step.Name)
		for _, subdomain := range stepResults {
			result := SubdomainResult{
				Subdomain: subdomain,
				Source:    string(stepType),
				Time:      time.Now().Format("2006-01-02 15:04:05"),
				Alive:     false, // 默认未检查存活状态
			}
			allResults = append(allResults, result)
		}
		allSubdomains = append(allSubdomains, stepResults...)

		logger.Infof("Step %s completed, found %d subdomains (Total: %d)",
			step.Name, len(stepResults), len(allSubdomains))
	}

	logger.Infof("=== Collection modules completed ===")
	logger.Infof("Total subdomains collected: %d", len(allSubdomains))

	// 执行验证模块（如果启用）
	if enableValidation && len(allSubdomains) > 0 {
		logger.Infof("=== Running validation module ===")
		logger.Info("=== Starting domain validation and deduplication ===")

		// 验证域名
		validationResults := d.validator.ValidateDomains(allSubdomains, concurrency)

		// 更新结果中的验证信息
		for i, result := range allResults {
			for _, validationResult := range validationResults {
				if validationResult.Subdomain == result.Subdomain {
					result.IP = validationResult.IP
					result.Alive = validationResult.Alive
					result.DNSResolved = validationResult.DNSResolved
					result.PingAlive = validationResult.PingAlive
					result.StatusCode = validationResult.StatusCode
					result.StatusText = validationResult.StatusText
					result.Provider = validationResult.Provider
					allResults[i] = result
					break
				}
			}
		}

		// 获取验证统计信息
		stats := d.validator.GetValidationStats(validationResults)
		logger.Infof("Validation completed: %d total, %d alive (%.1f%%), DNS: %d (%.1f%%), Ping: %d (%.1f%%)",
			stats["total_domains"], stats["alive_domains"], stats["alive_percentage"],
			stats["dns_resolved"], stats["dns_percentage"],
			stats["ping_alive"], stats["ping_percentage"])
	}

	logger.Infof("Library call completed. Returning %d results", len(allResults))
	return allResults, nil
}

// getModulesForStep 根据步骤名称获取对应的模块
func (d *Dispatcher) getModulesForStep(stepName string, excludeBrute bool) []Module {
	logger.Debugf("Getting modules for step: %s (excludeBrute: %t)", stepName, excludeBrute)

	var modules []Module
	switch stepName {
	case "Fast Search":
		modules = d.searchModules
	case "Dataset":
		modules = d.getDatasetModules()
	case "Certificate":
		modules = d.getCertificateModules()
	case "Crawl":
		modules = d.crawlModules
	case "DNS Lookup":
		modules = d.dnsLookupModules
	case "File Check":
		modules = d.checkModules
	case "Intelligence":
		modules = d.getIntelligenceModules()
	case "Brute Force":
		if excludeBrute {
			logger.Debugf("Excluding brute force modules as requested")
			modules = []Module{}
		} else {
			modules = d.bruteModules
			logger.Debugf("Brute force modules found: %d", len(modules))
			for i, module := range modules {
				logger.Debugf("  - Brute module %d: %s", i+1, module.Name())
			}
		}
	case "Enrich":
		modules = d.enrichModules
	case "Validation":
		// 验证步骤不返回模块，因为它是一个特殊的处理步骤
		modules = []Module{}
		logger.Debugf("Validation step - no modules to execute")
	default:
		logger.Warnf("Unknown step name: %s", stepName)
		return []Module{}
	}

	logger.Debugf("Found %d modules for step '%s'", len(modules), stepName)
	return modules
}

// getEnrichModules 获取enrich模块
func (d *Dispatcher) getEnrichModules() []Module {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.enrichModules
}

// getModuleTypeForStep 根据步骤名称获取模块类型
func (d *Dispatcher) getModuleTypeForStep(stepName string) ModuleType {
	switch stepName {
	case "Fast Search":
		return ModuleTypeSearch
	case "Dataset":
		return ModuleTypeEnrich
	case "Certificate":
		return ModuleTypeEnrich
	case "Crawl":
		return ModuleTypeCrawl
	case "DNS Lookup":
		return ModuleTypeDNSLookup
	case "File Check":
		return ModuleTypeCheck
	case "Intelligence":
		return ModuleTypeEnrich
	case "Brute Force":
		return ModuleTypeBrute
	case "Enrich":
		return ModuleTypeEnrich
	case "Validation":
		return ModuleTypeEnrich // 验证步骤归类为Enrich类型
	default:
		return ModuleTypeSearch
	}
}

// getDatasetModules 获取数据集模块
func (d *Dispatcher) getDatasetModules() []Module {
	// 从所有注册的模块中筛选数据集模块
	var datasetModules []Module

	// 检查所有模块，根据模块名称判断是否为数据集模块
	allModules := append(append(append(d.searchModules, d.bruteModules...), d.dnsLookupModules...), d.enrichModules...)

	for _, module := range allModules {
		name := module.Name()
		// 根据模块名称判断是否为数据集模块
		if isDatasetModule(name) {
			datasetModules = append(datasetModules, module)
		}
	}

	return datasetModules
}

// getCertificateModules 获取证书模块
func (d *Dispatcher) getCertificateModules() []Module {
	// 从所有注册的模块中筛选证书模块
	var certificateModules []Module

	allModules := append(append(append(d.searchModules, d.bruteModules...), d.dnsLookupModules...), d.enrichModules...)

	for _, module := range allModules {
		name := module.Name()
		// 根据模块名称判断是否为证书模块
		if isCertificateModule(name) {
			certificateModules = append(certificateModules, module)
		}
	}

	return certificateModules
}

// getIntelligenceModules 获取情报模块
func (d *Dispatcher) getIntelligenceModules() []Module {
	// 从所有注册的模块中筛选情报模块
	var intelligenceModules []Module

	allModules := append(append(append(d.searchModules, d.bruteModules...), d.dnsLookupModules...), d.enrichModules...)

	for _, module := range allModules {
		name := module.Name()
		// 根据模块名称判断是否为情报模块
		if isIntelligenceModule(name) {
			intelligenceModules = append(intelligenceModules, module)
		}
	}

	return intelligenceModules
}

// getModuleType 获取模块类型
func (d *Dispatcher) getModuleType(module Module) ModuleType {
	moduleName := module.Name()

	// 根据模块名称判断类型
	if isSearchModule(moduleName) {
		return ModuleTypeSearch
	} else if isDatasetModule(moduleName) {
		return ModuleTypeEnrich
	} else if isCertificateModule(moduleName) {
		return ModuleTypeEnrich
	} else if isCrawlModule(moduleName) {
		return ModuleTypeCrawl
	} else if isDNSLookupModule(moduleName) {
		return ModuleTypeDNSLookup
	} else if isCheckModule(moduleName) {
		return ModuleTypeCheck
	} else if isIntelligenceModule(moduleName) {
		return ModuleTypeEnrich
	} else if isBruteModule(moduleName) {
		return ModuleTypeBrute
	} else if isEnrichModule(moduleName) {
		return ModuleTypeEnrich
	}

	return ModuleTypeSearch
}

// isSearchModule 判断是否为搜索模块
func isSearchModule(name string) bool {
	searchModules := []string{"GoogleSearch", "BingSearch", "BaiduSearch", "YahooSearch", "SogouSearch", "YandexSearch", "SoSearch", "AskSearch", "GithubAPISearch", "GiteeSearch", "FoFaAPISearch", "ShodanAPISearch", "ZoomEyeAPISearch", "QuakeAPISearch", "HunterAPISearch", "BingAPISearch", "GoogleAPISearch", "DNSDumpsterQuery", "SecurityTrailsAPIQuery", "AnubisQuery", "BeVigilOsintApi", "BinaryEdgeAPIQuery", "ChinazQuery", "ChinazAPIQuery", "CirclAPIQuery", "CloudFlareAPIQuery", "DNSdbAPIQuery", "DnsgrepQuery", "FullHuntAPIQuery", "HackerTargetQuery", "IP138Query", "IPv4InfoAPIQuery", "NetCraftQuery", "PassiveDnsQuery", "QianXunQuery", "RapidDNSQuery", "RiddlerQuery", "RobtexQuery", "SiteDossierQuery", "SpyseAPIQuery", "Sublist3rQuery", "UrlscanQuery", "CensysAPIQuery", "CertSpotterQuery", "CrtshQuery", "GoogleQuery", "MySSLQuery", "RacentQuery", "AXFRCheck", "CrossDomainCheck", "CertInfo", "CSPCheck", "NSECCheck", "RobotsCheck", "SitemapCheck", "ArchiveCrawl", "CommonCrawl", "NSQuery", "QueryMX", "QuerySOA", "QuerySPF", "QueryTXT", "AlienVaultQuery", "RiskIQAPIQuery", "ThreatBookAPIQuery", "ThreatMinerQuery", "VirusTotalQuery", "VirusTotalAPIQuery"}
	return containsInSlice(searchModules, name)
}

// isDatasetModule 判断是否为数据集模块
func isDatasetModule(name string) bool {
	datasetModules := []string{"AnubisQuery", "BeVigilOsintApi", "BinaryEdgeAPIQuery", "CebaiduQuery", "ChinazQuery", "CirclAPIQuery", "CloudFlareAPIQuery", "DNSdbAPIQuery", "DNSDumpsterQuery", "DnsgrepQuery", "FullHuntAPIQuery", "HackerTargetQuery", "IP138Query", "IPv4InfoAPIQuery", "NetCraftQuery", "PassiveDnsQuery", "QianXunQuery", "RapidDNSQuery", "RiddlerQuery", "RobtexQuery", "SecurityTrailsAPIQuery", "SiteDossierQuery", "SpyseAPIQuery", "Sublist3rQuery", "UrlscanQuery"}
	return containsInSlice(datasetModules, name)
}

// isCertificateModule 判断是否为证书模块
func isCertificateModule(name string) bool {
	certificateModules := []string{"CensysAPIQuery", "CertSpotterQuery", "CrtshQuery", "GoogleQuery", "MySSLQuery", "RacentQuery"}
	return containsInSlice(certificateModules, name)
}

// isCrawlModule 判断是否为爬虫模块
func isCrawlModule(name string) bool {
	crawlModules := []string{"ArchiveCrawl", "CommonCrawl"}
	return containsInSlice(crawlModules, name)
}

// isDNSLookupModule 判断是否为DNS查询模块
func isDNSLookupModule(name string) bool {
	dnsLookupModules := []string{"NSQuery", "QueryMX", "QuerySOA", "QuerySPF", "QueryTXT"}
	return containsInSlice(dnsLookupModules, name)
}

// isCheckModule 判断是否为检查模块
func isCheckModule(name string) bool {
	checkModules := []string{"AXFRCheck", "CrossDomainCheck", "CertInfo", "CSPCheck", "NSECCheck", "RobotsCheck", "SitemapCheck"}
	return containsInSlice(checkModules, name)
}

// isIntelligenceModule 判断是否为情报模块
func isIntelligenceModule(name string) bool {
	intelligenceModules := []string{"AlienVaultQuery", "RiskIQAPIQuery", "ThreatBookAPIQuery", "ThreatMinerQuery", "VirusTotalQuery", "VirusTotalAPIQuery"}
	return containsInSlice(intelligenceModules, name)
}

// isBruteModule 判断是否为爆破模块
func isBruteModule(name string) bool {
	bruteModules := []string{"Brute", "Alt"}
	return containsInSlice(bruteModules, name)
}

// isEnrichModule 判断是否为丰富模块
func isEnrichModule(name string) bool {
	enrichModules := []string{"enrich"}
	return containsInSlice(enrichModules, name)
}

// containsInSlice 检查字符串是否在切片中
func containsInSlice(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

// containsSubstring 检查字符串中间是否包含子字符串
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// runModulesWithConcurrency 使用指定并发数运行模块
func (d *Dispatcher) runModulesWithConcurrency(modules []Module, domain string, concurrency int, timeout time.Duration, isBruteStep bool) ([]string, error) {
	if len(modules) == 0 {
		return []string{}, nil
	}

	var allResults []string
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errors []error

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, concurrency)

	// 创建超时控制（爆破模块不设置超时）
	var done chan bool
	if !isBruteStep {
		done = make(chan bool, 1)
		go func() {
			time.Sleep(timeout)
			done <- true
		}()
	}

	for _, module := range modules {
		if !module.IsEnabled() {
			logger.Debugf("Module %s is disabled, skipping", module.Name())
			continue
		}

		wg.Add(1)
		go func(module Module) {
			defer wg.Done()

			// 添加异常处理
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic in module %s: %v", module.Name(), r)
					mutex.Lock()
					errors = append(errors, fmt.Errorf("panic in %s: %v", module.Name(), r))
					mutex.Unlock()
				}
			}()

			// 获取信号量
			if !isBruteStep {
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				case <-done:
					logger.Warnf("Module %s skipped due to timeout", module.Name())
					return
				}
			} else {
				// 爆破模块不设置超时，直接获取信号量
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
			}

			logger.Debugf("Starting module: %s", module.Name())
			startTime := time.Now()

			results, err := module.Run(domain)
			if err != nil {
				logger.Errorf("Module %s failed: %v", module.Name(), err)
				mutex.Lock()
				errors = append(errors, fmt.Errorf("%s: %v", module.Name(), err))
				mutex.Unlock()
				return
			}

			elapsed := time.Since(startTime)
			mutex.Lock()
			allResults = append(allResults, results...)
			mutex.Unlock()

			logger.Infof("Module %s completed in %v, found %d subdomains",
				module.Name(), elapsed, len(results))
		}(module)
	}

	// 等待所有模块完成或超时
	completed := make(chan bool, 1)
	go func() {
		wg.Wait()
		completed <- true
	}()

	if !isBruteStep {
		select {
		case <-completed:
			logger.Debugf("All modules completed successfully")
		case <-done:
			logger.Warnf("Some modules timed out after %v", timeout)
		}
	} else {
		// 爆破模块等待完成，不设置超时
		<-completed
		logger.Debugf("All brute force modules completed successfully")
	}

	if len(errors) > 0 {
		logger.Warnf("Some modules failed: %v", errors)
	}

	return allResults, nil
}

// runModules 运行指定类型的模块（兼容旧版本）
func (d *Dispatcher) runModules(modules []Module, domain string) ([]string, error) {
	return d.runModulesWithConcurrency(modules, domain, 10, 60*time.Second, false)
}

// GetModuleStats 获取模块统计信息
func (d *Dispatcher) GetModuleStats() map[ModuleType]int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	stats := make(map[ModuleType]int)
	stats[ModuleTypeSearch] = len(d.searchModules)
	stats[ModuleTypeBrute] = len(d.bruteModules)
	stats[ModuleTypeDNSLookup] = len(d.dnsLookupModules)
	stats[ModuleTypeResolve] = len(d.resolveModules)
	stats[ModuleTypeCheck] = len(d.checkModules)
	stats[ModuleTypeCrawl] = len(d.crawlModules)
	stats[ModuleTypeEnrich] = len(d.enrichModules)

	return stats
}

// ListModules 列出所有模块
func (d *Dispatcher) ListModules() {
	stats := d.GetModuleStats()
	logger.Info("Available modules:")
	for moduleType, count := range stats {
		logger.Infof("  %s: %d modules", moduleType, count)
	}

	logger.Info("Execution steps:")
	for i, step := range d.executionSteps {
		enabled := "Enabled"
		if !step.Enabled {
			enabled = "Disabled"
		}
		logger.Infof("  Step %d: %s (%s, Concurrency: %d, Timeout: %v)",
			i+1, step.Name, enabled, step.Concurrency, step.Timeout)
	}
}

// GetModules 获取指定类型的模块
func (d *Dispatcher) GetModules(moduleType ModuleType) []Module {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	switch moduleType {
	case ModuleTypeSearch:
		return d.searchModules
	case ModuleTypeBrute:
		return d.bruteModules
	case ModuleTypeDNSLookup:
		return d.dnsLookupModules
	case ModuleTypeResolve:
		return d.resolveModules
	case ModuleTypeCheck:
		return d.checkModules
	case ModuleTypeCrawl:
		return d.crawlModules
	case ModuleTypeEnrich:
		return d.enrichModules
	default:
		return []Module{}
	}
}

// SetStepEnabled 设置步骤启用状态
func (d *Dispatcher) SetStepEnabled(stepIndex int, enabled bool) {
	if stepIndex >= 0 && stepIndex < len(d.executionSteps) {
		d.executionSteps[stepIndex].Enabled = enabled
	}
}

// SetStepConcurrency 设置步骤并发数
func (d *Dispatcher) SetStepConcurrency(stepIndex int, concurrency int) {
	if stepIndex >= 0 && stepIndex < len(d.executionSteps) && concurrency > 0 {
		d.executionSteps[stepIndex].Concurrency = concurrency
	}
}

// SetStepTimeout 设置步骤超时时间
func (d *Dispatcher) SetStepTimeout(stepIndex int, timeout time.Duration) {
	if stepIndex >= 0 && stepIndex < len(d.executionSteps) {
		d.executionSteps[stepIndex].Timeout = timeout
	}
}
