package utils

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LoadDomainsFromFile 从文件加载域名列表
func LoadDomainsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" && !strings.HasPrefix(domain, "#") {
			domains = append(domains, domain)
		}
	}

	return domains, scanner.Err()
}

// GetMainDomain 获取主域名
func GetMainDomain(domain string) string {
	// 移除协议前缀
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// 移除路径
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// 移除端口
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// IsValidDomain 验证域名格式
func IsValidDomain(domain string) bool {
	// 简单的域名验证正则表达式
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`
	matched, _ := regexp.MatchString(pattern, domain)
	return matched
}

// CheckInternetConnection 检查网络连接
func CheckInternetConnection() bool {
	// 尝试连接 Google DNS
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 5*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// CheckHTTPConnection 检查 HTTP 连接
func CheckHTTPConnection() bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 尝试访问 Google
	resp, err := client.Get("http://www.google.com")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// CreateDirectory 创建目录
func CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirectoryExists 检查目录是否存在
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	// 读取源文件
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// 写入目标文件
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %v", err)
	}

	return nil
}

// RemoveFile 删除文件
func RemoveFile(path string) error {
	return os.Remove(path)
}

// RemoveDirectory 删除目录
func RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

// ListFiles 列出目录中的文件
func ListFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ListDirectories 列出目录中的子目录
func ListDirectories(dir string) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// GetCurrentTime 获取当前时间字符串
func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetCurrentTimestamp 获取当前时间戳
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// Deduplicate 去重字符串切片
func Deduplicate(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// FilterStrings 过滤字符串切片
func FilterStrings(items []string, predicate func(string) bool) []string {
	result := make([]string, 0)

	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

// MapStrings 映射字符串切片
func MapStrings(items []string, mapper func(string) string) []string {
	result := make([]string, len(items))

	for i, item := range items {
		result[i] = mapper(item)
	}

	return result
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// IndexOfString 获取字符串在切片中的索引
func IndexOfString(items []string, target string) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return -1
}
