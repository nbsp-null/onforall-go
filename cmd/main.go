package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/joho/godotenv"
	"github.com/oneforall-go/internal/alt"
	brutepkg "github.com/oneforall-go/internal/brute"
	"github.com/oneforall-go/internal/certificates"
	"github.com/oneforall-go/internal/check"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
	"github.com/oneforall-go/internal/crawl"
	"github.com/oneforall-go/internal/datasets"
	"github.com/oneforall-go/internal/dnsquery"
	"github.com/oneforall-go/internal/enrich"
	"github.com/oneforall-go/internal/intelligence"
	"github.com/oneforall-go/internal/search"
	"github.com/oneforall-go/internal/validator"
	"github.com/oneforall-go/pkg/logger"
	"github.com/oneforall-go/pkg/utils"
)

var (
	// 命令行参数
	target    string
	targets   string
	brute     bool
	dns       bool
	req       bool
	port      string
	alive     bool
	outputFmt string
	path      string
	takeover  bool
	show      bool
	debug     bool
	verbose   bool
	thread    int
	timeout   int

	// 新架构参数
	searchModules bool
	dnsLookup     bool
	resolve       bool
	checkModules  bool
	crawlModules  bool
	enrichModules bool

	// 库调用参数
	enableValidation bool
	enableBruteForce bool
	libConcurrency   int
	libTimeout       int
)

// OneForAll OneForAll 主程序
type OneForAll struct {
	config     *config.Config
	dispatcher *core.Dispatcher
	output     *core.OutputManager
	domains    []string
}

// NewOneForAll 创建 OneForAll 实例
func NewOneForAll() *OneForAll {
	cfg := config.GetConfig()

	return &OneForAll{
		config:     cfg,
		dispatcher: core.NewDispatcher(cfg),
		output:     core.NewOutputManager(cfg),
		domains:    make([]string, 0),
	}
}

// run 运行主程序
func (o *OneForAll) run() error {
	logger.Info("Starting OneForAll...")

	// 配置参数
	o.configParam()

	// 加载域名
	if err := o.loadDomains(); err != nil {
		return err
	}

	// 注册模块
	o.registerModules()

	// 列出模块
	o.dispatcher.ListModules()

	// 处理每个域名
	for _, domain := range o.domains {
		logger.Infof("Processing domain: %s", domain)

		// 运行所有模块
		results, validationResults, err := o.dispatcher.RunAllModules(domain)
		if err != nil {
			logger.Errorf("Failed to run modules for %s: %v", domain, err)
			continue
		}

		// 处理结果
		o.processResults(domain, results, validationResults)
	}

	// 导出结果
	if err := o.output.Export(); err != nil {
		return fmt.Errorf("failed to export results: %v", err)
	}

	// 显示统计信息
	o.showStats()

	return nil
}

// runLib 运行库调用
func (o *OneForAll) runLib() error {
	logger.Info("Starting OneForAll Library Call...")

	// 配置参数
	o.configParam()

	// 加载域名
	if err := o.loadDomains(); err != nil {
		return err
	}

	// 注册模块
	o.registerModules()

	// 处理每个域名
	for _, domain := range o.domains {
		logger.Infof("Processing domain: %s", domain)

		// 准备库调用选项
		options := map[string]interface{}{
			"enable_validation":  enableValidation,
			"enable_brute_force": enableBruteForce,
			"concurrency":        libConcurrency,
			"timeout":            time.Duration(libTimeout) * time.Second,
		}

		// 运行库调用
		results, err := o.dispatcher.RunLib(domain, options)
		if err != nil {
			logger.Errorf("Failed to run library call for %s: %v", domain, err)
			continue
		}

		// 处理库调用结果
		o.processLibResults(domain, results)
	}

	// 导出结果
	if err := o.output.Export(); err != nil {
		return fmt.Errorf("failed to export results: %v", err)
	}

	// 显示统计信息
	o.showStats()

	return nil
}

