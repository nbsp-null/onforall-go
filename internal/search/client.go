package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// SearchClient 搜索引擎查询客户端
type SearchClient struct {
	timeout time.Duration
	client  *http.Client
	apiKeys map[string]string
}

// NewSearchClient 创建新的搜索引擎查询客户端
func NewSearchClient(timeout int) *SearchClient {
	cfg := config.GetConfig()

	return &SearchClient{
		timeout: time.Duration(timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKeys: cfg.APIKeys,
	}
}

// QuerySearchEngines 执行搜索引擎查询
func (s *SearchClient) QuerySearchEngines(domain string) ([]string, error) {
	logger.Infof("Starting search engine query for domain: %s", domain)

	subdomains := make([]string, 0)

	// 1. 从 Google 查询
	googleResults, err := s.queryGoogle(domain)
	if err != nil {
		logger.Debugf("Failed to query Google: %v", err)
	} else {
		subdomains = append(subdomains, googleResults...)
	}

	// 2. 从 Bing 查询
	bingResults, err := s.queryBing(domain)
	if err != nil {
		logger.Debugf("Failed to query Bing: %v", err)
	} else {
		subdomains = append(subdomains, bingResults...)
	}

	// 3. 从 Baidu 查询
	baiduResults, err := s.queryBaidu(domain)
	if err != nil {
		logger.Debugf("Failed to query Baidu: %v", err)
	} else {
		subdomains = append(subdomains, baiduResults...)
	}

	// 4. 从 GitHub 查询
	githubResults, err := s.queryGitHub(domain)
	if err != nil {
		logger.Debugf("Failed to query GitHub: %v", err)
	} else {
		subdomains = append(subdomains, githubResults...)
	}

	// 5. 从 Yahoo 查询
	yahooResults, err := s.queryYahoo(domain)
	if err != nil {
		logger.Debugf("Failed to query Yahoo: %v", err)
	} else {
		subdomains = append(subdomains, yahooResults...)
	}

	// 6. 从 DuckDuckGo 查询
	duckduckgoResults, err := s.queryDuckDuckGo(domain)
	if err != nil {
		logger.Debugf("Failed to query DuckDuckGo: %v", err)
	} else {
		subdomains = append(subdomains, duckduckgoResults...)
	}

	// 去重
	subdomains = s.deduplicate(subdomains)

	logger.Infof("Search engine query completed, found %d subdomains", len(subdomains))
	return subdomains, nil
}

// queryGoogle 从 Google 查询
func (s *SearchClient) queryGoogle(domain string) ([]string, error) {
	apiKey := s.apiKeys["google_api"]
	if apiKey == "" {
		// 如果没有 API key，使用普通搜索（需要解析 HTML）
		return s.queryGoogleWeb(domain)
	}

	// 使用 Google Custom Search API
	url := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=YOUR_CX&q=site:%s", apiKey, domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Link string `json:"link"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, item := range result.Items {
		subdomain := s.extractSubdomainFromURL(item.Link, domain)
		if subdomain != "" {
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains, nil
}

// queryGoogleWeb 从 Google 网页搜索查询（无 API key）
func (s *SearchClient) queryGoogleWeb(domain string) ([]string, error) {
	searchQuery := fmt.Sprintf("site:%s", domain)
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&num=100", url.QueryEscape(searchQuery))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析 HTML 响应（简化实现）
	// 实际实现需要更复杂的 HTML 解析
	return s.parseGoogleResults(resp.Body), nil
}

// queryBing 从 Bing 查询
func (s *SearchClient) queryBing(domain string) ([]string, error) {
	apiKey := s.apiKeys["bing_api"]
	if apiKey == "" {
		return s.queryBingWeb(domain)
	}

	url := fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=site:%s&count=50", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bing API returned status: %d", resp.StatusCode)
	}

	var result struct {
		WebPages struct {
			Value []struct {
				URL string `json:"url"`
			} `json:"value"`
		} `json:"webPages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, item := range result.WebPages.Value {
		subdomain := s.extractSubdomainFromURL(item.URL, domain)
		if subdomain != "" {
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains, nil
}

// queryBingWeb 从 Bing 网页搜索查询
func (s *SearchClient) queryBingWeb(domain string) ([]string, error) {
	searchQuery := fmt.Sprintf("site:%s", domain)
	searchURL := fmt.Sprintf("https://www.bing.com/search?q=%s&count=50", url.QueryEscape(searchQuery))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析 HTML 响应
	return s.parseBingResults(resp.Body), nil
}

// queryBaidu 从百度查询
func (s *SearchClient) queryBaidu(domain string) ([]string, error) {
	searchQuery := fmt.Sprintf("site:%s", domain)
	searchURL := fmt.Sprintf("https://www.baidu.com/s?wd=%s&rn=50", url.QueryEscape(searchQuery))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析 HTML 响应
	return s.parseBaiduResults(resp.Body), nil
}

// queryGitHub 从 GitHub 查询
func (s *SearchClient) queryGitHub(domain string) ([]string, error) {
	apiKey := s.apiKeys["github_api"]

	searchQuery := fmt.Sprintf("%s", domain)
	searchURL := fmt.Sprintf("https://api.github.com/search/code?q=%s", url.QueryEscape(searchQuery))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "token "+apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			HTMLURL string `json:"html_url"`
			Path    string `json:"path"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, item := range result.Items {
		// 从 GitHub 代码中提取可能的子域名
		subdomains = append(subdomains, s.extractSubdomainsFromText(item.Path)...)
	}

	return subdomains, nil
}

// queryYahoo 从 Yahoo 查询
func (s *SearchClient) queryYahoo(domain string) ([]string, error) {
	searchQuery := fmt.Sprintf("site:%s", domain)
	searchURL := fmt.Sprintf("https://search.yahoo.com/search?p=%s&n=50", url.QueryEscape(searchQuery))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析 HTML 响应
	return s.parseYahooResults(resp.Body), nil
}

// queryDuckDuckGo 从 DuckDuckGo 查询
func (s *SearchClient) queryDuckDuckGo(domain string) ([]string, error) {
	searchQuery := fmt.Sprintf("site:%s", domain)
	searchURL := fmt.Sprintf("https://duckduckgo.com/html/?q=%s", url.QueryEscape(searchQuery))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析 HTML 响应
	return s.parseDuckDuckGoResults(resp.Body), nil
}

// extractSubdomainFromURL 从 URL 中提取子域名
func (s *SearchClient) extractSubdomainFromURL(urlStr, domain string) string {
	// 简单的子域名提取逻辑
	if strings.Contains(urlStr, domain) {
		// 提取子域名
		re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9.-]+\.%s)`, regexp.QuoteMeta(domain)))
		matches := re.FindStringSubmatch(urlStr)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// extractSubdomainsFromText 从文本中提取子域名
func (s *SearchClient) extractSubdomainsFromText(text string) []string {
	var subdomains []string
	// 简单的正则表达式匹配子域名
	re := regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		if strings.Contains(match, ".") {
			subdomains = append(subdomains, match)
		}
	}

	return subdomains
}

// parseGoogleResults 解析 Google 搜索结果（简化实现）
func (s *SearchClient) parseGoogleResults(body io.Reader) []string {
	// 这里应该实现 HTML 解析逻辑
	// 简化实现，返回空结果
	return []string{}
}

// parseBingResults 解析 Bing 搜索结果
func (s *SearchClient) parseBingResults(body io.Reader) []string {
	// 这里应该实现 HTML 解析逻辑
	return []string{}
}

// parseBaiduResults 解析百度搜索结果
func (s *SearchClient) parseBaiduResults(body io.Reader) []string {
	// 这里应该实现 HTML 解析逻辑
	return []string{}
}

// parseYahooResults 解析 Yahoo 搜索结果
func (s *SearchClient) parseYahooResults(body io.Reader) []string {
	// 这里应该实现 HTML 解析逻辑
	return []string{}
}

// parseDuckDuckGoResults 解析 DuckDuckGo 搜索结果
func (s *SearchClient) parseDuckDuckGoResults(body io.Reader) []string {
	// 这里应该实现 HTML 解析逻辑
	return []string{}
}

// deduplicate 去重
func (s *SearchClient) deduplicate(items []string) []string {
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
