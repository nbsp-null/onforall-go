# OneForAll-Go API

OneForAll-Go API 提供了一个简单易用的接口，方便外部包引用和调用子域名枚举功能。

## 快速开始

### 基本用法

```go
package main

import (
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
    
    // 运行子域名枚举
    result, err := oneforallAPI.RunSubdomainEnumeration(options)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // 打印结果
    fmt.Printf("Found %d subdomains (%d alive)\n", 
        result.TotalSubdomains, result.AliveSubdomains)
}
```

### 配置选项

#### Options 结构体

```go
type Options struct {
    // 基本配置
    Target string // 目标域名（必需）
    
    // 功能开关
    EnableValidation bool // 是否启用域名验证
    EnableBruteForce bool // 是否启用爆破攻击
    
    // 性能配置
    Concurrency int           // 并发数
    Timeout     time.Duration // 超时时间
    
    // 模块开关
    EnableSearchModules     bool // 搜索模块
    EnableDatasetModules    bool // 数据集模块
    EnableCertificateModules bool // 证书模块
    EnableCrawlModules      bool // 爬虫模块
    EnableCheckModules      bool // 检查模块
    EnableIntelligenceModules bool // 智能模块
    EnableEnrichModules     bool // 丰富模块
    
    // 日志配置
    Debug   bool // 调试模式
    Verbose bool // 详细日志
    
    // 输出配置
    OutputFormat string // 输出格式 (csv/json/txt)
    OutputPath   string // 输出路径
}
```

### 结果结构

#### Result 结构体

```go
type Result struct {
    Domain           string             // 目标域名
    TotalSubdomains  int                // 总子域名数
    AliveSubdomains  int                // 存活子域名数
    AlivePercentage  float64            // 存活百分比
    Results          []SubdomainResult  // 详细结果
    ExecutionTime    time.Duration      // 执行时间
    Error            string             // 错误信息
}
```

#### SubdomainResult 结构体

```go
type SubdomainResult struct {
    Subdomain    string   // 子域名
    Source       string   // 来源模块
    Time         string   // 发现时间
    Alive        bool     // 是否存活
    IP           []string // IP地址列表
    DNSResolved  bool     // DNS解析状态
    PingAlive    bool     // Ping存活状态
    StatusCode   int      // 状态码
    StatusText   string   // 状态文本
    Provider     string   // IP提供商
}
```

## 使用示例

### 1. 基本枚举

```go
options := api.GetDefaultOptions()
options.Target = "example.com"
options.EnableValidation = true

result, err := oneforallAPI.RunSubdomainEnumeration(options)
```

### 2. 高性能枚举（禁用验证）

```go
options := api.GetDefaultOptions()
options.Target = "example.com"
options.EnableValidation = false
options.EnableBruteForce = true
options.Concurrency = 20
options.Timeout = 120 * time.Second

result, err := oneforallAPI.RunSubdomainEnumeration(options)
```

### 3. 自定义模块配置

```go
options := api.Options{
    Target:                 "example.com",
    EnableValidation:       true,
    EnableBruteForce:       false,
    Concurrency:            10,
    Timeout:                60 * time.Second,
    EnableSearchModules:     true,
    EnableDatasetModules:    true,
    EnableCertificateModules: true,
    EnableCrawlModules:      false, // 禁用爬虫模块
    EnableCheckModules:      false, // 禁用检查模块
    EnableIntelligenceModules: false, // 禁用智能模块
    EnableEnrichModules:     true,
    Debug:                  true,
    OutputFormat:           "json",
    OutputPath:             "./results",
}

result, err := oneforallAPI.RunSubdomainEnumeration(options)
```

### 4. 批量处理

```go
domains := []string{"example.com", "test.org", "demo.net"}

for _, domain := range domains {
    options := api.GetDefaultOptions()
    options.Target = domain
    options.EnableBruteForce = false // 批量处理时禁用爆破
    
    result, err := oneforallAPI.RunSubdomainEnumeration(options)
    if err != nil {
        log.Printf("Error processing %s: %v", domain, err)
        continue
    }
    
    fmt.Printf("Domain: %s, Found: %d, Alive: %d\n", 
        domain, result.TotalSubdomains, result.AliveSubdomains)
}
```

### 5. 获取JSON输出

```go
result, err := oneforallAPI.RunSubdomainEnumeration(options)
if err != nil {
    log.Fatal(err)
}

// 输出JSON格式
jsonData, err := json.MarshalIndent(result, "", "  ")
if err != nil {
    log.Fatal(err)
}

fmt.Println(string(jsonData))
```

## 默认配置

使用 `api.GetDefaultOptions()` 获取默认配置：

```go
options := api.GetDefaultOptions()
// 默认配置：
// - EnableValidation: true
// - EnableBruteForce: false
// - Concurrency: 10
// - Timeout: 60s
// - 所有模块都启用
// - Debug: false
// - OutputFormat: "json"
// - OutputPath: "./results"
```

## 错误处理

```go
result, err := oneforallAPI.RunSubdomainEnumeration(options)
if err != nil {
    // 处理错误
    log.Printf("Enumeration failed: %v", err)
    return
}

if result.Error != "" {
    // 检查结果中的错误信息
    log.Printf("Enumeration completed with errors: %s", result.Error)
}
```

## 性能优化建议

1. **批量处理时禁用爆破模块**：提高处理速度
2. **调整并发数**：根据网络和系统性能调整
3. **设置合理的超时时间**：避免长时间等待
4. **禁用不需要的模块**：减少不必要的请求
5. **使用适当的日志级别**：生产环境使用 warn 级别

## 注意事项

1. **网络连接**：确保有稳定的网络连接
2. **API限制**：某些模块可能受到API调用限制
3. **资源使用**：高并发可能消耗较多系统资源
4. **法律合规**：确保在合法范围内使用
5. **目标域名**：确保有权限对目标域名进行枚举

## 依赖项

确保项目中包含以下依赖：

```go
go get github.com/oneforall-go/pkg/api
```

## 许可证

本项目遵循相应的开源许可证。 