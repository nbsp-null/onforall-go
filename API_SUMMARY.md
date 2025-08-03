# OneForAll-Go API 封装接口总结

## 概述

我已经成功为OneForAll-Go创建了一个完整的API封装接口，支持通过参数传递调用`runLib`的整个功能，方便外部包引用调用。

## 主要功能

### 1. API接口设计

#### 核心结构体

```go
// Options - 配置选项
type Options struct {
    Target string // 目标域名（必需）
    EnableValidation bool // 是否启用域名验证
    EnableBruteForce bool // 是否启用爆破攻击
    Concurrency int // 并发数
    Timeout time.Duration // 超时时间
    // ... 其他模块开关和配置
}

// Result - 执行结果
type Result struct {
    Domain string // 目标域名
    TotalSubdomains int // 总子域名数
    AliveSubdomains int // 存活子域名数
    AlivePercentage float64 // 存活百分比
    Results []SubdomainResult // 详细结果
    ExecutionTime time.Duration // 执行时间
    Error string // 错误信息
}

// SubdomainResult - 子域名结果
type SubdomainResult struct {
    Subdomain string // 子域名
    Source string // 来源模块
    Time string // 发现时间
    Alive bool // 是否存活
    IP []string // IP地址列表
    DNSResolved bool // DNS解析状态
    PingAlive bool // Ping存活状态
    StatusCode int // 状态码
    StatusText string // 状态文本
    Provider string // IP提供商
}
```

#### 核心方法

```go
// NewOneForAllAPI - 创建API实例
func NewOneForAllAPI() *OneForAllAPI

// RunSubdomainEnumeration - 运行子域名枚举
func (api *OneForAllAPI) RunSubdomainEnumeration(options Options) (*Result, error)

// GetDefaultOptions - 获取默认配置
func GetDefaultOptions() Options
```

### 2. 使用示例

#### 基本用法

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

#### 批量处理

```go
domains := []string{"ex.cn", "example.com", "test.org"}

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

#### 自定义配置

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
    OutputFormat: "json",
    OutputPath: "./results",
}

result, err := oneforallAPI.RunSubdomainEnumeration(options)
```

### 3. 主要特性

#### ✅ 已实现功能

1. **完整的API封装**：提供简单易用的接口
2. **参数化配置**：支持所有主要功能的开关和参数调整
3. **错误处理**：完善的错误处理和返回机制
4. **JSON支持**：所有结构体都支持JSON序列化/反序列化
5. **默认配置**：提供合理的默认配置选项
6. **模块化设计**：支持按需启用/禁用不同模块
7. **性能优化**：支持并发控制和超时设置
8. **详细结果**：返回完整的子域名信息和状态
9. **统计信息**：提供总数量、存活数量、存活百分比等统计
10. **执行时间**：记录执行时间用于性能分析

#### 🔧 技术实现

1. **类型安全**：使用强类型结构体确保类型安全
2. **空值处理**：正确处理nil值和空结果
3. **参数验证**：验证必需参数和设置默认值
4. **日志集成**：集成现有的日志系统
5. **配置继承**：继承现有的配置系统
6. **模块注册**：支持动态模块注册和配置

### 4. 文件结构

```
oneforall-go/
├── pkg/api/
│   ├── api.go          # 主要API实现
│   ├── api_test.go     # 单元测试
│   └── README.md       # 使用文档
├── examples/
│   └── api_usage.go    # 使用示例
└── API_SUMMARY.md      # 本文档
```

### 5. 测试覆盖

- ✅ `TestGetDefaultOptions` - 测试默认配置
- ✅ `TestNewOneForAllAPI` - 测试API实例创建
- ✅ `TestRunSubdomainEnumeration_EmptyTarget` - 测试空目标错误处理
- ✅ `TestSubdomainResult_JSON` - 测试JSON序列化
- ✅ `TestResult_JSON` - 测试结果JSON序列化

### 6. 使用优势

1. **简单易用**：只需几行代码即可使用
2. **高度可配置**：支持所有主要功能的开关和参数调整
3. **类型安全**：使用Go的强类型系统确保安全
4. **错误友好**：提供详细的错误信息和处理机制
5. **性能可控**：支持并发控制和超时设置
6. **结果丰富**：返回详细的子域名信息和统计
7. **易于集成**：可以作为库被其他项目引用

### 7. 外部包引用

其他项目可以通过以下方式引用：

```go
import "github.com/oneforall-go/pkg/api"

// 使用API
api := api.NewOneForAllAPI()
options := api.GetDefaultOptions()
options.Target = "example.com"
result, err := api.RunSubdomainEnumeration(options)
```

## 总结

这个API封装接口提供了：

1. **完整的封装**：将复杂的`runLib`功能封装成简单易用的API
2. **参数化调用**：支持通过参数传递控制所有功能
3. **数据结构返回**：返回结构化的数据而不是简单的字符串数组
4. **外部包支持**：方便其他项目引用和集成
5. **完善的文档**：提供详细的使用说明和示例

这个API接口使得OneForAll-Go可以作为一个强大的子域名枚举库被其他项目使用，同时保持了原有的所有功能。 