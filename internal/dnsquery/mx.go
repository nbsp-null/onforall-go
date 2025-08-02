package dnsquery

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// MX MX 查询模块
type MX struct {
	*core.Query
}

// NewMX 创建 MX 查询模块
func NewMX(cfg *config.Config) *MX {
	return &MX{
		Query: core.NewQuery("QueryMX", cfg),
	}
}

// Run 执行查询
func (m *MX) Run(domain string) ([]string, error) {
	m.SetDomain(domain)
	m.Begin()
	defer m.Finish()

	// 执行 MX 查询
	if err := m.query(domain); err != nil {
		return nil, err
	}

	return m.GetSubdomains(), nil
}

// query 执行查询
func (m *MX) query(domain string) error {
	// 创建 DNS 查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeMX)
	msg.RecursionDesired = true

	// 发送查询
	resp, err := dns.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to query MX records: %v", err)
	}

	// 处理响应
	for _, answer := range resp.Answer {
		if mx, ok := answer.(*dns.MX); ok {
			host := strings.TrimSuffix(mx.Mx, ".")
			if m.IsValidSubdomain(host, domain) {
				m.AddSubdomain(host)
			}
		}
	}

	return nil
}