// configParam 配置参数
func (o *OneForAll) configParam() {
	// 设置输出格式
	if outputFmt != "" {
		o.config.ResultSaveFormat = outputFmt
	}
	if path != "" {
		o.config.ResultSavePath = path
	}
	o.config.ResultExportAlive = alive

	// 设置模块开关
	if !brute {
		o.config.EnableBruteModule = false
	}
	if !dns {
		o.config.EnableDNSResolve = false
	}
	if !req {
		o.config.EnableHTTPRequest = false
	}
	if !takeover {
		o.config.EnableTakeoverCheck = false
	}

	// 新架构模块开关
	if !searchModules {
		o.config.EnableSearchModules = false
	}
	if !dnsLookup {
		o.config.EnableDNSLookup = false
	}
	if !resolve {
		o.config.EnableResolve = false
	}
	if !checkModules {
		o.config.EnableCheckModules = false
	}
	if !crawlModules {
		o.config.EnableCrawlModules = false
	}
	if !enrichModules {
		o.config.EnableEnrichModules = false
	}
}

// loadDomains 加载域名
func (o *OneForAll) loadDomains() error {
	if target != "" {
		o.domains = append(o.domains, target)
	}
	if targets != "" {
		domains, err := utils.LoadDomainsFromFile(targets)
		if err != nil {
			return fmt.Errorf("failed to load domains from file: %v", err)
		}
		o.domains = append(o.domains, domains...)
	}

	if len(o.domains) == 0 {
		return fmt.Errorf("no valid domains provided")
	}

	return nil
}

// registerModules 注册模块
func (o *OneForAll) registerModules() {
	logger.Info("Registering modules...")

	// 注册搜索引擎模块
	o.registerSearchModules()

	// 注册数据集模块
	o.registerDatasetModules()

	// 注册证书模块
	o.registerCertificateModules()

	// 注册检查模块
	o.registerCheckModules()

	// 注册爬虫模块
	o.registerCrawlModules()

	// 注册 DNS 查询模块
	o.registerDNSQueryModules()

	// 注册情报模块
	o.registerIntelligenceModules()

	// 注册爆破模块
	o.registerBruteModules()

	// 注册丰富模块
	o.registerEnrichModules()
}

// registerSearchModules 注册搜索引擎模块
func (o *OneForAll) registerSearchModules() {
	// 基础搜索引擎
	o.dispatcher.RegisterModule(search.NewGoogle(o.config))
	o.dispatcher.RegisterModule(search.NewBing(o.config))
	o.dispatcher.RegisterModule(search.NewBaidu(o.config))
	o.dispatcher.RegisterModule(search.NewYahoo(o.config))
	o.dispatcher.RegisterModule(search.NewAsk(o.config))
	o.dispatcher.RegisterModule(search.NewSogou(o.config))
	o.dispatcher.RegisterModule(search.NewYandex(o.config))
	o.dispatcher.RegisterModule(search.NewGitee(o.config))
	o.dispatcher.RegisterModule(search.NewSO(o.config))
	o.dispatcher.RegisterModule(search.NewWZSearch(o.config))

	// API 搜索引擎
	o.dispatcher.RegisterModule(search.NewGitHub(o.config))
	o.dispatcher.RegisterModule(search.NewShodan(o.config))
	o.dispatcher.RegisterModule(search.NewFofa(o.config))
	o.dispatcher.RegisterModule(search.NewHunter(o.config))
	o.dispatcher.RegisterModule(search.NewQuake(o.config))
	o.dispatcher.RegisterModule(search.NewZoomEye(o.config))
	o.dispatcher.RegisterModule(search.NewBingAPI(o.config))
	o.dispatcher.RegisterModule(search.NewGoogleAPI(o.config))
}

