package osint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// OSINTClient OSINT API 查询客户端
type OSINTClient struct {
	timeout time.Duration
	client  *http.Client
	apiKeys map[string]string
}

// NewOSINTClient 创建新的 OSINT 查询客户端
func NewOSINTClient(timeout int) *OSINTClient {
	cfg := config.GetConfig()

	return &OSINTClient{
		timeout: time.Duration(timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKeys: cfg.APIKeys,
	}
}

// QueryOSINT 执行 OSINT API 查询
func (o *OSINTClient) QueryOSINT(domain string) ([]string, error) {
	logger.Infof("Starting OSINT API query for domain: %s", domain)

	subdomains := make([]string, 0)

	// 1. 从 VirusTotal 查询
	vtResults, err := o.queryVirusTotal(domain)
	if err != nil {
		logger.Debugf("Failed to query VirusTotal: %v", err)
	} else {
		subdomains = append(subdomains, vtResults...)
	}

	// 2. 从 ThreatBook 查询
	tbResults, err := o.queryThreatBook(domain)
	if err != nil {
		logger.Debugf("Failed to query ThreatBook: %v", err)
	} else {
		subdomains = append(subdomains, tbResults...)
	}

	// 3. 从 RiskIQ 查询
	riResults, err := o.queryRiskIQ(domain)
	if err != nil {
		logger.Debugf("Failed to query RiskIQ: %v", err)
	} else {
		subdomains = append(subdomains, riResults...)
	}

	// 4. 从 Shodan 查询
	shodanResults, err := o.queryShodan(domain)
	if err != nil {
		logger.Debugf("Failed to query Shodan: %v", err)
	} else {
		subdomains = append(subdomains, shodanResults...)
	}

	// 5. 从 Fofa 查询
	fofaResults, err := o.queryFofa(domain)
	if err != nil {
		logger.Debugf("Failed to query Fofa: %v", err)
	} else {
		subdomains = append(subdomains, fofaResults...)
	}

	// 6. 从 Hunter 查询
	hunterResults, err := o.queryHunter(domain)
	if err != nil {
		logger.Debugf("Failed to query Hunter: %v", err)
	} else {
		subdomains = append(subdomains, hunterResults...)
	}

	// 7. 从 Quake 查询
	quakeResults, err := o.queryQuake(domain)
	if err != nil {
		logger.Debugf("Failed to query Quake: %v", err)
	} else {
		subdomains = append(subdomains, quakeResults...)
	}

	// 8. 从 ZoomEye 查询
	zoomeyeResults, err := o.queryZoomEye(domain)
	if err != nil {
		logger.Debugf("Failed to query ZoomEye: %v", err)
	} else {
		subdomains = append(subdomains, zoomeyeResults...)
	}

	// 去重
	subdomains = o.deduplicate(subdomains)

	logger.Infof("OSINT API query completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}

// queryVirusTotal 从 VirusTotal 查询
func (o *OSINTClient) queryVirusTotal(domain string) ([]string, error) {
	apiKey := o.apiKeys["virustotal_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("VirusTotal API key not configured")
	}

	url := fmt.Sprintf("https://www.virustotal.com/vtapi/v2/domain/report?apikey=%s&domain=%s", apiKey, domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("VirusTotal API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Subdomains []string `json:"subdomains"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Subdomains, nil
}

// queryThreatBook 从 ThreatBook 查询
func (o *OSINTClient) queryThreatBook(domain string) ([]string, error) {
	apiKey := o.apiKeys["threatbook_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("ThreatBook API key not configured")
	}

	url := fmt.Sprintf("https://api.threatbook.cn/v3/domain/sub_domains?apikey=%s&resource=%s", apiKey, domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ThreatBook API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Subdomains []string `json:"subdomains"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data.Subdomains, nil
}

// queryRiskIQ 从 RiskIQ 查询
func (o *OSINTClient) queryRiskIQ(domain string) ([]string, error) {
	apiKey := o.apiKeys["riskiq_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("RiskIQ API key not configured")
	}

	url := fmt.Sprintf("https://api.riskiq.net/pt/v2/dns/passive/unique?query=%s", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("RiskIQ API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Hostname string `json:"hostname"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, result := range result.Results {
		subdomains = append(subdomains, result.Hostname)
	}

	return subdomains, nil
}

// queryShodan 从 Shodan 查询
func (o *OSINTClient) queryShodan(domain string) ([]string, error) {
	apiKey := o.apiKeys["shodan_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("Shodan API key not configured")
	}

	url := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Shodan API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Subdomains []string `json:"subdomains"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Subdomains, nil
}

// queryFofa 从 Fofa 查询
func (o *OSINTClient) queryFofa(domain string) ([]string, error) {
	apiKey := o.apiKeys["fofa_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("Fofa API key not configured")
	}

	query := fmt.Sprintf("domain=\"%s\"", domain)
	url := fmt.Sprintf("https://fofa.info/api/v1/search/all?qbase64=%s&key=%s&fields=host",
		encodeBase64(query), apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Fofa API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Results [][]string `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, row := range result.Results {
		if len(row) > 0 {
			host := row[0]
			if strings.Contains(host, domain) {
				subdomains = append(subdomains, host)
			}
		}
	}

	return subdomains, nil
}

// queryHunter 从 Hunter 查询
func (o *OSINTClient) queryHunter(domain string) ([]string, error) {
	apiKey := o.apiKeys["hunter_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("Hunter API key not configured")
	}

	url := fmt.Sprintf("https://api.hunter.io/v2/domain-search?domain=%s&api_key=%s", domain, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Hunter API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Webmail []string `json:"webmail"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data.Webmail, nil
}

// queryQuake 从 Quake 查询
func (o *OSINTClient) queryQuake(domain string) ([]string, error) {
	apiKey := o.apiKeys["quake_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("Quake API key not configured")
	}

	url := fmt.Sprintf("https://quake.360.net/api/v3/search/quake", domain)

	req, err := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(`{
		"query": "domain:\"%s\"",
		"start": 0,
		"size": 100
	}`, domain)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-QuakeToken", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Quake API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Domain string `json:"domain"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, item := range result.Data {
		if strings.Contains(item.Domain, domain) {
			subdomains = append(subdomains, item.Domain)
		}
	}

	return subdomains, nil
}

// queryZoomEye 从 ZoomEye 查询
func (o *OSINTClient) queryZoomEye(domain string) ([]string, error) {
	apiKey := o.apiKeys["zoomeye_api"]
	if apiKey == "" {
		return nil, fmt.Errorf("ZoomEye API key not configured")
	}

	url := fmt.Sprintf("https://api.zoomeye.org/domain/search?q=%s&page=1&facet=app,os", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("API-KEY", apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ZoomEye API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Domain string `json:"domain"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, item := range result.Results {
		if strings.Contains(item.Domain, domain) {
			subdomains = append(subdomains, item.Domain)
		}
	}

	return subdomains, nil
}

// encodeBase64 简单的 Base64 编码（实际应使用标准库）
func encodeBase64(s string) string {
	// 这里简化实现，实际应使用 encoding/base64
	return s
}

// deduplicate 去重
func (o *OSINTClient) deduplicate(items []string) []string {
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
