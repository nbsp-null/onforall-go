package dnsquery

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// NS NS 查询模块
type NS struct {
	*core.Query
}

// NewNS 创建 NS 查询模块
func NewNS(cfg *config.Config) *NS {
	return &NS{
		Query: core.NewQuery("NSQuery", cfg),
	}
}

// Run 执行查询
func (n *NS) Run(domain string) ([]string, error) {
	n.SetDomain(domain)
	n.Begin()
	defer n.Finish()

	// 执行 NS 查询
	if err := n.query(domain); err != nil {
		return nil, err
	}

	return n.GetSubdomains(), nil
}

// query 执行查询
func (n *NS) query(domain string) error {
	// 创建 DNS 查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	msg.RecursionDesired = true

	// 发送查询
	resp, err := dns.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("failed to query NS records: %v", err)
	}

	// 处理响应
	for _, answer := range resp.Answer {
		if ns, ok := answer.(*dns.NS); ok {
			nameserver := strings.TrimSuffix(ns.Ns, ".")
			if n.IsValidSubdomain(nameserver, domain) {
				n.AddSubdomain(nameserver)
			}
		}
	}

	return nil
}
