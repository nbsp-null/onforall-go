package check

import (
	"strings"

	"github.com/miekg/dns"
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// NSEC NSEC 检查模块
type NSEC struct {
	*core.Check
	domain string
}

// NewNSEC 创建 NSEC 检查模块
func NewNSEC(cfg *config.Config) *NSEC {
	return &NSEC{
		Check: core.NewCheck("NSECCheck", cfg),
	}
}

// Run 执行检查
func (n *NSEC) Run(domain string) ([]string, error) {
	n.SetDomain(domain)
	n.Begin()
	defer n.Finish()

	// 执行检查
	if err := n.walk(domain); err != nil {
		return nil, err
	}

	return n.GetSubdomains(), nil
}

// walk 执行 NSEC 区域遍历
func (n *NSEC) walk(domain string) error {
	currentDomain := domain

	for {
		// 查询 NSEC 记录
		answer, err := n.dnsQuery(currentDomain, "NSEC")
		if err != nil {
			break
		}

		if len(answer) == 0 {
			break
		}

		// 处理 NSEC 记录
		var subdomain string
		for _, item := range answer {
			record := item.String()
			subdomains := n.ExtractSubdomains(record, domain)

			// 通常只有一个子域名
			if len(subdomains) > 0 {
				subdomain = subdomains[0]
				n.AddSubdomain(subdomain)
			}
		}

		// 检查是否完成循环
		if subdomain == domain {
			break
		}

		// 防止无限循环
		if currentDomain != domain {
			currentParts := strings.Split(currentDomain, ".")
			subdomainParts := strings.Split(subdomain, ".")
			if len(currentParts) > 0 && len(subdomainParts) > 0 {
				if currentParts[0] == subdomainParts[0] {
					break
				}
			}
		}

		currentDomain = subdomain
	}

	return nil
}

// dnsQuery 执行 DNS 查询
func (n *NSEC) dnsQuery(domain, recordType string) ([]dns.RR, error) {
	// 创建 DNS 客户端
	client := new(dns.Client)

	// 创建查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNSEC)
	msg.RecursionDesired = true

	// 使用默认 DNS 服务器
	resp, _, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return nil, err
	}

	return resp.Answer, nil
}
