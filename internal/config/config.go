package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 配置结构
type Config struct {
	// 调试和日志配置
	DebugEnabled   bool   `mapstructure:"debug_enabled"`
	LogLevel       string `mapstructure:"log_level"`
	LogFile        string `mapstructure:"log_file"`
	VerboseLogging bool   `mapstructure:"verbose_logging"`

	// 模块开关
	EnableBruteModule   bool `mapstructure:"enable_brute_module"`
	EnableDNSResolve    bool `mapstructure:"enable_dns_resolve"`
	EnableHTTPRequest   bool `mapstructure:"enable_http_request"`
	EnableTakeoverCheck bool `mapstructure:"enable_takeover_check"`
	EnableFinderModule  bool `mapstructure:"enable_finder_module"`
	EnableAltdnsModule  bool `mapstructure:"enable_altdns_module"`
	EnableEnrichModule  bool `mapstructure:"enable_enrich_module"`
	EnableSearchModules bool `mapstructure:"enable_search_modules"`
	EnableDNSLookup     bool `mapstructure:"enable_dns_lookup"`
	EnableResolve       bool `mapstructure:"enable_resolve"`
	EnableCheckModules  bool `mapstructure:"enable_check_modules"`
	EnableCrawlModules  bool `mapstructure:"enable_crawl_modules"`
	EnableEnrichModules bool `mapstructure:"enable_enrich_modules"`

	// 搜索配置
	EnableRecursiveSearch bool `mapstructure:"enable_recursive_search"`
	SearchRecursiveTimes  int  `mapstructure:"search_recursive_times"`
	EnableFullSearch      bool `mapstructure:"enable_full_search"`

	// 结果配置
	ResultSaveFormat  string `mapstructure:"result_save_format"`
	ResultSavePath    string `mapstructure:"result_save_path"`
	ResultExportAlive bool   `mapstructure:"result_export_alive"`

	// HTTP配置
	HTTPRequestPort string `mapstructure:"http_request_port"`

	// DNS配置
	DNSResolveTimeout     int `mapstructure:"dns_resolve_timeout"`
	DNSResolveConcurrency int `mapstructure:"dns_resolve_concurrency"`

	// 爆破配置
	BruteConcurrency int `mapstructure:"brute_concurrency"`
	BruteTimeout     int `mapstructure:"brute_timeout"`

	// 域名验证配置
	EnableDomainValidation bool  `mapstructure:"enable_domain_validation"`
	ValidationConcurrency  int   `mapstructure:"validation_concurrency"`
	ValidationTimeout      int   `mapstructure:"validation_timeout"`
	ExcludePrivateIP       bool  `mapstructure:"exclude_private_ip"`
	ExportAliveOnly        bool  `mapstructure:"export_alive_only"`
	EnableTCPValidation    bool  `mapstructure:"enable_tcp_validation"`
	TCPValidationPorts     []int `mapstructure:"tcp_validation_ports"`

	// 多线程配置
	MultiThreading MultiThreadingConfig `mapstructure:"multi_threading"`

	// API密钥
	APIKeys map[string]string `mapstructure:"api_keys"`

	// 泛解析检测配置
	WildcardTestCount             int     `mapstructure:"wildcard_test_count"`
	WildcardSuccessRateThreshold  float64 `mapstructure:"wildcard_success_rate_threshold"`
	WildcardIPRepeatRateThreshold float64 `mapstructure:"wildcard_ip_repeat_rate_threshold"`

	// 其他配置
	CommonSubnames string `mapstructure:"common_subnames"`
}

