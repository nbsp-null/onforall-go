package certificates

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// CRTSh CRTSh 证书模块
type CRTSh struct {
	*core.Query
	baseURL string
}

// CRTShRecord CRTSh 记录结构
type CRTShRecord struct {
	NameValue string `json:"name_value"`
}

// NewCRTSh 创建 CRTSh 证书模块
func NewCRTSh(cfg *config.Config) *CRTSh {
	return &CRTSh{
		Query:   core.NewQuery("CrtshQuery", cfg),
		baseURL: "https://crt.sh/",
	}
}

// Run 执行查询
func (c *CRTSh) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 执行查询
	if err := c.query(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// query 执行查询
func (c *CRTSh) query(domain string) error {
	// 设置请求头
	c.SetHeader("User-Agent", c.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("q", fmt.Sprintf("%%.%s", domain))
	params.Set("output", "json")

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	resp, err := c.HTTPGet(queryURL, c.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query CRTSh: %v", err)
	}

	// 读取响应
	body, err := c.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 处理响应文本
	text := strings.ReplaceAll(body, "\\n", " ")

	// 解析 JSON 数据
	var jsonData []CRTShRecord
	if err := json.Unmarshal([]byte(text), &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 处理通配符域名
	subDomains := make(map[string]bool)
	for _, record := range jsonData {
		nameValue := record.NameValue
		if strings.Contains(nameValue, "*") {
			// 读取 altdns 字典文件
			altdnsWords, err := c.readAltDNSWords()
			if err == nil {
				for _, word := range altdnsWords {
					result := strings.Replace(nameValue, "*", word, 1)
					if strings.Contains(result, domain) {
						subDomains[result] = true
					}
				}
			}
		}
	}

	// 将处理后的通配符域名添加到文本中
	for subdomain := range subDomains {
		text += "," + subdomain + ","
	}

	// 提取子域名
	subdomains := c.ExtractSubdomains(text, domain)
	for _, subdomain := range subdomains {
		c.AddSubdomain(subdomain)
	}

	return nil
}

// readAltDNSWords 读取 altdns 字典文件
func (c *CRTSh) readAltDNSWords() ([]string, error) {
	// 尝试读取字典文件
	dictPath := "data/altdns_wordlist.txt"
	content, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, err
	}

	// 按行分割
	lines := strings.Split(string(content), "\n")
	var words []string
	for _, line := range lines {
		word := strings.TrimSpace(line)
		if word != "" {
			words = append(words, word)
		}
	}

	return words, nil
}
