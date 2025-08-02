package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oneforall-go/internal/collector"
	"github.com/oneforall-go/pkg/logger"
)

// Exporter 导出器
type Exporter struct {
	format string
	output string
}

// NewExporter 创建新的导出器
func NewExporter(format, output string) *Exporter {
	return &Exporter{
		format: format,
		output: output,
	}
}

// Export 导出结果
func (e *Exporter) Export(subdomains []collector.Subdomain) error {
	switch e.format {
	case "csv":
		return e.exportCSV(subdomains)
	case "json":
		return e.exportJSON(subdomains)
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

// exportCSV 导出为 CSV 格式
func (e *Exporter) exportCSV(subdomains []collector.Subdomain) error {
	// 确定输出文件路径
	outputPath := e.getOutputPath("csv")

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// 创建 CSV 写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"subdomain", "ip", "status", "title", "port", "alive", "source", "time"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// 写入数据
	for _, subdomain := range subdomains {
		// 将 IP 切片转换为字符串
		ipStr := ""
		if len(subdomain.IP) > 0 {
			ipStr = subdomain.IP[0] // 只取第一个 IP
		}

		// 将 alive 布尔值转换为字符串
		aliveStr := "false"
		if subdomain.Alive {
			aliveStr = "true"
		}

		row := []string{
			subdomain.Subdomain,
			ipStr,
			fmt.Sprintf("%d", subdomain.Status),
			subdomain.Title,
			fmt.Sprintf("%d", subdomain.Port),
			aliveStr,
			subdomain.Source,
			subdomain.Time,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	logger.Infof("Exported %d subdomains to %s", len(subdomains), outputPath)
	return nil
}

// exportJSON 导出为 JSON 格式
func (e *Exporter) exportJSON(subdomains []collector.Subdomain) error {
	// 确定输出文件路径
	outputPath := e.getOutputPath("json")

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// 创建 JSON 编码器
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// 编码数据
	if err := encoder.Encode(subdomains); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	logger.Infof("Exported %d subdomains to %s", len(subdomains), outputPath)
	return nil
}

// getOutputPath 获取输出文件路径
func (e *Exporter) getOutputPath(extension string) string {
	if e.output != "" {
		return e.output
	}

	// 生成默认文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("oneforall_results_%s.%s", timestamp, extension)

	return filepath.Join("results", filename)
}

// SetFormat 设置导出格式
func (e *Exporter) SetFormat(format string) {
	e.format = format
}

// GetFormat 获取导出格式
func (e *Exporter) GetFormat() string {
	return e.format
}

// SetOutput 设置输出路径
func (e *Exporter) SetOutput(output string) {
	e.output = output
}

// GetOutput 获取输出路径
func (e *Exporter) GetOutput() string {
	return e.output
}
