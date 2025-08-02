package dnsquery

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// TXT TXT 查询模块
type TXT struct {
	*core.Query
}

// NewTXT 创建 TXT 查询模块
func NewTXT(cfg *config.Config) *TXT {
	return &TXT{
		Query: core.NewQuery("QueryTXT", cfg),
	}
}

// Run 执行查询
func (t *TXT) Run(domain string) ([]string, error) {
	t.SetDomain(domain)
	t.Begin()
	defer t.Finish()

	// 执行 TXT 查询
	if err := t.query(domain); err != nil {
		return nil, err
	}

	return t.GetSubdomains(), nil
}

// query 执行查询
func (t *TXT) query(domain string) error {
	// 创建 DNS 查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	msg.RecursionDesired = true

	// 发送查询
	resp, err := dns.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to query TXT records: %v", err)
	}

	// 处理响应
	for _, answer := range resp.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			for _, record := range txt.Txt {
				// 从 TXT 记录中提取子域名
				t.extractSubdomainsFromTXT(record, domain)
			}
		}
	}

	return nil
}

// extractSubdomainsFromTXT 从 TXT 记录中提取子域名
func (t *TXT) extractSubdomainsFromTXT(txtRecord, domain string) {
	// 使用正则表达式或字符串匹配来提取子域名
	// 这里简化处理，实际可能需要更复杂的解析
	words := strings.Fields(txtRecord)
	for _, word := range words {
		// 检查是否包含域名
		if strings.Contains(word, domain) {
			// 提取可能的子域名
			if t.IsValidSubdomain(word, domain) {
				t.AddSubdomain(word)
			}
		}
	}
}
