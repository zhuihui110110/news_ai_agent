// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: newsserver 配置聚合(参照 stock_ai soptions.Options, 聚合各组件配置)

package soptions

import (
	"fmt"
	"strings"

	"news_ai_server/internal/innerpkg/options"
)

// Options newsserver 全量配置
type Options struct {
	*options.MySQLOptions    `json:"mysql" mapstructure:"mysql"`
	*options.LLMOptions      `json:"llm" mapstructure:"llm"`
	*options.NewsOptions     `json:"news" mapstructure:"news"`
	*options.WechatOptions   `json:"wechat" mapstructure:"wechat"`
	*options.TaskOptions     `json:"task" mapstructure:"task"`
	*options.LogOptions      `json:"log" mapstructure:"log"`
	*options.ServingOptions  `json:"serving" mapstructure:"serving"`
}

// NewOptions 默认配置
func NewOptions() *Options {
	return &Options{
		MySQLOptions:   options.NewMySQLOptions(),
		LLMOptions:     options.NewLLMOptions(),
		NewsOptions:    options.NewNewsOptions(),
		WechatOptions:  options.NewWechatOptions(),
		TaskOptions:    options.NewTaskOptions(),
		LogOptions:     options.NewLogOptions(),
		ServingOptions: options.NewServingOptions(),
	}
}

// Complete 配置补全
func (o *Options) Complete() error {
	if err := o.LLMOptions.Complete(); err != nil {
		return err
	}
	return nil
}

// Validate 配置校验(聚合各组件)
func (o *Options) Validate() []error {
	errs := make([]error, 0)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.LLMOptions.Validate()...)
	errs = append(errs, o.NewsOptions.Validate()...)
	errs = append(errs, o.WechatOptions.Validate()...)
	errs = append(errs, o.TaskOptions.Validate()...)
	errs = append(errs, o.LogOptions.Validate()...)
	errs = append(errs, o.ServingOptions.Validate()...)
	return errs
}

// String 配置摘要(脱敏)
func (o *Options) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "mysql=[%s:%d/%s] ", o.MySQLOptions.Host, o.MySQLOptions.Port, o.MySQLOptions.Database)
	fmt.Fprintf(sb, "llm=[%s model=%s apiKeySet=%v] ", o.LLMOptions.BaseURL, o.LLMOptions.Model, o.LLMOptions.APIKey != "")
	fmt.Fprintf(sb, "providers=%v ", o.NewsOptions.SearchProviders)
	fmt.Fprintf(sb, "wechatEnabled=%v ", o.WechatOptions.Enabled)
	fmt.Fprintf(sb, "addr=%s", o.ServingOptions.Address())
	return sb.String()
}
