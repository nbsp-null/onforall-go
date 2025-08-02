package collector

import (
	"fmt"
	"time"

	"github.com/oneforall-go/internal/brute"
	"github.com/oneforall-go/internal/certificates"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/dns"
	"github.com/oneforall-go/internal/http"
	"github.com/oneforall-go/internal/modules"
	"github.com/oneforall-go/internal/osint"
	"github.com/oneforall-go/internal/search"
	"github.com/oneforall-go/pkg/logger"
)

// Subdomain 子域信息结构
type Subdomain struct {
	Subdomain string   `json:"subdomain" csv:"subdomain"`
	IP        []string `json:"ip" csv:"ip"`
	Status    int      `json:"status" csv:"status"`
	Title     string   `json:"title" csv:"title"`
	Port      int      `json:"port" csv:"port"`
	Alive     bool     `json:"alive" csv:"alive"`
	Source    string   `json:"source" csv:"source"`
	Time      string   `json:"time" csv:"time"`
}

// Collector 子域收集器
type Collector struct {
	domain        string
	dnsClient     *dns.Client
	reflectClient *dns.ReflectClient
	httpClient    *http.Client
	bruteClient   *brute.Client
	certClient    *certificates.CertificateClient
	osintClient   *osint.OSINTClient
	searchClient  *search.SearchClient
	moduleManager *modules.Manager
	results       []Subdomain
}

// NewCollector 创建新的收集器
func NewCollector(domain string) *Collector {
	cfg := config.GetConfig()

	// 创建模块管理器
	moduleManager := modules.NewManager(cfg)

	return &Collector{
		domain:        domain,
		dnsClient:     dns.NewClient(cfg.DNSResolveTimeout, cfg.DNSResolveConcurrency),
		reflectClient: dns.NewReflectClient(cfg.DNSResolveTimeout, cfg.DNSResolveConcurrency),
		httpClient:    http.NewClient(),
		bruteClient:   brute.NewClient(cfg.BruteConcurrency, cfg.BruteTimeout),
		certClient:    certificates.NewCertificateClient(cfg.DNSResolveTimeout),
		osintClient:   osint.NewOSINTClient(cfg.DNSResolveTimeout),
		searchClient:  search.NewSearchClient(cfg.DNSResolveTimeout),
		moduleManager: moduleManager,
		results:       make([]Subdomain, 0),
	}
}

// Collect 执行子域收集
func (c *Collector) Collect() ([]Subdomain, error) {
	logger.Infof("Starting subdomain collection for domain: %s", c.domain)

	// 从各种来源收集子域
	subdomains := make([]Subdomain, 0)

	// 1. DNS 反射匹配查询
	reflectResults, err := c.collectFromDNSReflect()
	if err != nil {
		logger.Errorf("Failed to collect from DNS reflection: %v", err)
	} else {
		subdomains = append(subdomains, reflectResults...)
	}

	// 2. 证书透明度查询
	certResults, err := c.collectFromCertificates()
	if err != nil {
		logger.Errorf("Failed to collect from certificates: %v", err)
	} else {
		subdomains = append(subdomains, certResults...)
	}

	// 3. OSINT API 查询
	osintResults, err := c.collectFromOSINT()
	if err != nil {
		logger.Errorf("Failed to collect from OSINT APIs: %v", err)
	} else {
		subdomains = append(subdomains, osintResults...)
	}

	// 4. 搜索引擎查询
	searchResults, err := c.collectFromSearchEngines()
	if err != nil {
		logger.Errorf("Failed to collect from search engines: %v", err)
	} else {
		subdomains = append(subdomains, searchResults...)
	}

	// 5. 模块化收集
	moduleResults, err := c.collectFromModules()
	if err != nil {
		logger.Errorf("Failed to collect from modules: %v", err)
	} else {
		subdomains = append(subdomains, moduleResults...)
	}

	// 6. 从 DNS 记录收集
	dnsResults, err := c.collectFromDNS()
	if err != nil {
		logger.Errorf("Failed to collect from DNS records: %v", err)
	} else {
		subdomains = append(subdomains, dnsResults...)
	}

	// 7. 从在线数据集收集
	datasetResults, err := c.collectFromDatasets()
	if err != nil {
		logger.Errorf("Failed to collect from datasets: %v", err)
	} else {
		subdomains = append(subdomains, datasetResults...)
	}

	// 去重
	subdomains = c.deduplicate(subdomains)

	logger.Infof("Collected %d unique subdomains for %s", len(subdomains), c.domain)
	return subdomains, nil
}

