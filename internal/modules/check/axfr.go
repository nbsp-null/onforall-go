package check

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/modules"
)

// AXFR AXFR 检查模块
type AXFR struct {
	*modules.BaseModule
}

// NewAXFR 创建 AXFR 模块
func NewAXFR(cfg *config.Config) *AXFR {
	return &AXFR{
		BaseModule: modules.NewBaseModule("AXFR", modules.ModuleTypeCheck, cfg),
	}
}

// Run 执行检查
func (a *AXFR) Run(domain string) ([]string, error) {
	a.LogInfo("Starting AXFR check for domain: %s", domain)

	var allSubdomains []string

	// 获取域名服务器
	nameservers, err := a.getNameservers(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get nameservers: %v", err)
	}

	// 对每个域名服务器尝试域传送
	for _, ns := range nameservers {
		subdomains, err := a.performZoneTransfer(domain, ns)
		if err != nil {
			a.LogDebug("Zone transfer failed for %s: %v", ns, err)
			continue
		}
		allSubdomains = append(allSubdomains, subdomains...)
	}

	// 去重
	allSubdomains = a.Deduplicate(allSubdomains)

	a.LogInfo("AXFR check completed, found %d subdomains", len(allSubdomains))
	return allSubdomains, nil
}

// getNameservers 获取域名服务器
func (a *AXFR) getNameservers(domain string) ([]string, error) {
	client := &dns.Client{Timeout: 10 * 1000000000} // 10 seconds
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	msg.RecursionDesired = true

	resp, _, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return nil, err
	}

	var nameservers []string
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
	a.LogDebug("Trying zone transfer for %s on %s", domain, nameserver)

	// 尝试域传送
	transfer := new(dns.Transfer)
	msg := new(dns.Msg)
	msg.SetAxfr(dns.Fqdn(domain))

	response, err := transfer.In(msg, nameserver+":53")
	if err != nil {
		return nil, err
	}

	var subdomains []string
	for env := range response {
		if env.Error != nil {
			continue
		}

		for _, rr := range env.RR {
			switch v := rr.(type) {
			case *dns.A:
				name := strings.TrimSuffix(v.Header().Name, ".")
				if a.IsValidSubdomain(name, domain) {
					subdomains = append(subdomains, name)
				}
			case *dns.CNAME:
				name := strings.TrimSuffix(v.Header().Name, ".")
				if a.IsValidSubdomain(name, domain) {
					subdomains = append(subdomains, name)
				}
			case *dns.MX:
				name := strings.TrimSuffix(v.Header().Name, ".")
				if a.IsValidSubdomain(name, domain) {
					subdomains = append(subdomains, name)
				}
			case *dns.TXT:
				name := strings.TrimSuffix(v.Header().Name, ".")
				if a.IsValidSubdomain(name, domain) {
					subdomains = append(subdomains, name)
				}
			}
		}
	}

	if len(subdomains) > 0 {
		a.LogDebug("Zone transfer successful for %s on %s", domain, nameserver)
	}

	return subdomains, nil
}
