// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 配置包装(参照 stock_ai sconfig.CreateConfigFromOptions)

package sconfig

import (
	"fmt"

	"news_ai_server/internal/newsserver/soptions"
)

// Config 服务配置
type Config struct {
	*soptions.Options
}

// CreateConfigFromOptions 由 Options 构造 Config
func CreateConfigFromOptions(opts *soptions.Options) (*Config, error) {
	if opts == nil {
		return nil, fmt.Errorf("options 为空")
	}
	return &Config{Options: opts}, nil
}
