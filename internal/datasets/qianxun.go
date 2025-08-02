package datasets

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Qianxun Qianxun 数据集模块
type Qianxun struct {
	*core.Query
	baseURL string
}

// NewQianxun 创建 Qianxun 数据集模块
func NewQianxun(cfg *config.Config) *Qianxun {
	return &Qianxun{
		Query:   core.NewQuery("QianXunQuery", cfg),
		baseURL: "https://www.dnsscan.cn/dns.html",
	}
}

// Run 执行查询
func (q *Qianxun) Run(domain string) ([]string, error) {
	q.SetDomain(domain)
	q.Begin()
	defer q.Finish()

	// 执行查询
	if err := q.query(domain); err != nil {
		return nil, err
	}

	return q.GetSubdomains(), nil
}

// query 执行查询
func (q *Qianxun) query(domain string) error {
	num := 1
	for {
		// 设置请求头
		q.SetHeader("User-Agent", q.GetRandomUserAgent())

		// 构建 POST 数据
		data := url.Values{}
		data.Set("ecmsfrom", "")
		data.Set("show", "")
		data.Set("num", "")
		data.Set("classid", "0")
		data.Set("keywords", domain)

		// 构建查询 URL
		queryURL := fmt.Sprintf("%s?keywords=%s&page=%d", q.baseURL, domain, num)

		// 发送 POST 请求
		resp, err := q.HTTPPost(queryURL, data, q.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query Qianxun: %v", err)
		}

		// 读取响应
		body, err := q.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := q.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break // 没有发现子域名则停止查询
		}

		for _, subdomain := range subdomains {
			q.AddSubdomain(subdomain)
		}

		// 检查是否有分页
		if !strings.Contains(body, `<div id="page" class="pagelist">`) {
			break
		}

		// 检查是否到达最后一页
		if strings.Contains(body, `<li class="disabled"><span>&raquo;</span></li>`) {
			break
		}

		num++
	}

	return nil
}
