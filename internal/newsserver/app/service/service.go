// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 业务 Service 层(参照 stock_ai app/service): 面向 REST API 的任务/草稿管理

package service

import (
	"news_ai_server/internal/domain/infrastructure/mysql/factory"
	"news_ai_server/internal/innerpkg/options"
	"news_ai_server/internal/innerpkg/wechat_client"
)

// Service 业务服务
type Service struct {
	factory     factory.Factory
	wechat      *wechat_client.Client
	wechatOpts  *options.WechatOptions
	newsOpts    *options.NewsOptions
}

// NewService 构造 Service
func NewService(f factory.Factory, wc *wechat_client.Client,
	wechatOpts *options.WechatOptions, newsOpts *options.NewsOptions) *Service {
	return &Service{factory: f, wechat: wc, wechatOpts: wechatOpts, newsOpts: newsOpts}
}
