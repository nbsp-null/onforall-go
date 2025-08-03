# OneForAll-Go API

OneForAll-Go API 提供了一个简单易用的接口，方便外部包引用和调用子域名枚举功能。**API仅返回数据结构，不保存到本地文件**，适合作为库被其他项目引用。

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
    
    // 运行子域名枚举（仅返回数据结构，不保存到本地）
    result, err := oneforallAPI.RunSubdomainEnumeration(options)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // 处理返回的数据结构
    fmt.Printf("Found %d subdomains (%d alive)\n", 
        result.TotalSubdomains, result.AliveSubdomains)
    
    // 进一步处理结果
    for _, subdomain := range result.Results {
        fmt.Printf("Subdomain: %s, Alive: %v\n", subdomain.Subdomain, subdomain.Alive)
    }
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
    Target: "example.com",
    EnableValidation: true,
    EnableBruteForce: false,
    Concurrency: 10,
    Timeout: 60 * time.Second,
    EnableSearchModules: true,
    EnableDatasetModules: true,
    EnableCertificateModules: true,
    EnableCrawlModules: false, // 禁用爬虫模块
    EnableCheckModules: false, // 禁用检查模块
    EnableIntelligenceModules: false, // 禁用智能模块
    EnableEnrichModules: true,
    Debug: true,
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
    
    // 处理结果，例如保存到数据库
    saveToDatabase(result)
}
```

### 5. 结果处理示例

```go
result, err := oneforallAPI.RunSubdomainEnumeration(options)
if err != nil {
    log.Fatal(err)
}

// 按存活状态分组
var aliveSubdomains []string
var deadSubdomains []string

for _, subdomain := range result.Results {
    if subdomain.Alive {
        aliveSubdomains = append(aliveSubdomains, subdomain.Subdomain)
    } else {
        deadSubdomains = append(deadSubdomains, subdomain.Subdomain)
    }
}

// 按来源模块分组
sourceMap := make(map[string][]string)
for _, subdomain := range result.Results {
    sourceMap[subdomain.Source] = append(sourceMap[subdomain.Source], subdomain.Subdomain)
}

// 转换为JSON发送到其他服务
jsonData, err := json.Marshal(result)
if err != nil {
    log.Fatal(err)
}

// 发送到外部API
sendToExternalAPI(jsonData)
```

### 6. 发送到外部服务

```go
func sendToExternalAPI(result *api.Result) {
    // 将结果发送到外部服务
    jsonData, err := json.Marshal(result)
    if err != nil {
        log.Printf("Error marshaling: %v", err)
        return
    }
    
    // 发送HTTP请求
    resp, err := http.Post("https://api.example.com/subdomains", 
        "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Error sending to external API: %v", err)
        return
    }
    defer resp.Body.Close()
    
    log.Printf("Successfully sent results to external API")
}
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

## 重要特性

### ✅ 已实现功能

1. **纯数据结构返回**：不保存到本地文件，只返回内存中的数据结构
2. **完整的API封装**：提供简单易用的接口
3. **参数化配置**：支持所有主要功能的开关和参数调整
4. **错误处理**：完善的错误处理和返回机制
5. **JSON支持**：所有结构体都支持JSON序列化/反序列化
6. **默认配置**：提供合理的默认配置选项
7. **模块化设计**：支持按需启用/禁用不同模块
8. **性能优化**：支持并发控制和超时设置
9. **详细结果**：返回完整的子域名信息和状态
10. **统计信息**：提供总数量、存活数量、存活百分比等统计
11. **执行时间**：记录执行时间用于性能分析

### 🔧 技术实现

1. **类型安全**：使用强类型结构体确保类型安全
2. **空值处理**：正确处理nil值和空结果
3. **参数验证**：验证必需参数和设置默认值
4. **日志集成**：集成现有的日志系统
5. **配置继承**：继承现有的配置系统
6. **模块注册**：支持动态模块注册和配置
7. **无文件依赖**：不依赖本地文件系统，适合容器化部署

## 性能优化建议

1. **批量处理时禁用爆破模块**：提高处理速度
2. **调整并发数**：根据网络和系统性能调整
3. **设置合理的超时时间**：避免长时间等待
4. **禁用不需要的模块**：减少不必要的请求
5. **使用适当的日志级别**：生产环境使用 warn 级别
6. **内存管理**：处理大量结果时注意内存使用

## 注意事项

1. **网络连接**：确保有稳定的网络连接
2. **API限制**：某些模块可能受到API调用限制
3. **资源使用**：高并发可能消耗较多系统资源
4. **法律合规**：确保在合法范围内使用
5. **目标域名**：确保有权限对目标域名进行枚举
6. **内存使用**：大量结果可能占用较多内存
7. **无本地存储**：API不会保存任何文件到本地

## 依赖项

确保项目中包含以下依赖：

```go
go get github.com/oneforall-go/pkg/api
```

## 许可证

本项目遵循相应的开源许可证。

## 与命令行版本的区别

| 特性 | 命令行版本 | API版本 |
|------|------------|---------|
| 本地文件保存 | ✅ 保存到本地文件 | ❌ 仅返回数据结构 |
| 外部包引用 | ❌ 不适合 | ✅ 专为库使用设计 |
| 参数传递 | 命令行参数 | 结构体参数 |
| 结果处理 | 固定格式输出 | 灵活的数据结构 |
| 集成性 | 独立工具 | 可嵌入其他项目 |
| 容器化 | 需要文件系统 | 无文件依赖 | 