// ResolveDNS 解析子域 DNS
func (c *Collector) ResolveDNS(subdomains []Subdomain) ([]Subdomain, error) {
	logger.Info("Starting DNS resolution")

	for i := range subdomains {
		ips, err := c.dnsClient.Resolve(subdomains[i].Subdomain)
		if err != nil {
			logger.Debugf("Failed to resolve %s: %v", subdomains[i].Subdomain, err)
			continue
		}
		subdomains[i].IP = ips
	}

	logger.Info("DNS resolution completed")
	return subdomains, nil
}

// RequestHTTP 执行 HTTP 请求
func (c *Collector) RequestHTTP(subdomains []Subdomain) ([]Subdomain, error) {
	logger.Info("Starting HTTP requests")

	for i := range subdomains {
		if len(subdomains[i].IP) == 0 {
			continue
		}

		status, title, err := c.httpClient.Request(subdomains[i].Subdomain, subdomains[i].IP[0])
		if err != nil {
			logger.Debugf("Failed to request %s: %v", subdomains[i].Subdomain, err)
			continue
		}

		subdomains[i].Status = status
		subdomains[i].Title = title
		subdomains[i].Alive = status > 0
		subdomains[i].Time = time.Now().Format("2006-01-02 15:04:05")
	}

	logger.Info("HTTP requests completed")
	return subdomains, nil
}

// BruteForce 执行暴力破解
func (c *Collector) BruteForce() ([]Subdomain, error) {
	logger.Info("Starting brute force")

	bruteResults, err := c.bruteClient.Brute(c.domain)
	if err != nil {
		return nil, fmt.Errorf("brute force failed: %v", err)
	}

	// 转换 brute.Subdomain 到 collector.Subdomain
	results := make([]Subdomain, len(bruteResults))
	for i, bruteResult := range bruteResults {
		results[i] = Subdomain{
			Subdomain: bruteResult.Subdomain,
			IP:        bruteResult.IP,
			Status:    bruteResult.Status,
			Title:     bruteResult.Title,
			Port:      bruteResult.Port,
			Alive:     bruteResult.Alive,
			Source:    bruteResult.Source,
			Time:      bruteResult.Time,
		}
	}

	logger.Infof("Brute force completed, found %d subdomains", len(results))
	return results, nil
}

// FilterAlive 过滤存活子域
func (c *Collector) FilterAlive(subdomains []Subdomain) []Subdomain {
	alive := make([]Subdomain, 0)
	for _, subdomain := range subdomains {
		if subdomain.Alive {
			alive = append(alive, subdomain)
		}
	}
	return alive
}

