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

// Client DNS 客户端
type Client struct {
	timeout     time.Duration
	concurrency int64
	semaphore   *semaphore.Weighted
	resolvers   []string
}

// NewClient 创建新的 DNS 客户端
func NewClient(timeout int, concurrency int) *Client {
	return &Client{
		timeout:     time.Duration(timeout) * time.Second,
		concurrency: int64(concurrency),
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		resolvers:   getDefaultResolvers(),
	}
}

// Resolve 解析域名
func (c *Client) Resolve(domain string) ([]string, error) {
	// 获取信号量
	if err := c.semaphore.Acquire(context.Background(), 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %v", err)
	}
	defer c.semaphore.Release(1)

	ips := make([]string, 0)

	// 尝试多个 DNS 服务器
	for _, resolver := range c.resolvers {
		resolvedIPs, err := c.resolveWithServer(domain, resolver)
		if err != nil {
			logger.Debugf("Failed to resolve %s with %s: %v", domain, resolver, err)
			continue
		}
		ips = append(ips, resolvedIPs...)
	}

	// 去重
	ips = c.deduplicate(ips)

	return ips, nil
}

// resolveWithServer 使用指定服务器解析
func (c *Client) resolveWithServer(domain, server string) ([]string, error) {
	// 创建 DNS 客户端
	client := &dns.Client{
		Timeout: c.timeout,
	}

	// 创建查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	msg.RecursionDesired = true

	// 发送查询
	resp, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, fmt.Errorf("DNS query failed: %v", err)
	}

	// 解析响应
	ips := make([]string, 0)
	for _, answer := range resp.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}

	return ips, nil
}

// ResolveCNAME 解析 CNAME 记录
func (c *Client) ResolveCNAME(domain string) ([]string, error) {
	// 获取信号量
	if err := c.semaphore.Acquire(context.Background(), 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %v", err)
	}
	defer c.semaphore.Release(1)

	cnames := make([]string, 0)

	// 尝试多个 DNS 服务器
	for _, resolver := range c.resolvers {
		resolvedCNAMEs, err := c.resolveCNAMEWithServer(domain, resolver)
		if err != nil {
			logger.Debugf("Failed to resolve CNAME for %s with %s: %v", domain, resolver, err)
			continue
		}
		cnames = append(cnames, resolvedCNAMEs...)
	}

	// 去重
	cnames = c.deduplicate(cnames)

	return cnames, nil
}

// resolveCNAMEWithServer 使用指定服务器解析 CNAME
func (c *Client) resolveCNAMEWithServer(domain, server string) ([]string, error) {
	// 创建 DNS 客户端
	client := &dns.Client{
		Timeout: c.timeout,
	}

	// 创建查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)
	msg.RecursionDesired = true

	// 发送查询
	resp, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, fmt.Errorf("DNS CNAME query failed: %v", err)
	}

	// 解析响应
	cnames := make([]string, 0)
	for _, answer := range resp.Answer {
		if cname, ok := answer.(*dns.CNAME); ok {
			cnames = append(cnames, strings.TrimSuffix(cname.Target, "."))
		}
	}

	return cnames, nil
}

// ResolveMX 解析 MX 记录
func (c *Client) ResolveMX(domain string) ([]string, error) {
	// 获取信号量
	if err := c.semaphore.Acquire(context.Background(), 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %v", err)
	}
	defer c.semaphore.Release(1)

	mxs := make([]string, 0)

	// 尝试多个 DNS 服务器
	for _, resolver := range c.resolvers {
		resolvedMXs, err := c.resolveMXWithServer(domain, resolver)
		if err != nil {
			logger.Debugf("Failed to resolve MX for %s with %s: %v", domain, resolver, err)
			continue
		}
		mxs = append(mxs, resolvedMXs...)
	}

	// 去重
	mxs = c.deduplicate(mxs)

	return mxs, nil
}

// resolveMXWithServer 使用指定服务器解析 MX
func (c *Client) resolveMXWithServer(domain, server string) ([]string, error) {
	// 创建 DNS 客户端
	client := &dns.Client{
		Timeout: c.timeout,
	}

	// 创建查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeMX)
	msg.RecursionDesired = true

	// 发送查询
	resp, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, fmt.Errorf("DNS MX query failed: %v", err)
	}

	// 解析响应
	mxs := make([]string, 0)
	for _, answer := range resp.Answer {
		if mx, ok := answer.(*dns.MX); ok {
			mxs = append(mxs, strings.TrimSuffix(mx.Mx, "."))
		}
	}

	return mxs, nil
}

// ResolveNS 解析 NS 记录
func (c *Client) ResolveNS(domain string) ([]string, error) {
	// 获取信号量
	if err := c.semaphore.Acquire(context.Background(), 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %v", err)
	}
	defer c.semaphore.Release(1)

	nss := make([]string, 0)

	// 尝试多个 DNS 服务器
	for _, resolver := range c.resolvers {
		resolvedNSs, err := c.resolveNSWithServer(domain, resolver)
		if err != nil {
			logger.Debugf("Failed to resolve NS for %s with %s: %v", domain, resolver, err)
			continue
		}
		nss = append(nss, resolvedNSs...)
	}

	// 去重
	nss = c.deduplicate(nss)

	return nss, nil
}

// resolveNSWithServer 使用指定服务器解析 NS
func (c *Client) resolveNSWithServer(domain, server string) ([]string, error) {
	// 创建 DNS 客户端
	client := &dns.Client{
		Timeout: c.timeout,
	}

	// 创建查询消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	msg.RecursionDesired = true

	// 发送查询
	resp, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, fmt.Errorf("DNS NS query failed: %v", err)
	}

	// 解析响应
	nss := make([]string, 0)
	for _, answer := range resp.Answer {
		if ns, ok := answer.(*dns.NS); ok {
			nss = append(nss, strings.TrimSuffix(ns.Ns, "."))
		}
	}

	return nss, nil
}

// deduplicate 去重
func (c *Client) deduplicate(items []string) []string {
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

// getDefaultResolvers 获取默认 DNS 服务器
func getDefaultResolvers() []string {
	return []string{
		"8.8.8.8:53",         // Google DNS
		"8.8.4.4:53",         // Google DNS
		"1.1.1.1:53",         // Cloudflare DNS
		"1.0.0.1:53",         // Cloudflare DNS
		"114.114.114.114:53", // 114 DNS
		"223.5.5.5:53",       // AliDNS
	}
}

// SetResolvers 设置 DNS 服务器
func (c *Client) SetResolvers(resolvers []string) {
	c.resolvers = resolvers
}

// GetResolvers 获取 DNS 服务器
func (c *Client) GetResolvers() []string {
	return c.resolvers
}
