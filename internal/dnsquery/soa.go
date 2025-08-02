package dnsquery

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// SOA SOA 查询模块
type SOA struct {
	*core.Query
}

// NewSOA 创建 SOA 查询模块
func NewSOA(cfg *config.Config) *SOA {
	return &SOA{
		Query: core.NewQuery("QuerySOA", cfg),
	}
}

// Run 执行查询
func (s *SOA) Run(domain string) ([]string, error) {
	s.SetDomain(domain)
	s.Begin()
	defer s.Finish()

	// 执行 SOA 查询
	if err := s.query(domain); err != nil {
		return nil, err
	}

	return s.GetSubdomains(), nil
}

// query 执行查询
func (s *SOA) query(domain string) error {
	// 创建 DNS 查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeSOA)
	msg.RecursionDesired = true

	// 发送查询
	resp, err := dns.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to query SOA records: %v", err)
	}

	// 处理响应
	for _, answer := range resp.Answer {
		if soa, ok := answer.(*dns.SOA); ok {
			// SOA 记录中的 NS 字段可能包含子域名
			ns := strings.TrimSuffix(soa.Ns, ".")
			if s.IsValidSubdomain(ns, domain) {
				s.AddSubdomain(ns)
			}
		}
	}

	return nil
}
