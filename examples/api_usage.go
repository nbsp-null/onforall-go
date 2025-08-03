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

	// 运行子域名枚举
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

	// 输出JSON格式
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Printf("\n=== JSON Output ===\n")
		fmt.Println(string(jsonData))
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
		Verbose:                   false,
		OutputFormat:              "json",
		OutputPath:                "./custom_results",
	}

	result, err := oneforallAPI.RunSubdomainEnumeration(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Custom config result: %d subdomains found\n", result.TotalSubdomains)
}
