// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 微信公众号配置(未启用时草稿走 mock, 便于本地跑通流程)

package options

import (
	"fmt"
	"time"
)

// WechatOptions 公众号配置项
type WechatOptions struct {
	Enabled          bool          `mapstructure:"enabled"`
	AppID            string        `mapstructure:"app-id"`
	AppSecret        string        `mapstructure:"app-secret"`
	BaseURL          string        `mapstructure:"base-url"`
	DefaultCoverPath string        `mapstructure:"default-cover-path"`
	Timeout          time.Duration `mapstructure:"timeout"`
}

// NewWechatOptions 默认值
func NewWechatOptions() *WechatOptions {
	return &WechatOptions{
		BaseURL:          "https://api.weixin.qq.com",
		DefaultCoverPath: "./configs/assets/default_cover.jpg",
		Timeout:          30 * time.Second,
	}
}

// Validate 参数校验
func (o *WechatOptions) Validate() []error {
	errs := make([]error, 0)
	if o.Enabled {
		if o.AppID == "" || o.AppSecret == "" {
			errs = append(errs, fmt.Errorf("wechat.enabled=true 时 app-id/app-secret 不能为空"))
		}
	}
	if o.BaseURL == "" {
		errs = append(errs, fmt.Errorf("wechat.base-url 不能为空"))
	}
	return errs
}
