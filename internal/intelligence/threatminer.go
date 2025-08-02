package intelligence

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// ThreatMiner ThreatMiner 情报模块
type ThreatMiner struct {
	*core.Query
	baseURL string
}

// NewThreatMiner 创建 ThreatMiner 情报模块
func NewThreatMiner(cfg *config.Config) *ThreatMiner {
	return &ThreatMiner{
		Query:   core.NewQuery("ThreatMinerQuery", cfg),
		baseURL: "https://api.threatminer.org/v2/domain.php",
	}
}

// Run 执行查询
func (t *ThreatMiner) Run(domain string) ([]string, error) {
	t.SetDomain(domain)
	t.Begin()
	defer t.Finish()

	// 执行查询
	if err := t.query(domain); err != nil {
		return nil, err
	}

	return t.GetSubdomains(), nil
}

// query 执行查询
func (t *ThreatMiner) query(domain string) error {
	// 设置请求头
	t.SetHeader("User-Agent", t.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("q", domain)
	params.Set("rt", "5")

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", t.baseURL, params.Encode())
	resp, err := t.HTTPGet(queryURL, t.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query ThreatMiner: %v", err)
	}

	// 读取响应
	body, err := t.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := t.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		t.AddSubdomain(subdomain)
	}

	return nil
}
