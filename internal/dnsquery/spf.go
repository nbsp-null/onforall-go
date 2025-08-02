package dnsquery

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// SPF SPF 查询模块
type SPF struct {
	*core.Query
}

// NewSPF 创建 SPF 查询模块
func NewSPF(cfg *config.Config) *SPF {
	return &SPF{
		Query: core.NewQuery("QuerySPF", cfg),
	}
}

// Run 执行查询
func (s *SPF) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 执行 SPF 查询
	if err := s.query(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// query 执行查询
func (s *SPF) query(domain string) error {
	// 创建 DNS 查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	msg.RecursionDesired = true

	// 发送查询
	resp, err := dns.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to query SPF records: %v", err)
	}

	// 处理响应
	for _, answer := range resp.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			for _, record := range txt.Txt {
				// 检查是否是 SPF 记录
				if strings.HasPrefix(record, "v=spf1") {
					// 从 SPF 记录中提取子域名
					s.extractSubdomainsFromSPF(record, domain)
				}
			}
		}
	}

	return nil
}

// extractSubdomainsFromSPF 从 SPF 记录中提取子域名
func (s *SPF) extractSubdomainsFromSPF(spfRecord, domain string) {
	// SPF 记录格式：v=spf1 include:_spf.google.com ~all
	// 查找 include: 后面的域名
	parts := strings.Fields(spfRecord)
	for _, part := range parts {
		if strings.HasPrefix(part, "include:") {
			host := strings.TrimPrefix(part, "include:")
			if s.IsValidSubdomain(host, domain) {
				s.AddSubdomain(host)
			}
		}
	}
}
