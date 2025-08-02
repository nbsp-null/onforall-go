package alt

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
	"github.com/oneforall-go/pkg/logger"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Alt Alt 模块
type Alt struct {
	*core.BaseModule
	domain            string
	words             map[string]bool
	nowSubdomains     map[string]bool
	newSubdomains     map[string]bool
	wordLen           int
	numCount          int
	enableIncreaseNum bool
	enableDecreaseNum bool
	enableReplaceWord bool
	enableInsertWord  bool
	enableAddWord     bool
}

// NewAlt 创建 Alt 模块
func NewAlt(cfg *config.Config) *Alt {
	return &Alt{
		BaseModule:        core.NewBaseModule("Alt", core.ModuleTypeBrute, cfg),
		words:             make(map[string]bool),
		nowSubdomains:     make(map[string]bool),
		newSubdomains:     make(map[string]bool),
		wordLen:           6,
		numCount:          3,
		enableIncreaseNum: true,
		enableDecreaseNum: true,
		enableReplaceWord: true,
		enableInsertWord:  true,
		enableAddWord:     true,
	}
}

// Run 执行 Alt 模块
func (a *Alt) Run(domain string) ([]string, error) {
	logger.Infof("=== Starting Alt module for domain: %s ===", domain)

	// 获取字典
	logger.Debugf("Loading altdns wordlist...")
	if err := a.getWords(); err != nil {
		logger.Errorf("Failed to load altdns wordlist: %v", err)
		return nil, err
	}
	logger.Debugf("Loaded %d words from altdns wordlist", len(a.words))

	// 加载现有子域名（这里需要从其他模块获取）
	// 暂时使用空列表，实际使用时需要从其他模块的结果中获取
	existingSubdomains := []string{}
	for _, subdomain := range existingSubdomains {
		a.nowSubdomains[subdomain] = true
	}
	logger.Debugf("Loaded %d existing subdomains", len(existingSubdomains))

	// 提取单词
	logger.Debugf("Extracting words from existing subdomains...")
	a.extractWords()
	logger.Debugf("Extracted %d unique words", len(a.words))

	// 生成新的子域名
	logger.Debugf("Generating new subdomains...")
	a.genNewSubdomains()
	logger.Debugf("Generated %d new subdomains", len(a.newSubdomains))

	// 提取新生成的子域名
	var newSubdomains []string
	for subdomain := range a.newSubdomains {
		if !a.nowSubdomains[subdomain] {
			newSubdomains = append(newSubdomains, subdomain)
		}
	}

	logger.Infof("Alt module completed: generated %d new subdomains", len(newSubdomains))
	if len(newSubdomains) > 0 {
		logger.Debugf("Sample new subdomains: %v", newSubdomains[:min(5, len(newSubdomains))])
	}

	return newSubdomains, nil
}

// getWords 获取字典
func (a *Alt) getWords() error {
	path := filepath.Join("data", "altdns_wordlist.txt")
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open altdns wordlist: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if word != "" {
			a.words[word] = true
		}
	}

	return nil
}

// extractWords 从现有子域名中提取单词
func (a *Alt) extractWords() {
	for subdomain := range a.nowSubdomains {
		subname, _ := a.splitDomain(subdomain)

		// 提取单词
		tokens := make(map[string]bool)

		// 分割子域名部分
		subnameParts := strings.Split(subname, ".")
		for _, part := range subnameParts {
			tokens[strings.ToLower(part)] = true

			// 按连字符分割
			hyphenParts := strings.Split(part, "-")
			for _, hyphenPart := range hyphenParts {
				if len(hyphenPart) >= a.wordLen {
					tokens[strings.ToLower(hyphenPart)] = true
				}
			}
		}

		// 添加到单词字典
		for token := range tokens {
			if len(token) >= a.wordLen {
				a.words[token] = true
			}
		}
	}
}

// splitDomain 分割域名
func (a *Alt) splitDomain(domain string) (string, []string) {
	// 移除主域名部分
	mainDomain := a.domain
	subname := strings.TrimSuffix(domain, "."+mainDomain)

	// 分割子域名部分
	parts := strings.Split(subname, ".")

	return subname, parts
}

