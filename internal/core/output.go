package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/validator"
	"github.com/oneforall-go/pkg/logger"
)

// SubdomainResult 子域名结果
type SubdomainResult struct {
	Subdomain   string   `json:"subdomain"`
	IP          []string `json:"ip"`
	Status      int      `json:"status"`
	Title       string   `json:"title"`
	Port        int      `json:"port"`
	Alive       bool     `json:"alive"`
	Source      string   `json:"source"`
	Time        string   `json:"time"`
	Provider    string   `json:"provider"`
	DNSResolved bool     `json:"dns_resolved"`
	PingAlive   bool     `json:"ping_alive"`
	StatusCode  int      `json:"status_code"`
	StatusText  string   `json:"status_text"`
}

// OutputManager 输出管理器
type OutputManager struct {
	config     *config.Config
	results    []SubdomainResult
	outputPath string
	format     string
}

// NewOutputManager 创建输出管理器
func NewOutputManager(cfg *config.Config) *OutputManager {
	return &OutputManager{
		config:  cfg,
		results: make([]SubdomainResult, 0),
		format:  cfg.ResultSaveFormat,
	}
}

// AddResult 添加结果
func (o *OutputManager) AddResult(result SubdomainResult) {
	o.results = append(o.results, result)
}

// AddResults 添加多个结果
func (o *OutputManager) AddResults(results []SubdomainResult) {
	o.results = append(o.results, results...)
}

// AddValidationResults 添加验证结果
func (o *OutputManager) AddValidationResults(results []validator.ValidationResult) {
	for _, result := range results {
		// 添加所有验证结果，包括验证不通过的域名
		o.results = append(o.results, SubdomainResult{
			Subdomain:   result.Subdomain,
			IP:          result.IP,
			Status:      result.Status,
			Title:       result.Title,
			Port:        result.Port,
			Alive:       result.Alive,
			Source:      result.Source,
			Time:        result.Time,
			Provider:    result.Provider,
			DNSResolved: result.DNSResolved,
			PingAlive:   result.PingAlive,
			StatusCode:  result.StatusCode,
			StatusText:  result.StatusText,
		})
	}
}

// SetOutputPath 设置输出路径
func (o *OutputManager) SetOutputPath(path string) {
	o.outputPath = path
}

// SetFormat 设置输出格式
func (o *OutputManager) SetFormat(format string) {
	o.format = format
}

// GetResults 获取所有结果
func (o *OutputManager) GetResults() []SubdomainResult {
	return o.results
}

// FilterAlive 过滤存活结果
func (o *OutputManager) FilterAlive() []SubdomainResult {
	var aliveResults []SubdomainResult
	for _, result := range o.results {
		if result.Alive {
			aliveResults = append(aliveResults, result)
		}
	}
	return aliveResults
}

// Deduplicate 去重
func (o *OutputManager) Deduplicate() {
	seen := make(map[string]bool)
	var uniqueResults []SubdomainResult

	for _, result := range o.results {
		if !seen[result.Subdomain] {
			seen[result.Subdomain] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	o.results = uniqueResults
}

// Export 导出结果
func (o *OutputManager) Export() error {
	if len(o.results) == 0 {
		logger.Warn("No results to export")
		return nil
	}

	// 去重
	o.Deduplicate()

	// 根据配置过滤存活结果
	if o.config.ResultExportAlive {
		o.results = o.FilterAlive()
	}

	// 生成输出路径
	if o.outputPath == "" {
		o.outputPath = o.generateOutputPath()
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(o.outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 根据格式导出
	switch o.format {
	case "csv":
		return o.exportCSV()
	case "json":
		return o.exportJSON()
	default:
		return fmt.Errorf("unsupported format: %s", o.format)
	}
}

// exportCSV 导出为 CSV
func (o *OutputManager) exportCSV() error {
	file, err := os.Create(o.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"subdomain", "ip", "status", "title", "port", "alive", "source", "time", "provider", "dns_resolved", "ping_alive", "status_code", "status_text"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// 写入数据
	for _, result := range o.results {
		row := []string{
			result.Subdomain,
			strings.Join(result.IP, ","),
			fmt.Sprintf("%d", result.Status),
			result.Title,
			fmt.Sprintf("%d", result.Port),
			fmt.Sprintf("%t", result.Alive),
			result.Source,
			result.Time,
			result.Provider,
			fmt.Sprintf("%t", result.DNSResolved),
			fmt.Sprintf("%t", result.PingAlive),
			fmt.Sprintf("%d", result.StatusCode),
			result.StatusText,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	logger.Infof("Exported %d results to CSV: %s", len(o.results), o.outputPath)
	return nil
}

// exportJSON 导出为 JSON
func (o *OutputManager) exportJSON() error {
	file, err := os.Create(o.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(o.results); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	logger.Infof("Exported %d results to JSON: %s", len(o.results), o.outputPath)
	return nil
}

// generateOutputPath 生成输出路径
func (o *OutputManager) generateOutputPath() string {
	timestamp := time.Now().Format("20060102_150405")
	domain := "unknown"
	if len(o.results) > 0 {
		// 从第一个结果中提取域名
		subdomain := o.results[0].Subdomain
		if parts := strings.Split(subdomain, "."); len(parts) >= 2 {
			domain = strings.Join(parts[len(parts)-2:], ".")
		}
	}

	filename := fmt.Sprintf("%s_%s.%s", domain, timestamp, o.format)
	return filepath.Join(o.config.ResultSavePath, filename)
}

// GetOutputPath 获取输出路径
func (o *OutputManager) GetOutputPath() string {
	return o.outputPath
}

// GetStats 获取统计信息
func (o *OutputManager) GetStats() map[string]interface{} {
	total := len(o.results)
	alive := 0
	sources := make(map[string]int)
	providers := make(map[string]int)

	for _, result := range o.results {
		if result.Alive {
			alive++
		}
		sources[result.Source]++
		if result.Provider != "" {
			providers[result.Provider]++
		}
	}

	return map[string]interface{}{
		"total":     total,
		"alive":     alive,
		"dead":      total - alive,
		"sources":   sources,
		"providers": providers,
	}
}
