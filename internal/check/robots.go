package check

import (
	"github.com/oneforall-go/internal/config"
	"github.com/oneforall-go/internal/core"
)

// Robots Robots 检查模块
type Robots struct {
	*core.Check
}

// NewRobots 创建 Robots 检查模块
func NewRobots(cfg *config.Config) *Robots {
	return &Robots{
		Check: core.NewCheck("RobotsCheck", cfg),
	}
}

// Run 执行检查
func (r *Robots) Run(domain string) ([]string, error) {
	r.SetDomain(domain)
	r.Begin()
	defer r.Finish()

	// 执行检查
	if err := r.check(domain); err != nil {
		return nil, err
	}

	return r.GetSubdomains(), nil
}

// check 执行检查
func (r *Robots) check(domain string) error {
	// 设置要检查的文件名
	filenames := []string{"robots.txt"}

	// 调用基类的 ToCheck 方法
	r.ToCheck(filenames)

	return nil
}
