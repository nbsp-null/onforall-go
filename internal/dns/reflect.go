package dns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/oneforall-go/pkg/logger"
	"golang.org/x/sync/semaphore"
)

// ReflectClient DNS 反射查询客户端
type ReflectClient struct {
	timeout     time.Duration
	concurrency int64
	semaphore   *semaphore.Weighted
	resolvers   []string
}

// NewReflectClient 创建新的 DNS 反射查询客户端
func NewReflectClient(timeout int, concurrency int) *ReflectClient {
	return &ReflectClient{
		timeout:     time.Duration(timeout) * time.Second,
		concurrency: int64(concurrency),
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		resolvers:   getDefaultResolvers(),
	}
}

// QueryReflect 执行 DNS 反射查询
func (r *ReflectClient) QueryReflect(domain string) ([]string, error) {
	logger.Infof("Starting DNS reflection query for domain: %s", domain)

	// 获取信号量
	if err := r.semaphore.Acquire(context.Background(), 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %v", err)
	}
	defer r.semaphore.Release(1)

	subdomains := make([]string, 0)

	// 1. 查询 NS 记录
	nsRecords, err := r.queryNSRecords(domain)
	if err != nil {
		logger.Debugf("Failed to query NS records: %v", err)
	} else {
		subdomains = append(subdomains, nsRecords...)
	}

	// 2. 查询 MX 记录
	mxRecords, err := r.queryMXRecords(domain)
	if err != nil {
		logger.Debugf("Failed to query MX records: %v", err)
	} else {
		subdomains = append(subdomains, mxRecords...)
	}

	// 3. 查询 TXT 记录
	txtRecords, err := r.queryTXTRecords(domain)
	if err != nil {
		logger.Debugf("Failed to query TXT records: %v", err)
	} else {
		subdomains = append(subdomains, txtRecords...)
	}

	// 4. 查询 SOA 记录
	soaRecords, err := r.querySOARecords(domain)
	if err != nil {
		logger.Debugf("Failed to query SOA records: %v", err)
	} else {
		subdomains = append(subdomains, soaRecords...)
	}

	// 5. 查询 SRV 记录
	srvRecords, err := r.querySRVRecords(domain)
	if err != nil {
		logger.Debugf("Failed to query SRV records: %v", err)
	} else {
		subdomains = append(subdomains, srvRecords...)
	}

	// 去重
	subdomains = r.deduplicate(subdomains)

	logger.Infof("DNS reflection query completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}

// queryNSRecords 查询 NS 记录
func (r *ReflectClient) queryNSRecords(domain string) ([]string, error) {
	client := &dns.Client{Timeout: r.timeout}
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	msg.RecursionDesired = true

	resp, _, err := client.Exchange(msg, r.resolvers[0])
	if err != nil {
		return nil, err
	}

	var subdomains []string
	for _, answer := range resp.Answer {
		if ns, ok := answer.(*dns.NS); ok {
			subdomain := strings.TrimSuffix(ns.Ns, ".")
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains, nil
}

// queryMXRecords 查询 MX 记录
func (r *ReflectClient) queryMXRecords(domain string) ([]string, error) {
	client := &dns.Client{Timeout: r.timeout}
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeMX)
	msg.RecursionDesired = true

	resp, _, err := client.Exchange(msg, r.resolvers[0])
	if err != nil {
		return nil, err
	}

	var subdomains []string
	for _, answer := range resp.Answer {
		if mx, ok := answer.(*dns.MX); ok {
			subdomain := strings.TrimSuffix(mx.Mx, ".")
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains, nil
}

// queryTXTRecords 查询 TXT 记录
func (r *ReflectClient) queryTXTRecords(domain string) ([]string, error) {
	client := &dns.Client{Timeout: r.timeout}
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	msg.RecursionDesired = true

	resp, _, err := client.Exchange(msg, r.resolvers[0])
	if err != nil {
		return nil, err
	}

	var subdomains []string
	for _, answer := range resp.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			// 从 TXT 记录中提取可能的子域名
			for _, txtStr := range txt.Txt {
				// 简单的子域名提取逻辑
				if strings.Contains(txtStr, ".") && !strings.Contains(txtStr, " ") {
					subdomains = append(subdomains, txtStr)
				}
			}
		}
	}

	return subdomains, nil
}

// querySOARecords 查询 SOA 记录
func (r *ReflectClient) querySOARecords(domain string) ([]string, error) {
	client := &dns.Client{Timeout: r.timeout}
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeSOA)
	msg.RecursionDesired = true

	resp, _, err := client.Exchange(msg, r.resolvers[0])
	if err != nil {
		return nil, err
	}

	var subdomains []string
	for _, answer := range resp.Answer {
		if soa, ok := answer.(*dns.SOA); ok {
			// SOA 记录中的 MName 和 RName 可能包含子域名
			mname := strings.TrimSuffix(soa.Mbox, ".")
			if mname != domain {
				subdomains = append(subdomains, mname)
			}
		}
	}

	return subdomains, nil
}

// querySRVRecords 查询 SRV 记录
func (r *ReflectClient) querySRVRecords(domain string) ([]string, error) {
	// 常见的 SRV 记录前缀
	srvPrefixes := []string{"_ldap", "_kerberos", "_kpasswd", "_dns", "_ntp", "_sip", "_xmpp", "_imap", "_pop3", "_smtp"}

	var allSubdomains []string
	for _, prefix := range srvPrefixes {
		srvDomain := prefix + "._tcp." + domain
		client := &dns.Client{Timeout: r.timeout}
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(srvDomain), dns.TypeSRV)
		msg.RecursionDesired = true

		resp, _, err := client.Exchange(msg, r.resolvers[0])
		if err != nil {
			continue
		}

		for _, answer := range resp.Answer {
			if srv, ok := answer.(*dns.SRV); ok {
				target := strings.TrimSuffix(srv.Target, ".")
				if target != domain {
					allSubdomains = append(allSubdomains, target)
				}
			}
		}
	}

	return allSubdomains, nil
}

// deduplicate 去重
func (r *ReflectClient) deduplicate(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