// MultiThreadingConfig 多线程配置
type MultiThreadingConfig struct {
	EnableStepExecution bool `mapstructure:"enable_step_execution"`

	// 并发数配置
	FastSearchConcurrency   int `mapstructure:"fast_search_concurrency"`
	DatasetConcurrency      int `mapstructure:"dataset_concurrency"`
	CertificateConcurrency  int `mapstructure:"certificate_concurrency"`
	CrawlConcurrency        int `mapstructure:"crawl_concurrency"`
	DNSLookupConcurrency    int `mapstructure:"dns_lookup_concurrency"`
	FileCheckConcurrency    int `mapstructure:"file_check_concurrency"`
	IntelligenceConcurrency int `mapstructure:"intelligence_concurrency"`
	BruteForceConcurrency   int `mapstructure:"brute_force_concurrency"`
	EnrichConcurrency       int `mapstructure:"enrich_concurrency"`

	// 超时配置
	FastSearchTimeout   int `mapstructure:"fast_search_timeout"`
	DatasetTimeout      int `mapstructure:"dataset_timeout"`
	CertificateTimeout  int `mapstructure:"certificate_timeout"`
	CrawlTimeout        int `mapstructure:"crawl_timeout"`
	DNSLookupTimeout    int `mapstructure:"dns_lookup_timeout"`
	FileCheckTimeout    int `mapstructure:"file_check_timeout"`
	IntelligenceTimeout int `mapstructure:"intelligence_timeout"`
	BruteForceTimeout   int `mapstructure:"brute_force_timeout"`
	EnrichTimeout       int `mapstructure:"enrich_timeout"`

	// 启用配置
	EnableFastSearch   bool `mapstructure:"enable_fast_search"`
	EnableDataset      bool `mapstructure:"enable_dataset"`
	EnableCertificate  bool `mapstructure:"enable_certificate"`
	EnableCrawl        bool `mapstructure:"enable_crawl"`
	EnableDNSLookup    bool `mapstructure:"enable_dns_lookup"`
	EnableFileCheck    bool `mapstructure:"enable_file_check"`
	EnableIntelligence bool `mapstructure:"enable_intelligence"`
	EnableBruteForce   bool `mapstructure:"enable_brute_force"`
	EnableEnrich       bool `mapstructure:"enable_enrich"`
}

var config *Config

// GetConfig 获取配置实例
func GetConfig() *Config {
	if config == nil {
		config = loadConfig()
	}
	return config
}

// loadConfig 加载配置
func loadConfig() *Config {
	// 加载.env文件
	loadEnvFile()

	cfg := &Config{}
	setDefaults(cfg)
	loadFromEnv(cfg)
	loadFromYAML(cfg)

	return cfg
}

