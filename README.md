# OneForAll-Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://golang.org/)
[![License](https://img.shields.io/github/license/oneforall-go/oneforall-go)](LICENSE)
[![Release](https://img.shields.io/badge/release-v1.0.0-brightgreen)](https://github.com/oneforall-go/oneforall-go/releases)

👊 **OneForAll-Go 是一款功能强大的子域收集工具（Go 语言重构版）**

这是原 Python 版本 OneForAll 的 Go 语言重构版本，保持了原有的功能特性，同时提供了更好的性能和跨平台支持。

## 🚀 特性

- 🔍 **强大的子域收集能力** - 支持多种收集方式
- 🌐 **DNS 解析** - 高效的 DNS 查询和解析
- 🔗 **HTTP 请求** - 快速验证子域存活状态
- 🎯 **暴力破解** - 支持子域暴力破解
- 📊 **结果导出** - 支持 CSV、JSON 等多种格式
- 🎨 **现代化界面** - 清晰的命令行界面
- ⚡ **高性能** - Go 语言原生性能优势

## 📋 安装要求

- Go 1.21 或更高版本
- 网络连接（用于在线收集）

## 🛠️ 安装步骤

### 1. 克隆项目

```bash
git clone https://github.com/oneforall-go/oneforall-go.git
cd oneforall-go
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 编译

```bash
go build -o oneforall-go cmd/main.go
```

### 4. 运行

```bash
./oneforall-go --help
```

## 📖 使用指南

### 基本用法

```bash
# 收集单个域名的子域
./oneforall-go --target example.com run

# 从文件读取多个域名
./oneforall-go --targets domains.txt run

# 指定输出格式
./oneforall-go --target example.com --format csv run

# 只导出存活的子域
./oneforall-go --target example.com --alive run
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--target` | 目标域名 | - |
| `--targets` | 域名文件路径 | - |
| `--brute` | 启用暴力破解 | true |
| `--dns` | 启用 DNS 解析 | true |
| `--request` | 启用 HTTP 请求 | true |
| `--alive` | 只导出存活子域 | false |
| `--format` | 输出格式 (csv/json) | csv |
| `--output` | 输出文件路径 | - |

### 示例

```bash
# 收集 example.com 的子域
./oneforall-go --target example.com run

# 从文件读取域名列表
./oneforall-go --targets domains.txt run

# 只导出存活的子域，格式为 JSON
./oneforall-go --target example.com --alive --format json run

# 禁用暴力破解模块
./oneforall-go --target example.com --brute=false run
```

## 📁 项目结构

```
oneforall-go/
├── cmd/                    # 命令行入口
│   └── main.go
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── collector/         # 子域收集器
│   ├── dns/              # DNS 解析
│   ├── http/             # HTTP 请求
│   ├── brute/            # 暴力破解
│   └── export/           # 结果导出
├── pkg/                   # 公共包
│   ├── utils/            # 工具函数
│   └── logger/           # 日志管理
├── data/                  # 数据文件
│   ├── wordlists/        # 字典文件
│   └── config/           # 配置文件
├── results/               # 结果输出
├── go.mod
├── go.sum
└── README.md
```

## 🔧 配置

配置文件位于 `data/config/` 目录下：

- `config.yaml` - 主配置文件
- `api.yaml` - API 密钥配置
- `wordlists/` - 字典文件

## 📊 输出格式

### CSV 格式

```csv
subdomain,ip,status,title,port
www.example.com,93.184.216.34,200,Example Domain,80
api.example.com,93.184.216.35,200,API Server,443
```

### JSON 格式

```json
[
  {
    "subdomain": "www.example.com",
    "ip": "93.184.216.34",
    "status": 200,
    "title": "Example Domain",
    "port": 80
  }
]
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 GNU General Public License v3.0 许可证。

## 🙏 致谢

- 原 Python 版本 OneForAll 项目
- 所有贡献者和用户

## 📞 联系方式

- GitHub Issues: [https://github.com/oneforall-go/oneforall-go/issues](https://github.com/oneforall-go/oneforall-go/issues)
- 邮箱: support@oneforall-go.com

---

**注意**: 本项目仅用于安全研究和授权测试，请遵守相关法律法规。 