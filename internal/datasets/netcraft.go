package datasets

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Netcraft Netcraft 数据集模块
type Netcraft struct {
	*core.Query
	baseURL    string
	pageNum    int
	perPageNum int
	delay      time.Duration
}

// NewNetcraft 创建 Netcraft 数据集模块
func NewNetcraft(cfg *config.Config) *Netcraft {
	return &Netcraft{
		Query:      core.NewQuery("NetCraftQuery", cfg),
		baseURL:    "https://searchdns.netcraft.com/?restriction=site+contains&position=limited",
		pageNum:    1,
		perPageNum: 20,
		delay:      1 * time.Second,
	}
}

// Run 执行查询
func (n *Netcraft) Run(domain string) ([]string, error) {
	n.SetDomain(domain)
	n.Begin()
	defer n.Finish()

	// 执行查询
	if err := n.query(domain); err != nil {
		return nil, err
	}

	return n.GetSubdomains(), nil
}

// query 执行查询
func (n *Netcraft) query(domain string) error {
	last := ""
	for n.pageNum <= 500 {
		time.Sleep(n.delay)

		// 设置请求头
		n.SetHeader("User-Agent", n.GetRandomUserAgent())

		// 构建查询参数
		params := url.Values{}
		params.Set("host", "*."+domain)
		params.Set("from", fmt.Sprintf("%d", n.pageNum))

		// 发送 GET 请求
		queryURL := n.baseURL + last + "&" + params.Encode()
		resp, err := n.HTTPGet(queryURL, n.GetHeader())
		if err != nil {
			return fmt.Errorf("failed to query Netcraft: %v", err)
		}

		// 读取响应
		body, err := n.ReadResponseBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		// 提取子域名
		subdomains := n.ExtractSubdomains(body, domain)
		if len(subdomains) == 0 {
			break // 搜索没有发现子域名则停止搜索
		}

		for _, subdomain := range subdomains {
			n.AddSubdomain(subdomain)
		}

		// 检查是否有下一页
		if !strings.Contains(body, "Next Page") {
			break // 搜索页面没有出现下一页时停止搜索
		}

		// 提取 last 参数
		re := regexp.MustCompile(`&last=.*` + domain)
		matches := re.FindString(body)
		if matches == "" {
			break
		}
		last = matches

		n.pageNum += n.perPageNum
	}

	return nil
}
