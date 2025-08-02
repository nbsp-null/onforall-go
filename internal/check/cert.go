package check

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Cert Cert 检查模块
type Cert struct {
	*core.Check
}

// NewCert 创建 Cert 检查模块
func NewCert(cfg *config.Config) *Cert {
	return &Cert{
		Check: core.NewCheck("CertInfo", cfg),
	}
}

// Run 执行检查
func (c *Cert) Run(domain string) ([]string, error) {
	c.SetDomain(domain)
	c.Begin()
	defer c.Finish()

	// 执行检查
	if err := c.check(domain); err != nil {
		return nil, err
	}

	return c.GetSubdomains(), nil
}

// check 执行检查
func (c *Cert) check(domain string) error {
	// 创建 TLS 连接
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: 10 * time.Second,
	}, "tcp", domain+":443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to %s:443: %v", domain, err)
	}
	defer conn.Close()

	// 获取证书
	cert := conn.ConnectionState().PeerCertificates[0]

	// 从证书中提取子域名
	// 检查 Subject Alternative Names
	for _, name := range cert.DNSNames {
		if c.IsValidSubdomain(name, domain) {
			c.AddSubdomain(name)
		}
	}

	// 检查 Subject Common Name
	if cert.Subject.CommonName != "" {
		if c.IsValidSubdomain(cert.Subject.CommonName, domain) {
			c.AddSubdomain(cert.Subject.CommonName)
		}
	}

	return nil
}