// registerDatasetModules 注册数据集模块
func (o *OneForAll) registerDatasetModules() {
	o.dispatcher.RegisterModule(datasets.NewDNSDumpster(o.config))
	o.dispatcher.RegisterModule(datasets.NewSecurityTrails(o.config))
	o.dispatcher.RegisterModule(datasets.NewAnubis(o.config))
	o.dispatcher.RegisterModule(datasets.NewBeVigil(o.config))
	o.dispatcher.RegisterModule(datasets.NewBinaryEdge(o.config))
	o.dispatcher.RegisterModule(datasets.NewChinaz(o.config))
	o.dispatcher.RegisterModule(datasets.NewChinazAPI(o.config))
	o.dispatcher.RegisterModule(datasets.NewCircl(o.config))
	o.dispatcher.RegisterModule(datasets.NewCloudflare(o.config))
	o.dispatcher.RegisterModule(datasets.NewDNSDB(o.config))
	o.dispatcher.RegisterModule(datasets.NewDNSGrep(o.config))
	o.dispatcher.RegisterModule(datasets.NewFullHunt(o.config))
	o.dispatcher.RegisterModule(datasets.NewHackerTarget(o.config))
	o.dispatcher.RegisterModule(datasets.NewIP138(o.config))
	o.dispatcher.RegisterModule(datasets.NewIPv4Info(o.config))
	o.dispatcher.RegisterModule(datasets.NewNetcraft(o.config))
	o.dispatcher.RegisterModule(datasets.NewPassiveDNS(o.config))
	o.dispatcher.RegisterModule(datasets.NewQianxun(o.config))
	o.dispatcher.RegisterModule(datasets.NewRapidDNS(o.config))
	o.dispatcher.RegisterModule(datasets.NewRiddler(o.config))
	o.dispatcher.RegisterModule(datasets.NewRobtex(o.config))
	o.dispatcher.RegisterModule(datasets.NewSiteDossier(o.config))
	o.dispatcher.RegisterModule(datasets.NewSpyse(o.config))
	o.dispatcher.RegisterModule(datasets.NewSublist3r(o.config))
	o.dispatcher.RegisterModule(datasets.NewURLScan(o.config))
}

// registerCertificateModules 注册证书模块
func (o *OneForAll) registerCertificateModules() {
	o.dispatcher.RegisterModule(certificates.NewCensys(o.config))
	o.dispatcher.RegisterModule(certificates.NewCertSpotter(o.config))
	o.dispatcher.RegisterModule(certificates.NewCRTSh(o.config))
	o.dispatcher.RegisterModule(certificates.NewGoogle(o.config))
	o.dispatcher.RegisterModule(certificates.NewMySSL(o.config))
	o.dispatcher.RegisterModule(certificates.NewRacent(o.config))
}

// registerCheckModules 注册检查模块
func (o *OneForAll) registerCheckModules() {
	o.dispatcher.RegisterModule(check.NewAXFR(o.config))
	o.dispatcher.RegisterModule(check.NewCDX(o.config))
	o.dispatcher.RegisterModule(check.NewCert(o.config))
	o.dispatcher.RegisterModule(check.NewCSP(o.config))
	o.dispatcher.RegisterModule(check.NewNSEC(o.config))
	o.dispatcher.RegisterModule(check.NewRobots(o.config))
	o.dispatcher.RegisterModule(check.NewSitemap(o.config))
}

// registerCrawlModules 注册爬虫模块
func (o *OneForAll) registerCrawlModules() {
	o.dispatcher.RegisterModule(crawl.NewArchive(o.config))
	o.dispatcher.RegisterModule(crawl.NewCommonCrawl(o.config))
}

// registerDNSQueryModules 注册 DNS 查询模块
func (o *OneForAll) registerDNSQueryModules() {
	o.dispatcher.RegisterModule(dnsquery.NewNS(o.config))
	o.dispatcher.RegisterModule(dnsquery.NewMX(o.config))
	o.dispatcher.RegisterModule(dnsquery.NewSOA(o.config))
	o.dispatcher.RegisterModule(dnsquery.NewSPF(o.config))
	o.dispatcher.RegisterModule(dnsquery.NewTXT(o.config))
}

// registerIntelligenceModules 注册情报模块
func (o *OneForAll) registerIntelligenceModules() {
	o.dispatcher.RegisterModule(intelligence.NewAlienVault(o.config))
	o.dispatcher.RegisterModule(intelligence.NewRiskIQ(o.config))
	o.dispatcher.RegisterModule(intelligence.NewThreatBook(o.config))
	o.dispatcher.RegisterModule(intelligence.NewThreatMiner(o.config))
	o.dispatcher.RegisterModule(intelligence.NewVirusTotal(o.config))
	o.dispatcher.RegisterModule(intelligence.NewVirusTotalAPI(o.config))
}

