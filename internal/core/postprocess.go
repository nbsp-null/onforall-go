package core

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/pkg/logger"
)

// PostProcessHosts 按标题去重并对403做限流
// 仅当总数大于 cfg.ResultCheckLimit 时执行
func PostProcessHosts(hosts []string, cfg *config.Config) []SubdomainResult {
	limit := cfg.ResultCheckLimit
	if limit <= 0 {
		limit = 30
	}
	if len(hosts) <= limit {
		// 不处理，返回空表示无需替换
		return nil
	}

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 15 * time.Second}).DialContext,
			MaxIdleConns:        100,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
	insecureClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 15 * time.Second}).DialContext,
			MaxIdleConns:        100,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
	}

	titleRe := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`) // case-insensitive, dotall

	type tmp struct {
		host     string
		status   int
		title    string
		protocol string
	}

	var (
		mu        sync.Mutex
		okByTitle = make(map[string]tmp)
		okOrdered []tmp
		forbidden []tmp
	)

	sem := make(chan struct{}, cfg.ValidationConcurrency)
	var wg sync.WaitGroup

	fetch := func(host string) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()

		// 优先HTTPS（忽略证书）
		status, title := fetchOnce(insecureClient, ua, "https", host, titleRe)
		protocol := "https"
		if status != 200 {
			// 回退HTTP
			status, title = fetchOnce(client, ua, "http", host, titleRe)
			protocol = "http"
		}

		if status == 200 {
			mu.Lock()
			if _, exists := okByTitle[title]; !exists {
				rec := tmp{host: host, status: status, title: strings.TrimSpace(title), protocol: protocol}
				okByTitle[title] = rec
				okOrdered = append(okOrdered, rec)
			}
			mu.Unlock()
			return
		}
		if status == 403 {
			mu.Lock()
			forbidden = append(forbidden, tmp{host: host, status: status, title: strings.TrimSpace(title), protocol: protocol})
			mu.Unlock()
		}
	}

	for _, h := range uniqueStrings(hosts) {
		wg.Add(1)
		go fetch(h)
	}
	wg.Wait()

	// 403 限流
	appendForbidden := forbidden
	if len(forbidden) > limit {
		appendForbidden = forbidden[:1]
	}

	// 组装结果
	var final []SubdomainResult
	for _, r := range okOrdered {
		final = append(final, SubdomainResult{
			Subdomain:   r.host,
			Status:      r.status,
			Title:       r.title,
			Port:        map[string]int{"http": 80, "https": 443}[r.protocol],
			Alive:       true,
			Source:      "postprocess",
			Time:        time.Now().Format(time.RFC3339),
			DNSResolved: false,
			PingAlive:   false,
			StatusCode:  r.status,
			StatusText:  http.StatusText(r.status),
		})
	}
	for _, r := range appendForbidden {
		final = append(final, SubdomainResult{
			Subdomain:   r.host,
			Status:      r.status,
			Title:       r.title,
			Port:        map[string]int{"http": 80, "https": 443}[r.protocol],
			Alive:       false,
			Source:      "postprocess",
			Time:        time.Now().Format(time.RFC3339),
			DNSResolved: false,
			PingAlive:   false,
			StatusCode:  r.status,
			StatusText:  http.StatusText(r.status),
		})
	}

	logger.Infof("PostProcess: %d -> %d (200-title-dedup %d, 403 kept %d)", len(hosts), len(final), len(okOrdered), len(appendForbidden))
	return final
}

func fetchOnce(client *http.Client, ua, scheme, host string, titleRe *regexp.Regexp) (int, string) {
	url := fmt.Sprintf("%s://%s", scheme, host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, ""
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "identity") // 禁用压缩
	req.Header.Set("Connection", "close")

	resp, err := client.Do(req)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	// 限制读取大小避免内存问题
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	title := ""
	if resp.StatusCode == 200 {
		if m := titleRe.FindStringSubmatch(string(body)); len(m) >= 2 {
			t := strings.TrimSpace(htmlUnescape(m[1]))
			// 清理换行与多余空白
			t = strings.Join(strings.Fields(t), " ")
			title = t
		}
	}
	return resp.StatusCode, title
}

func htmlUnescape(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
	)
	return replacer.Replace(s)
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, v := range in {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
