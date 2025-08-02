package certificates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// CertificateClient 证书透明度查询客户端
type CertificateClient struct {
	timeout time.Duration
	client  *http.Client
	apiKey  string
}

// CertData 证书数据结构
type CertData struct {
	NameValue  string `json:"name_value"`
	IssuedDate string `json:"issued_date"`
	NotAfter   string `json:"not_after"`
	IssuerName string `json:"issuer_name"`
}

// NewCertificateClient 创建新的证书查询客户端
func NewCertificateClient(timeout int) *CertificateClient {
	cfg := config.GetConfig()
	return &CertificateClient{
		timeout: time.Duration(timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKey: cfg.APIKeys["censys_api"],
	}
}

// QueryCertificates 查询证书透明度日志
func (c *CertificateClient) QueryCertificates(domain string) ([]string, error) {
	logger.Infof("Starting certificate transparency query for domain: %s", domain)

	subdomains := make([]string, 0)

	// 1. 从 Censys 查询
	censysResults, err := c.queryCensys(domain)
	if err != nil {
		logger.Debugf("Failed to query Censys: %v", err)
	} else {
		subdomains = append(subdomains, censysResults...)
	}

	// 2. 从 CertSpotter 查询
	certspotterResults, err := c.queryCertSpotter(domain)
	if err != nil {
		logger.Debugf("Failed to query CertSpotter: %v", err)
	} else {
		subdomains = append(subdomains, certspotterResults...)
	}

	// 3. 从 crt.sh 查询
	crtshResults, err := c.queryCrtSh(domain)
	if err != nil {
		logger.Debugf("Failed to query crt.sh: %v", err)
	} else {
		subdomains = append(subdomains, crtshResults...)
	}

	// 4. 从 Google 透明度查询
	googleResults, err := c.queryGoogleTransparency(domain)
	if err != nil {
		logger.Debugf("Failed to query Google Transparency: %v", err)
	} else {
		subdomains = append(subdomains, googleResults...)
	}

	// 去重
	subdomains = c.deduplicate(subdomains)

	logger.Infof("Certificate transparency query completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}

// queryCensys 从 Censys 查询证书
func (c *CertificateClient) queryCensys(domain string) ([]string, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Censys API key not configured")
	}

	url := fmt.Sprintf("https://search.censys.io/api/v2/certificates/search?q=%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Censys API returned status: %d", resp.StatusCode)
	}

	// 解析响应（简化实现）
	var result struct {
		Result struct {
			Hits []struct {
				Names []string `json:"names"`
			} `json:"hits"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, hit := range result.Result.Hits {
		for _, name := range hit.Names {
			if strings.Contains(name, domain) {
				subdomains = append(subdomains, name)
			}
		}
	}

	return subdomains, nil
}

// queryCertSpotter 从 CertSpotter 查询证书
func (c *CertificateClient) queryCertSpotter(domain string) ([]string, error) {
	url := fmt.Sprintf("https://api.certspotter.com/v1/issuances?domain=%s&include_subdomains=true&expand=dns_names", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("CertSpotter API returned status: %d", resp.StatusCode)
	}

	var certs []CertData
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, cert := range certs {
		if strings.Contains(cert.NameValue, domain) {
			subdomains = append(subdomains, cert.NameValue)
		}
	}

	return subdomains, nil
}

// queryCrtSh 从 crt.sh 查询证书
func (c *CertificateClient) queryCrtSh(domain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("crt.sh API returned status: %d", resp.StatusCode)
	}

	var certs []struct {
		NameValue string `json:"name_value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, cert := range certs {
		names := strings.Split(cert.NameValue, "\n")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if strings.Contains(name, domain) && name != domain {
				subdomains = append(subdomains, name)
			}
		}
	}

	return subdomains, nil
}

// queryGoogleTransparency 从 Google 透明度日志查询
func (c *CertificateClient) queryGoogleTransparency(domain string) ([]string, error) {
	// Google 透明度日志查询（简化实现）
	// 实际实现需要更复杂的 CT 日志查询逻辑
	url := fmt.Sprintf("https://transparencyreport.google.com/transparencyreport/api/v3/httpsreport/ct?domain=%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 这里返回空结果，因为 Google 透明度日志的查询比较复杂
	// 实际实现需要解析 CT 日志格式
	return []string{}, nil
}

// deduplicate 去重
func (c *CertificateClient) deduplicate(items []string) []string {
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
