// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻检索与正文抓取配置

package options

import (
	"fmt"
	"time"
)

// NewsOptions 新闻配置项: Topics 为 领域->关键词列表, 支持的领域见 constants.TopicXxx
type NewsOptions struct {
	SearchProviders    []string          `mapstructure:"search-providers"`
	SearchModel        string            `mapstructure:"search-model"` // zhipu 渠道检索用的模型(建议轻量如 glm-4-flash)
	MaxArticlesPerTask int               `mapstructure:"max-articles-per-task"`
	FetchWorkers       int               `mapstructure:"fetch-workers"`
	FetchTimeout       time.Duration     `mapstructure:"fetch-timeout"`
	UserAgent          string            `mapstructure:"user-agent"`
	Topics             map[string][]string `mapstructure:"topics"`
}

// NewNewsOptions 默认值
func NewNewsOptions() *NewsOptions {
	return &NewsOptions{
		SearchProviders:    []string{"zhipu"},
		SearchModel:        "glm-4-flash",
		MaxArticlesPerTask: 12,
		FetchWorkers:       4,
		FetchTimeout:       15 * time.Second,
		UserAgent:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36",
		Topics: map[string][]string{
			"livelihood": {"民生保障", "就业", "教育医疗"},
			"economy":    {"宏观经济", "金融市场", "产业经济"},
			"ai":         {"AI大模型", "人工智能", "OpenAI"},
			"tech":       {"硬科技", "半导体", "智能硬件"},
		},
	}
}

// Validate 参数校验
func (o *NewsOptions) Validate() []error {
	errs := make([]error, 0)
	if len(o.SearchProviders) == 0 {
		errs = append(errs, fmt.Errorf("news.search-providers 不能为空"))
	}
	if o.MaxArticlesPerTask <= 0 {
		errs = append(errs, fmt.Errorf("news.max-articles-per-task 必须大于0"))
	}
	if o.FetchWorkers <= 0 {
		errs = append(errs, fmt.Errorf("news.fetch-workers 必须大于0"))
	}
	return errs
}
