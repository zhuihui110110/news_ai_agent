// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 大模型直连配置(智谱 GLM, OpenAI 兼容 /chat/completions 协议)

package options

import (
	"fmt"
	"os"
	"time"
)

// LLMOptions 大模型配置项
type LLMOptions struct {
	BaseURL     string        `mapstructure:"base-url"`
	APIKey      string        `mapstructure:"api-key"`
	Model       string        `mapstructure:"model"`
	Timeout     time.Duration `mapstructure:"timeout"`
	Temperature float64       `mapstructure:"temperature"`
	MaxTokens   int           `mapstructure:"max-tokens"`
}

// NewLLMOptions 默认值(模型: GLM 5.3)
func NewLLMOptions() *LLMOptions {
	return &LLMOptions{
		BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
		Model:       "glm-5.3",
		Timeout:     120 * time.Second,
		Temperature: 0.7,
		MaxTokens:   16384,
	}
}

// Complete 补全: api-key 允许通过环境变量 LLM_API_KEY 注入(优先级低于配置文件)
func (o *LLMOptions) Complete() error {
	if o.APIKey == "" {
		o.APIKey = os.Getenv("LLM_API_KEY")
	}
	return nil
}

// Validate 参数校验
func (o *LLMOptions) Validate() []error {
	errs := make([]error, 0)
	if o.BaseURL == "" {
		errs = append(errs, fmt.Errorf("llm.base-url 不能为空"))
	}
	if o.Model == "" {
		errs = append(errs, fmt.Errorf("llm.model 不能为空"))
	}
	return errs
}