// registerBruteModules 注册爆破模块
func (o *OneForAll) registerBruteModules() {
	logger.Debugf("Registering brute force modules...")

	// 注册爆破模块
	bruteModule := brutepkg.NewBrute(o.config)
	o.dispatcher.RegisterModule(bruteModule)
	logger.Infof("Registered brute force module: %s", bruteModule.Name())

	// 注册alt模块
	altModule := alt.NewAlt(o.config)
	o.dispatcher.RegisterModule(altModule)
	logger.Infof("Registered alt module: %s", altModule.Name())

	logger.Debugf("Brute force modules registration completed")
}

// registerEnrichModules 注册丰富模块
func (o *OneForAll) registerEnrichModules() {
	// 注册enrich模块
	enrichModule := enrich.NewEnrich(o.config)
	o.dispatcher.RegisterModule(enrichModule)
}

// processResults 处理结果
func (o *OneForAll) processResults(domain string, results map[core.ModuleType][]core.SubdomainResult, validationResults []validator.ValidationResult) {
	// 首先添加验证结果
	if len(validationResults) > 0 {
		logger.Infof("Adding %d validation results for %s", len(validationResults), domain)
		o.output.AddValidationResults(validationResults)
	}

	// 然后添加其他模块的结果
	for moduleType, subdomainResults := range results {
		logger.Infof("Module type %s found %d subdomains for %s", moduleType, len(subdomainResults), domain)

		// 直接添加SubdomainResult结构
		for _, result := range subdomainResults {
			o.output.AddResult(result)
		}
	}
}

// showStats 显示统计信息
func (o *OneForAll) showStats() {
	stats := o.output.GetStats()
	logger.Info("=== Collection Statistics ===")
	logger.Infof("Total subdomains: %d", stats["total"])
	logger.Infof("Alive subdomains: %d", stats["alive"])
	logger.Infof("Dead subdomains: %d", stats["dead"])

	if sources, ok := stats["sources"].(map[string]int); ok {
		logger.Info("Sources breakdown:")
		for source, count := range sources {
			logger.Infof("  %s: %d", source, count)
		}
	}

	if providers, ok := stats["providers"].(map[string]int); ok {
		logger.Info("IP providers breakdown:")
		for provider, count := range providers {
			logger.Infof("  %s: %d", provider, count)
		}
	}

	logger.Infof("Results saved to: %s", o.output.GetOutputPath())
}

// version 显示版本信息
func (o *OneForAll) version() {
	fmt.Println("OneForAll-Go v1.0.0")
	fmt.Println("A powerful subdomain integration tool")
	fmt.Println("GitHub: https://github.com/oneforall-go")
}