// collectFromDNSReflect 从 DNS 反射查询收集
func (c *Collector) collectFromDNSReflect() ([]Subdomain, error) {
	logger.Info("Starting DNS reflection collection")

	subdomainStrings, err := c.reflectClient.QueryReflect(c.domain)
	if err != nil {
		return nil, err
	}

	subdomains := make([]Subdomain, len(subdomainStrings))
	for i, subdomain := range subdomainStrings {
		subdomains[i] = Subdomain{
			Subdomain: subdomain,
			Source:    "dns_reflect",
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return subdomains, nil
}

// collectFromCertificates 从证书透明度日志收集
func (c *Collector) collectFromCertificates() ([]Subdomain, error) {
	logger.Info("Starting certificate transparency collection")

	subdomainStrings, err := c.certClient.QueryCertificates(c.domain)
	if err != nil {
		return nil, err
	}

	subdomains := make([]Subdomain, len(subdomainStrings))
	for i, subdomain := range subdomainStrings {
		subdomains[i] = Subdomain{
			Subdomain: subdomain,
			Source:    "certificates",
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return subdomains, nil
}

// collectFromOSINT 从 OSINT API 收集
func (c *Collector) collectFromOSINT() ([]Subdomain, error) {
	logger.Info("Starting OSINT API collection")

	subdomainStrings, err := c.osintClient.QueryOSINT(c.domain)
	if err != nil {
		return nil, err
	}

	subdomains := make([]Subdomain, len(subdomainStrings))
	for i, subdomain := range subdomainStrings {
		subdomains[i] = Subdomain{
			Subdomain: subdomain,
			Source:    "osint_api",
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return subdomains, nil
}

// collectFromSearchEngines 从搜索引擎收集
func (c *Collector) collectFromSearchEngines() ([]Subdomain, error) {
	logger.Info("Starting search engine collection")

	subdomainStrings, err := c.searchClient.QuerySearchEngines(c.domain)
	if err != nil {
		return nil, err
	}

	subdomains := make([]Subdomain, len(subdomainStrings))
	for i, subdomain := range subdomainStrings {
		subdomains[i] = Subdomain{
			Subdomain: subdomain,
			Source:    "search_engines",
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return subdomains, nil
}

// collectFromModules 从模块化系统收集
func (c *Collector) collectFromModules() ([]Subdomain, error) {
	logger.Info("Starting module-based collection")

	// 运行所有模块
	allResults, err := c.moduleManager.RunAllModules(c.domain)
	if err != nil {
		return nil, err
	}

	var allSubdomains []string
	for moduleType, results := range allResults {
		logger.Infof("Module type %s found %d subdomains", moduleType, len(results))
		allSubdomains = append(allSubdomains, results...)
	}

	// 转换为 Subdomain 结构
	subdomains := make([]Subdomain, len(allSubdomains))
	for i, subdomain := range allSubdomains {
		subdomains[i] = Subdomain{
			Subdomain: subdomain,
			Source:    "modules",
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return subdomains, nil
}

// collectFromSearch 从搜索引擎收集（传统方式）
func (c *Collector) collectFromSearch() ([]Subdomain, error) {
	// 实现从搜索引擎收集子域的逻辑
	// 这里简化实现
	return []Subdomain{}, nil
}

// collectFromDNS 从 DNS 记录收集
func (c *Collector) collectFromDNS() ([]Subdomain, error) {
	// 实现从 DNS 记录收集子域的逻辑
	// 这里简化实现
	return []Subdomain{}, nil
}

// collectFromDatasets 从在线数据集收集
func (c *Collector) collectFromDatasets() ([]Subdomain, error) {
	// 实现从在线数据集收集子域的逻辑
	// 这里简化实现
	return []Subdomain{}, nil
}

// deduplicate 去重
func (c *Collector) deduplicate(subdomains []Subdomain) []Subdomain {
	seen := make(map[string]bool)
	result := make([]Subdomain, 0)

	for _, subdomain := range subdomains {
		if !seen[subdomain.Subdomain] {
			seen[subdomain.Subdomain] = true
			result = append(result, subdomain)
		}
	}

	return result
}

// GetDomain 获取域名
func (c *Collector) GetDomain() string {
	return c.domain
}

// GetResults 获取结果
func (c *Collector) GetResults() []Subdomain {
	return c.results
}

// SetResults 设置结果
func (c *Collector) SetResults(results []Subdomain) {
	c.results = results
}

// RegisterModule 注册模块
func (c *Collector) RegisterModule(module modules.Module) {
	c.moduleManager.RegisterModule(module)
}
