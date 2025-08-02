package datasets

import (
	"fmt"
	"net/url"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Riddler Riddler 数据集模块
type Riddler struct {
	*core.Query
	baseURL string
}

// NewRiddler 创建 Riddler 数据集模块
func NewRiddler(cfg *config.Config) *Riddler {
	return &Riddler{
		Query:   core.NewQuery("RiddlerQuery", cfg),
		baseURL: "https://riddler.io/search",
	}
}

// Run 执行查询
func (r *Riddler) Run(domain string) ([]string, error) {
	r.SetDomain(domain)
	r.Begin()
	defer r.Finish()

	// 执行查询
	if err := r.query(domain); err != nil {
		return nil, err
	}

	return r.GetSubdomains(), nil
}

// query 执行查询
func (r *Riddler) query(domain string) error {
	// 设置请求头
	r.SetHeader("User-Agent", r.GetRandomUserAgent())

	// 构建查询参数
	params := url.Values{}
	params.Set("q", "pld:"+domain)

	// 发送 GET 请求
	queryURL := fmt.Sprintf("%s?%s", r.baseURL, params.Encode())
	resp, err := r.HTTPGet(queryURL, r.GetHeader())
	if err != nil {
		return fmt.Errorf("failed to query Riddler: %v", err)
	}

	// 读取响应
	body, err := r.ReadResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 提取子域名
	subdomains := r.ExtractSubdomains(body, domain)
	for _, subdomain := range subdomains {
		r.AddSubdomain(subdomain)
	}

	return nil
}
