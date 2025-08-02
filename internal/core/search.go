package core

import (
	"strings"

	"github.com/oneforall-go/internal/config"
)

// Search 搜索大类基础类（对应 Python 的 Search 基类）
type Search struct {
	*BaseModule
	pageNum         int
	perPageNum      int
	recursiveSearch bool
	recursiveTimes  int
	fullSearch      bool
}

// NewSearch 创建搜索基础类
func NewSearch(name string, cfg *config.Config) *Search {
	return &Search{
		BaseModule:      NewBaseModule(name, ModuleTypeSearch, cfg),
		pageNum:         0,
		perPageNum:      50,
		recursiveSearch: true, // 默认启用递归搜索
		recursiveTimes:  3,    // 默认递归3次
		fullSearch:      true, // 默认启用全搜索
	}
}

// Filter 生成搜索过滤语句
// 使用搜索引擎支持的-site:语法过滤掉搜索页面较多的子域以发现新域
func (s *Search) Filter(domain string, subdomains []string) []string {
	statementsList := []string{}

	// 获取常见子域名
	commonSubnames := []string{"www", "mail", "ftp", "localhost", "webmail", "smtp", "pop", "ns1", "webdisk", "ns2", "cpanel", "whm", "autodiscover", "autoconfig", "m", "imap", "test", "ns", "blog", "pop3", "dev", "www2", "admin", "forum", "news", "vpn", "ns3", "mail2", "remote", "mysql", "api", "ns4", "server", "new", "beta", "shop", "ftp2", "media", "www1", "secure", "support", "static", "cdn", "mta", "ns5", "web", "mx", "email", "images", "img", "download", "dns1", "dns2", "portal", "ns6", "dns", "dns3", "dns4", "dns5", "dns6", "dns7", "dns8", "dns9", "dns10", "dns11", "dns12", "dns13", "dns14", "dns15", "dns16", "dns17", "dns18", "dns19", "dns20", "dns21", "dns22", "dns23", "dns24", "dns25", "dns26", "dns27", "dns28", "dns29", "dns30", "dns31", "dns32", "dns33", "dns34", "dns35", "dns36", "dns37", "dns38", "dns39", "dns40", "dns41", "dns42", "dns43", "dns44", "dns45", "dns46", "dns47", "dns48", "dns49", "dns50", "dns51", "dns52", "dns53", "dns54", "dns55", "dns56", "dns57", "dns58", "dns59", "dns60", "dns61", "dns62", "dns63", "dns64", "dns65", "dns66", "dns67", "dns68", "dns69", "dns70", "dns71", "dns72", "dns73", "dns74", "dns75", "dns76", "dns77", "dns78", "dns79", "dns80", "dns81", "dns82", "dns83", "dns84", "dns85", "dns86", "dns87", "dns88", "dns89", "dns90", "dns91", "dns92", "dns93", "dns94", "dns95", "dns96", "dns97", "dns98", "dns99", "dns100"}
	subdomainsTemp := make(map[string]bool)

	// 构建常见子域名集合
	for _, subname := range commonSubnames {
		fullSubdomain := subname + "." + domain
		for _, subdomain := range subdomains {
			if subdomain == fullSubdomain {
				subdomainsTemp[fullSubdomain] = true
			}
		}
	}

	// 转换为切片
	var tempList []string
	for subdomain := range subdomainsTemp {
		tempList = append(tempList, subdomain)
	}

	// 生成过滤语句
	for i := 0; i < len(tempList); i += 2 {
		var filters []string
		end := i + 2
		if end > len(tempList) {
			end = len(tempList)
		}

		for j := i; j < end; j++ {
			filters = append(filters, " -site:"+tempList[j])
		}

		statementsList = append(statementsList, strings.Join(filters, ""))
	}

	return statementsList
}

// MatchLocation 匹配跳转之后的url
// 针对部分搜索引擎(如百度搜索)搜索展示url时有显示不全的情况
// 此函数会向每条结果的链接发送head请求获取响应头的location值并做子域匹配
func (s *Search) MatchLocation(url string) []string {
	resp, err := s.HTTPGet(url, map[string]string{
		"User-Agent": s.GetRandomUserAgent(),
	})
	if err != nil {
		return []string{}
	}

	location := resp.Header.Get("location")
	if location == "" {
		return []string{}
	}

	return s.ExtractSubdomains(location, s.domain)
}

// CheckSubdomains 检查搜索出的子域结果是否满足条件
func (s *Search) CheckSubdomains(newSubdomains []string) bool {
	if len(newSubdomains) == 0 {
		// 搜索没有发现子域名则停止搜索
		return false
	}

	if !s.fullSearch {
		// 在全搜索过程中发现搜索出的结果有完全重复的结果就停止搜索
		existingSubdomains := s.GetSubdomains()
		allRepeated := true
		for _, newSubdomain := range newSubdomains {
			found := false
			for _, existingSubdomain := range existingSubdomains {
				if newSubdomain == existingSubdomain {
					found = true
					break
				}
			}
			if !found {
				allRepeated = false
				break
			}
		}
		if allRepeated {
			return false
		}
	}

	return true
}

// RecursiveSubdomain 递归搜索下一层的子域
func (s *Search) RecursiveSubdomain() []string {
	var recursiveSubdomains []string
	subdomains := s.GetSubdomains()

	// 从1开始是之前已经做过1层子域搜索了,当前实际递归层数是layer+1
	for layerNum := 1; layerNum < s.recursiveTimes; layerNum++ {
		for _, subdomain := range subdomains {
			// 进行下一层子域搜索的限制条件
			count := strings.Count(subdomain, ".") - strings.Count(s.domain, ".")
			if count == layerNum {
				recursiveSubdomains = append(recursiveSubdomains, subdomain)
			}
		}
	}

	return recursiveSubdomains
}

// SetPageNum 设置页码
func (s *Search) SetPageNum(pageNum int) {
	s.pageNum = pageNum
}

// GetPageNum 获取页码
func (s *Search) GetPageNum() int {
	return s.pageNum
}

// SetPerPageNum 设置每页数量
func (s *Search) SetPerPageNum(perPageNum int) {
	s.perPageNum = perPageNum
}

// GetPerPageNum 获取每页数量
func (s *Search) GetPerPageNum() int {
	return s.perPageNum
}

// SetRecursiveSearch 设置递归搜索
func (s *Search) SetRecursiveSearch(recursiveSearch bool) {
	s.recursiveSearch = recursiveSearch
}

// IsRecursiveSearch 是否递归搜索
func (s *Search) IsRecursiveSearch() bool {
	return s.recursiveSearch
}

// SetRecursiveTimes 设置递归次数
func (s *Search) SetRecursiveTimes(recursiveTimes int) {
	s.recursiveTimes = recursiveTimes
}

// GetRecursiveTimes 获取递归次数
func (s *Search) GetRecursiveTimes() int {
	return s.recursiveTimes
}

// SetFullSearch 设置全搜索
func (s *Search) SetFullSearch(fullSearch bool) {
	s.fullSearch = fullSearch
}

// IsFullSearch 是否全搜索
func (s *Search) IsFullSearch() bool {
	return s.fullSearch
}
