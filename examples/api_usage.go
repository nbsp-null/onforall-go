package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/oneforall-go/pkg/api"
)

func main() {
	// 创建API实例
	oneforallAPI := api.NewOneForAllAPI()

	// 获取默认配置
	options := api.GetDefaultOptions()

	// 设置目标域名
	options.Target = "ex.cn"

	// 自定义配置
	options.EnableValidation = true
	options.EnableBruteForce = true
	options.Concurrency = 5
	options.Timeout = 30 * time.Second
	options.Debug = true

	// 运行子域名枚举（仅返回数据结构，不保存到本地）
	result, err := oneforallAPI.RunSubdomainEnumeration(options)
	if err != nil {
		log.Fatalf("Error running subdomain enumeration: %v", err)
	}

	// 打印结果
	fmt.Printf("=== Subdomain Enumeration Results ===\n")
	fmt.Printf("Domain: %s\n", result.Domain)
	fmt.Printf("Total Subdomains: %d\n", result.TotalSubdomains)
	fmt.Printf("Alive Subdomains: %d\n", result.AliveSubdomains)
	fmt.Printf("Alive Percentage: %.2f%%\n", result.AlivePercentage)
	fmt.Printf("Execution Time: %v\n", result.ExecutionTime)

	// 打印详细结果
	fmt.Printf("\n=== Detailed Results ===\n")
	for i, subdomain := range result.Results {
		fmt.Printf("%d. %s (Source: %s, Alive: %v)\n",
			i+1, subdomain.Subdomain, subdomain.Source, subdomain.Alive)
		if len(subdomain.IP) > 0 {
			fmt.Printf("   IP: %v\n", subdomain.IP)
		}
		if subdomain.Provider != "" {
			fmt.Printf("   Provider: %s\n", subdomain.Provider)
		}
	}

	// 输出JSON格式（供外部包处理）
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Printf("\n=== JSON Output (for external processing) ===\n")
		fmt.Println(string(jsonData))
	}

	// 演示如何进一步处理结果
	processResults(result)
}

// processResults 演示如何进一步处理结果
func processResults(result *api.Result) {
	fmt.Printf("\n=== Result Processing Demo ===\n")

	// 1. 按存活状态分组
	var aliveSubdomains []string
	var deadSubdomains []string

	for _, subdomain := range result.Results {
		if subdomain.Alive {
			aliveSubdomains = append(aliveSubdomains, subdomain.Subdomain)
		} else {
			deadSubdomains = append(deadSubdomains, subdomain.Subdomain)
		}
	}

	fmt.Printf("Alive subdomains (%d): %v\n", len(aliveSubdomains), aliveSubdomains)
	fmt.Printf("Dead subdomains (%d): %v\n", len(deadSubdomains), deadSubdomains)

	// 2. 按来源模块分组
	sourceMap := make(map[string][]string)
	for _, subdomain := range result.Results {
		sourceMap[subdomain.Source] = append(sourceMap[subdomain.Source], subdomain.Subdomain)
	}

	fmt.Printf("\nSubdomains by source:\n")
	for source, subdomains := range sourceMap {
		fmt.Printf("  %s: %d subdomains\n", source, len(subdomains))
	}

	// 3. 统计IP提供商
	providerMap := make(map[string]int)
	for _, subdomain := range result.Results {
		if subdomain.Provider != "" {
			providerMap[subdomain.Provider]++
		}
	}

	fmt.Printf("\nIP providers:\n")
	for provider, count := range providerMap {
		fmt.Printf("  %s: %d subdomains\n", provider, count)
	}
}

// 示例：批量处理多个域名
func batchProcessExample() {
	oneforallAPI := api.NewOneForAllAPI()

	domains := []string{"ex.cn", "example.com", "test.org"}

	for _, domain := range domains {
		options := api.GetDefaultOptions()
		options.Target = domain
		options.EnableValidation = true
		options.EnableBruteForce = false // 批量处理时禁用爆破以提高速度
		options.Concurrency = 3
		options.Timeout = 20 * time.Second

		fmt.Printf("Processing domain: %s\n", domain)
		result, err := oneforallAPI.RunSubdomainEnumeration(options)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", domain, err)
			continue
		}

		fmt.Printf("Found %d subdomains for %s (%d alive)\n",
			result.TotalSubdomains, domain, result.AliveSubdomains)

		// 这里可以进一步处理结果，比如保存到数据库、发送到其他服务等
		// 而不是保存到本地文件
	}
}

// 示例：自定义配置
func customConfigExample() {
	oneforallAPI := api.NewOneForAllAPI()

	options := api.Options{
		Target:                    "ex.cn",
		EnableValidation:          true,
		EnableBruteForce:          true,
		Concurrency:               10,
		Timeout:                   60 * time.Second,
		EnableSearchModules:       true,
		EnableDatasetModules:      true,
		EnableCertificateModules:  true,
		EnableCrawlModules:        false, // 禁用爬虫模块
		EnableCheckModules:        false, // 禁用检查模块
		EnableIntelligenceModules: false, // 禁用智能模块
		EnableEnrichModules:       true,
		Debug:                     true,
	}

	result, err := oneforallAPI.RunSubdomainEnumeration(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Custom config result: %d subdomains found\n", result.TotalSubdomains)

	// 将结果发送到其他服务或保存到数据库
	sendToExternalService(result)
}

// sendToExternalService 演示如何将结果发送到外部服务
func sendToExternalService(result *api.Result) {
	// 这里可以添加将结果发送到外部服务的逻辑
	// 例如：发送到API、保存到数据库、写入消息队列等

	fmt.Printf("Sending results to external service...\n")
	fmt.Printf("Domain: %s\n", result.Domain)
	fmt.Printf("Total subdomains: %d\n", result.TotalSubdomains)
	fmt.Printf("Alive subdomains: %d\n", result.AliveSubdomains)

	// 示例：将结果转换为JSON并发送
	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling for external service: %v", err)
		return
	}

	fmt.Printf("JSON data length: %d bytes\n", len(jsonData))
	// 这里可以添加实际的HTTP请求或其他发送逻辑
}
