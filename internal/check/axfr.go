package check

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// AXFR AXFR 检查模块
type AXFR struct {
	*core.Check
}

// NewAXFR 创建 AXFR 模块
func NewAXFR(cfg *config.Config) *AXFR {
	return &AXFR{
		Check: core.NewCheck("AXFRCheck", cfg),
	}
}

// Run 执行检查
func (a *AXFR) Run(domain string) ([]string, error) {
	a.SetDomain(domain)
	a.Begin()
	defer a.Finish()

	// 获取域名服务器
	nameservers, err := a.getNameservers(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get nameservers: %v", err)
	}

	// 对每个域名服务器执行域传送
	for _, nameserver := range nameservers {
		subdomains, err := a.performZoneTransfer(domain, nameserver)
		if err != nil {
			a.LogError("Failed to perform zone transfer on %s: %v", nameserver, err)
			continue
		}

		// 添加子域名
		for _, subdomain := range subdomains {
			a.AddSubdomain(subdomain)
		}
	}

	return a.GetSubdomains(), nil
}

// getNameservers 获取域名服务器
func (a *AXFR) getNameservers(domain string) ([]string, error) {
	var nameservers []string

	// 查询 NS 记录
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	msg.RecursionDesired = true

	resp, err := dns.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return nil, fmt.Errorf("failed to query NS records: %v", err)
	}

	// 提取域名服务器
	for _, answer := range resp.Answer {
		if ns, ok := answer.(*dns.NS); ok {
			nameserver := strings.TrimSuffix(ns.Ns, ".")
			nameservers = append(nameservers, nameserver)
		}
	}

	return nameservers, nil
}

// performZoneTransfer 执行域传送
func (a *AXFR) performZoneTransfer(domain, nameserver string) ([]string, error) {
	var subdomains []string

	// 创建传输对象
	transfer := new(dns.Transfer)
	msg := new(dns.Msg)
	msg.SetAxfr(dns.Fqdn(domain))

	// 执行域传送
	envelope, err := transfer.In(msg, nameserver+":53")
	if err != nil {
		return nil, fmt.Errorf("failed to initiate zone transfer: %v", err)
	}

	// 处理响应
	for env := range envelope {
		if env.Error != nil {
			continue
		}

		for _, rr := range env.RR {
			switch rr.Header().Rrtype {
			case dns.TypeA, dns.TypeCNAME, dns.TypeNS, dns.TypeMX, dns.TypeTXT:
				name := strings.TrimSuffix(rr.Header().Name, ".")
				if a.IsValidSubdomain(name, domain) {
					subdomains = append(subdomains, name)
				}
			}
		}
	}

	return subdomains, nil
}