// genNewSubdomains 生成新的子域名
func (a *Alt) genNewSubdomains() {
	for subdomain := range a.nowSubdomains {
		subname, parts := a.splitDomain(subdomain)
		subnames := strings.Split(subname, ".")

		if a.enableIncreaseNum {
			a.doIncreaseNum(subname)
		}
		if a.enableDecreaseNum {
			a.doDecreaseNum(subname)
		}
		if a.enableReplaceWord {
			a.doReplaceWord(subname)
		}
		if a.enableInsertWord {
			a.doInsertWord(parts)
		}
		if a.enableAddWord {
			a.doAddWord(subnames)
		}
	}
}

// doIncreaseNum 数字递增
func (a *Alt) doIncreaseNum(subname string) {
	digits := regexp.MustCompile(`\d{1,3}`).FindAllString(subname, -1)
	for _, d := range digits {
		for m := 0; m < a.numCount; m++ {
			replacement := fmt.Sprintf("%0*d", len(d), a.parseInt(d)+1+m)
			tmpDomain := strings.Replace(subname, d, replacement, 1)
			newDomain := fmt.Sprintf("%s.%s", tmpDomain, a.domain)
			a.newSubdomains[newDomain] = true
		}
	}
}

// doDecreaseNum 数字递减
func (a *Alt) doDecreaseNum(subname string) {
	digits := regexp.MustCompile(`\d{1,3}`).FindAllString(subname, -1)
	for _, d := range digits {
		for m := 0; m < a.numCount; m++ {
			newDigit := a.parseInt(d) - 1 - m
			if newDigit < 0 {
				break
			}
			replacement := fmt.Sprintf("%0*d", len(d), newDigit)
			tmpDomain := strings.Replace(subname, d, replacement, 1)
			newDomain := fmt.Sprintf("%s.%s", tmpDomain, a.domain)
			a.newSubdomains[newDomain] = true
		}
	}
}

// doReplaceWord 替换单词
func (a *Alt) doReplaceWord(subname string) {
	for word := range a.words {
		if !strings.Contains(subname, word) {
			continue
		}
		for wordAlt := range a.words {
			if word == wordAlt {
				continue
			}
			newSubname := strings.Replace(subname, word, wordAlt, -1)
			newSubdomain := fmt.Sprintf("%s.%s", newSubname, a.domain)
			a.newSubdomains[newSubdomain] = true
		}
	}
}

// doInsertWord 插入单词
func (a *Alt) doInsertWord(parts []string) {
	for word := range a.words {
		for index := range parts {
			tmpParts := make([]string, len(parts))
			copy(tmpParts, parts)

			// 在指定位置插入单词
			newParts := make([]string, 0, len(tmpParts)+1)
			newParts = append(newParts, tmpParts[:index]...)
			newParts = append(newParts, word)
			newParts = append(newParts, tmpParts[index:]...)

			newDomain := strings.Join(newParts, ".") + "." + a.domain
			a.newSubdomains[newDomain] = true
		}
	}
}

// doAddWord 添加单词
func (a *Alt) doAddWord(subnames []string) {
	for word := range a.words {
		for index, name := range subnames {
			// 前缀添加
			tmpSubnames := make([]string, len(subnames))
			copy(tmpSubnames, subnames)
			tmpSubnames[index] = fmt.Sprintf("%s-%s", word, name)
			newSubname := strings.Join(tmpSubnames, ".") + "." + a.domain
			a.newSubdomains[newSubname] = true

			// 后缀添加
			tmpSubnames = make([]string, len(subnames))
			copy(tmpSubnames, subnames)
			tmpSubnames[index] = fmt.Sprintf("%s-%s", name, word)
			newSubname = strings.Join(tmpSubnames, ".") + "." + a.domain
			a.newSubdomains[newSubname] = true
		}
	}
}

// parseInt 解析整数
func (a *Alt) parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