// loadEnvFile 加载.env文件
func loadEnvFile() {
	// 尝试加载.env文件
	if err := godotenv.Load(); err != nil {
		// 如果.env文件不存在，尝试加载env.example
		if err := godotenv.Load("env.example"); err != nil {
			// 如果都不存在，使用默认配置
		}
	}
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	// 调试和日志配置
	cfg.DebugEnabled = false
	cfg.LogLevel = "info"
	cfg.LogFile = "logs/oneforall.log"
	cfg.VerboseLogging = false

	// 模块开关
	cfg.EnableBruteModule = true
	cfg.EnableDNSResolve = true
	cfg.EnableHTTPRequest = true
	cfg.EnableTakeoverCheck = false
	cfg.EnableFinderModule = true
	cfg.EnableAltdnsModule = true
	cfg.EnableEnrichModule = true
	cfg.EnableSearchModules = true
	cfg.EnableDNSLookup = true
	cfg.EnableResolve = true
	cfg.EnableCheckModules = true
	cfg.EnableCrawlModules = true
	cfg.EnableEnrichModules = true

	// 搜索配置
	cfg.EnableRecursiveSearch = false
	cfg.SearchRecursiveTimes = 1
	cfg.EnableFullSearch = true

	// 结果配置
	cfg.ResultSaveFormat = "csv"
	cfg.ResultSavePath = "results"
	cfg.ResultExportAlive = true

	// HTTP配置
	cfg.HTTPRequestPort = "80,443"

	// DNS配置
	cfg.DNSResolveTimeout = 10
	cfg.DNSResolveConcurrency = 100

	// 爆破配置
	cfg.BruteConcurrency = 2000
	cfg.BruteTimeout = 300

	// 域名验证配置
	cfg.EnableDomainValidation = true
	cfg.ValidationConcurrency = 50
	cfg.ValidationTimeout = 30
	cfg.ExcludePrivateIP = true
	cfg.ExportAliveOnly = true
	cfg.EnableTCPValidation = true
	cfg.TCPValidationPorts = []int{80, 443, 8080, 8443}

	// 多线程配置
	cfg.MultiThreading = MultiThreadingConfig{
		EnableStepExecution: true,

		// 并发数
		FastSearchConcurrency:   10,
		DatasetConcurrency:      20,
		CertificateConcurrency:  15,
		CrawlConcurrency:        10,
		DNSLookupConcurrency:    50,
		FileCheckConcurrency:    10,
		IntelligenceConcurrency: 15,
		BruteForceConcurrency:   2000,
		EnrichConcurrency:       20,

		// 超时
		FastSearchTimeout:   30,
		DatasetTimeout:      60,
		CertificateTimeout:  45,
		CrawlTimeout:        30,
		DNSLookupTimeout:    30,
		FileCheckTimeout:    30,
		IntelligenceTimeout: 45,
		BruteForceTimeout:   300,
		EnrichTimeout:       60,

		// 启用
		EnableFastSearch:   true,
		EnableDataset:      true,
		EnableCertificate:  true,
		EnableCrawl:        true,
		EnableDNSLookup:    true,
		EnableFileCheck:    true,
		EnableIntelligence: true,
		EnableBruteForce:   true,
		EnableEnrich:       true,
	}

	// API密钥
	cfg.APIKeys = make(map[string]string)

	// 泛解析检测配置
	cfg.WildcardTestCount = 20
	cfg.WildcardSuccessRateThreshold = 90.0
	cfg.WildcardIPRepeatRateThreshold = 50.0

	// 其他配置
	cfg.CommonSubnames = "www,mail,ftp,admin,blog,api,dev,stage,prod,app,web,cdn,static,img,css,js,docs,help,support"
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	// 调试和日志配置
	if val := getEnvBool("DEBUG_ENABLED"); val != nil {
		cfg.DebugEnabled = *val
	}
	if val := getEnvString("LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}
	if val := getEnvString("LOG_FILE"); val != "" {
		cfg.LogFile = val
	}
	if val := getEnvBool("VERBOSE_LOGGING"); val != nil {
		cfg.VerboseLogging = *val
	}

	// 模块开关
	if val := getEnvBool("ENABLE_BRUTE_MODULE"); val != nil {
		cfg.EnableBruteModule = *val
	}
	if val := getEnvBool("ENABLE_DNS_RESOLVE"); val != nil {
		cfg.EnableDNSResolve = *val
	}
	if val := getEnvBool("ENABLE_HTTP_REQUEST"); val != nil {
		cfg.EnableHTTPRequest = *val
	}
	if val := getEnvBool("ENABLE_TAKEOVER_CHECK"); val != nil {
		cfg.EnableTakeoverCheck = *val
	}
	if val := getEnvBool("ENABLE_FINDER_MODULE"); val != nil {
		cfg.EnableFinderModule = *val
	}
	if val := getEnvBool("ENABLE_ALTDNS_MODULE"); val != nil {
		cfg.EnableAltdnsModule = *val
	}
	if val := getEnvBool("ENABLE_ENRICH_MODULE"); val != nil {
		cfg.EnableEnrichModule = *val
	}
	if val := getEnvBool("ENABLE_SEARCH_MODULES"); val != nil {
		cfg.EnableSearchModules = *val
	}
	if val := getEnvBool("ENABLE_DNS_LOOKUP"); val != nil {
		cfg.EnableDNSLookup = *val
	}
	if val := getEnvBool("ENABLE_RESOLVE"); val != nil {
		cfg.EnableResolve = *val
	}
	if val := getEnvBool("ENABLE_CHECK_MODULES"); val != nil {
		cfg.EnableCheckModules = *val
	}
	if val := getEnvBool("ENABLE_CRAWL_MODULES"); val != nil {
		cfg.EnableCrawlModules = *val
	}
	if val := getEnvBool("ENABLE_ENRICH_MODULES"); val != nil {
		cfg.EnableEnrichModules = *val
	}

	// 搜索配置
	if val := getEnvBool("ENABLE_RECURSIVE_SEARCH"); val != nil {
		cfg.EnableRecursiveSearch = *val
	}
	if val := getEnvInt("SEARCH_RECURSIVE_TIMES"); val != nil {
		cfg.SearchRecursiveTimes = *val
	}
	if val := getEnvBool("ENABLE_FULL_SEARCH"); val != nil {
		cfg.EnableFullSearch = *val
	}

	// 结果配置
	if val := getEnvString("RESULT_SAVE_FORMAT"); val != "" {
		cfg.ResultSaveFormat = val
	}
	if val := getEnvString("RESULT_SAVE_PATH"); val != "" {
		cfg.ResultSavePath = val
	}
	if val := getEnvBool("RESULT_EXPORT_ALIVE"); val != nil {
		cfg.ResultExportAlive = *val
	}

	// HTTP配置
	if val := getEnvString("HTTP_REQUEST_PORT"); val != "" {
		cfg.HTTPRequestPort = val
	}

	// DNS配置
	if val := getEnvInt("DNS_RESOLVE_TIMEOUT"); val != nil {
		cfg.DNSResolveTimeout = *val
	}
	if val := getEnvInt("DNS_RESOLVE_CONCURRENCY"); val != nil {
		cfg.DNSResolveConcurrency = *val
	}

	// 爆破配置
	if val := getEnvInt("BRUTE_CONCURRENCY"); val != nil {
		cfg.BruteConcurrency = *val
	}
	if val := getEnvInt("BRUTE_TIMEOUT"); val != nil {
		cfg.BruteTimeout = *val
	}

	// 域名验证配置
	if val := getEnvBool("ENABLE_DOMAIN_VALIDATION"); val != nil {
		cfg.EnableDomainValidation = *val
	}
	if val := getEnvInt("VALIDATION_CONCURRENCY"); val != nil {
		cfg.ValidationConcurrency = *val
	}
	if val := getEnvInt("VALIDATION_TIMEOUT"); val != nil {
		cfg.ValidationTimeout = *val
	}
	if val := getEnvBool("EXCLUDE_PRIVATE_IP"); val != nil {
		cfg.ExcludePrivateIP = *val
	}
	if val := getEnvBool("EXPORT_ALIVE_ONLY"); val != nil {
		cfg.ExportAliveOnly = *val
	}
	if val := getEnvBool("ENABLE_TCP_VALIDATION"); val != nil {
		cfg.EnableTCPValidation = *val
	}
	if val := getEnvString("TCP_VALIDATION_PORTS"); val != "" {
		cfg.TCPValidationPorts = parsePorts(val)
	}

	// 多线程配置
	if val := getEnvBool("ENABLE_STEP_EXECUTION"); val != nil {
		cfg.MultiThreading.EnableStepExecution = *val
	}

	// 并发数配置
	if val := getEnvInt("FAST_SEARCH_CONCURRENCY"); val != nil {
		cfg.MultiThreading.FastSearchConcurrency = *val
	}
	if val := getEnvInt("DATASET_CONCURRENCY"); val != nil {
		cfg.MultiThreading.DatasetConcurrency = *val
	}
	if val := getEnvInt("CERTIFICATE_CONCURRENCY"); val != nil {
		cfg.MultiThreading.CertificateConcurrency = *val
	}
	if val := getEnvInt("CRAWL_CONCURRENCY"); val != nil {
		cfg.MultiThreading.CrawlConcurrency = *val
	}
	if val := getEnvInt("DNS_LOOKUP_CONCURRENCY"); val != nil {
		cfg.MultiThreading.DNSLookupConcurrency = *val
	}
	if val := getEnvInt("FILE_CHECK_CONCURRENCY"); val != nil {
		cfg.MultiThreading.FileCheckConcurrency = *val
	}
	if val := getEnvInt("INTELLIGENCE_CONCURRENCY"); val != nil {
		cfg.MultiThreading.IntelligenceConcurrency = *val
	}
	if val := getEnvInt("BRUTE_FORCE_CONCURRENCY"); val != nil {
		cfg.MultiThreading.BruteForceConcurrency = *val
	}
	if val := getEnvInt("ENRICH_CONCURRENCY"); val != nil {
		cfg.MultiThreading.EnrichConcurrency = *val
	}

	// 超时配置
	if val := getEnvInt("FAST_SEARCH_TIMEOUT"); val != nil {
		cfg.MultiThreading.FastSearchTimeout = *val
	}
	if val := getEnvInt("DATASET_TIMEOUT"); val != nil {
		cfg.MultiThreading.DatasetTimeout = *val
	}
	if val := getEnvInt("CERTIFICATE_TIMEOUT"); val != nil {
		cfg.MultiThreading.CertificateTimeout = *val
	}
	if val := getEnvInt("CRAWL_TIMEOUT"); val != nil {
		cfg.MultiThreading.CrawlTimeout = *val
	}
	if val := getEnvInt("DNS_LOOKUP_TIMEOUT"); val != nil {
		cfg.MultiThreading.DNSLookupTimeout = *val
	}
	if val := getEnvInt("FILE_CHECK_TIMEOUT"); val != nil {
		cfg.MultiThreading.FileCheckTimeout = *val
	}
	if val := getEnvInt("INTELLIGENCE_TIMEOUT"); val != nil {
		cfg.MultiThreading.IntelligenceTimeout = *val
	}
	if val := getEnvInt("BRUTE_FORCE_TIMEOUT"); val != nil {
		cfg.MultiThreading.BruteForceTimeout = *val
	}
	if val := getEnvInt("ENRICH_TIMEOUT"); val != nil {
		cfg.MultiThreading.EnrichTimeout = *val
	}

	// 启用配置
	if val := getEnvBool("ENABLE_FAST_SEARCH"); val != nil {
		cfg.MultiThreading.EnableFastSearch = *val
	}
	if val := getEnvBool("ENABLE_DATASET"); val != nil {
		cfg.MultiThreading.EnableDataset = *val
	}
	if val := getEnvBool("ENABLE_CERTIFICATE"); val != nil {
		cfg.MultiThreading.EnableCertificate = *val
	}
	if val := getEnvBool("ENABLE_CRAWL"); val != nil {
		cfg.MultiThreading.EnableCrawl = *val
	}
	if val := getEnvBool("ENABLE_DNS_LOOKUP"); val != nil {
		cfg.MultiThreading.EnableDNSLookup = *val
	}
	if val := getEnvBool("ENABLE_FILE_CHECK"); val != nil {
		cfg.MultiThreading.EnableFileCheck = *val
	}
	if val := getEnvBool("ENABLE_INTELLIGENCE"); val != nil {
		cfg.MultiThreading.EnableIntelligence = *val
	}
	if val := getEnvBool("ENABLE_BRUTE_FORCE"); val != nil {
		cfg.MultiThreading.EnableBruteForce = *val
	}
	if val := getEnvBool("ENABLE_ENRICH"); val != nil {
		cfg.MultiThreading.EnableEnrich = *val
	}

	// API密钥
	apiKeys := []string{
		"GITHUB_API_TOKEN", "SHODAN_API_KEY", "FOFA_API_EMAIL", "FOFA_API_KEY",
		"HUNTER_API_KEY", "QUAKE_API_KEY", "ZOOMEYE_API_KEY", "VIRUSTOTAL_API_KEY",
		"SECURITYTRAILS_API_KEY", "CENSYS_API_KEY", "BINARYEDGE_API_KEY",
		"SPYSE_API_KEY", "RISKIQ_API_KEY", "THREATBOOK_API_KEY", "ANUBIS_API_KEY",
		"BEVIGIL_API_KEY",
	}

	for _, key := range apiKeys {
		if val := getEnvString(key); val != "" {
			cfg.APIKeys[strings.ToLower(key)] = val
		}
	}

	// 泛解析检测配置
	if val := getEnvInt("WILDCARD_TEST_COUNT"); val != nil {
		cfg.WildcardTestCount = *val
	}
	if val := getEnvFloat("WILDCARD_SUCCESS_RATE_THRESHOLD"); val != nil {
		cfg.WildcardSuccessRateThreshold = *val
	}
	if val := getEnvFloat("WILDCARD_IP_REPEAT_RATE_THRESHOLD"); val != nil {
		cfg.WildcardIPRepeatRateThreshold = *val
	}

	// 其他配置
	if val := getEnvString("COMMON_SUBNAMES"); val != "" {
		cfg.CommonSubnames = val
	}
}

// loadFromYAML 从YAML文件加载配置（保留兼容性）
func loadFromYAML(cfg *Config) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("data/config")

	if err := viper.ReadInConfig(); err == nil {
		// 如果YAML文件存在，使用YAML配置覆盖环境变量
		viper.Unmarshal(cfg)
	}
}

// 辅助函数
func getEnvString(key string) string {
	return os.Getenv(key)
}

func getEnvBool(key string) *bool {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return nil
	}
	return &b
}

func getEnvInt(key string) *int {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	return &i
}

func getEnvFloat(key string) *float64 {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil
	}
	return &f
}

func parsePorts(portsStr string) []int {
	var ports []int
	for _, portStr := range strings.Split(portsStr, ",") {
		if port, err := strconv.Atoi(strings.TrimSpace(portStr)); err == nil {
			ports = append(ports, port)
		}
	}
	return ports
}