// check 检查环境
func (o *OneForAll) check() error {
	logger.Info("Checking environment...")

	// 检查网络连接
	if !utils.CheckInternetConnection() {
		return fmt.Errorf("no internet connection")
	}

	// 检查配置
	if o.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// 检查输出目录
	if err := os.MkdirAll(o.config.ResultSavePath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	logger.Info("Environment check passed")
	return nil
}

// 创建根命令
var rootCmd = &cobra.Command{
	Use:   "oneforall-go",
	Short: "OneForAll-Go is a powerful subdomain enumeration tool",
	Long: `OneForAll-Go is a powerful subdomain enumeration tool written in Go.
It supports various modules for subdomain discovery including search, datasets, 
certificates, crawl, check, intelligence, brute force, and validation.`,
}

// 创建运行命令
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run subdomain enumeration",
	Long:  `Run subdomain enumeration with specified target domain or file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOneForAll()
	},
}

// 创建版本命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		oneforall := NewOneForAll()
		oneforall.version()
	},
}

// 创建检查命令
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		oneforall := NewOneForAll()
		return oneforall.check()
	},
}

// 创建runlib命令
var runLibCmd = &cobra.Command{
	Use:   "runlib",
	Short: "Run subdomain collection as library",
	Long:  `Run subdomain collection as library with configurable options`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibCall()
	},
}

// runLib 运行库调用
func runOneForAll() error {
	oneforall := NewOneForAll()
	return oneforall.run()
}

func runLibCall() error {
	oneforall := NewOneForAll()
	return oneforall.runLib()
}

// processLibResults 处理库调用结果
func (o *OneForAll) processLibResults(domain string, results []core.SubdomainResult) {
	logger.Infof("Processing %d results for domain %s", len(results), domain)

	// 直接添加所有结果
	for _, result := range results {
		o.output.AddResult(result)
	}

	// 显示结果统计
	aliveCount := 0
	for _, result := range results {
		if result.Alive {
			aliveCount++
		}
	}

	logger.Infof("Domain %s: %d total subdomains, %d alive (%.1f%%)",
		domain, len(results), aliveCount, float64(aliveCount)/float64(len(results))*100)
}

func init() {
	// 加载环境变量
	loadEnvironment()

	// 根据环境变量设置日志级别
	logLevel := "info"
	if debugEnabled := os.Getenv("DEBUG_ENABLED"); debugEnabled == "true" {
		logLevel = "debug"
	} else if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		logLevel = logLevelEnv
	}

	// 初始化日志
	logger.Init(logLevel, "")

	// 设置根命令
	rootCmd.AddCommand(runCmd, versionCmd, checkCmd, runLibCmd)

	// 设置run命令的参数
	runCmd.Flags().StringVarP(&target, "target", "t", "", "Target domain (required)")
	runCmd.Flags().StringVarP(&targets, "targets", "f", "", "目标域名文件")
	runCmd.Flags().BoolVarP(&brute, "brute", "b", false, "启用爆破模块")
	runCmd.Flags().BoolVarP(&dns, "dns", "d", false, "启用DNS解析")
	runCmd.Flags().BoolVarP(&req, "req", "r", false, "启用HTTP请求")
	runCmd.Flags().StringVarP(&port, "port", "p", "80,443", "HTTP请求端口")
	runCmd.Flags().BoolVarP(&alive, "alive", "a", false, "只导出存活域名")
	runCmd.Flags().StringVarP(&outputFmt, "format", "o", "csv", "输出格式 (csv/json)")
	runCmd.Flags().StringVarP(&path, "path", "", "results", "结果保存路径")
	runCmd.Flags().BoolVarP(&takeover, "takeover", "", false, "启用接管检查")
	runCmd.Flags().BoolVarP(&show, "show", "s", false, "显示帮助信息")

	// 新架构参数
	runCmd.Flags().BoolVarP(&searchModules, "search", "", false, "启用搜索模块")
	runCmd.Flags().BoolVarP(&dnsLookup, "dns-lookup", "", false, "启用DNS查询")
	runCmd.Flags().BoolVarP(&resolve, "resolve", "", false, "启用解析模块")
	runCmd.Flags().BoolVarP(&checkModules, "check", "", false, "启用检查模块")
	runCmd.Flags().BoolVarP(&crawlModules, "crawl", "", false, "启用爬虫模块")
	runCmd.Flags().BoolVarP(&enrichModules, "enrich", "", false, "启用丰富模块")

	// 库调用参数
	runLibCmd.Flags().StringVarP(&target, "target", "t", "", "Target domain (required)")
	runLibCmd.Flags().BoolVar(&enableValidation, "enable-validation", true, "Enable domain validation")
	runLibCmd.Flags().BoolVar(&enableBruteForce, "enable-brute-force", false, "Enable brute force attack")
	runLibCmd.Flags().IntVar(&libConcurrency, "concurrency", 10, "Concurrency level")
	runLibCmd.Flags().IntVar(&libTimeout, "timeout", 60, "Timeout in seconds")
	runLibCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")
	runLibCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	// 设置必需参数
	runCmd.MarkFlagRequired("target")

	// runLibCmd 参数
	// runLibCmd.Flags().Bool("enable-validation", true, "Enable validation results")
	// runLibCmd.Flags().Bool("enable-brute-force", false, "Enable brute force module")
	// runLibCmd.Flags().Int("concurrency", 10, "Number of concurrent requests")
	// runLibCmd.Flags().Duration("timeout", 0, "Request timeout")
}

// loadEnvironment 加载环境变量
func loadEnvironment() {
	// 尝试加载.env文件
	if err := godotenv.Load(); err != nil {
		// 如果.env文件不存在，尝试加载env.example
		if err := godotenv.Load("env.example"); err != nil {
			// 如果都不存在，使用默认配置
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
