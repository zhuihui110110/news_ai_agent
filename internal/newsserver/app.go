// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: newsserver 应用入口: 新闻 AI Agent 服务
// (新闻检索 -> 正文抓取 -> GLM 生成 -> 公众号草稿 -> 人工确认发布)

package newsserver

import (
	"news_ai_server/internal/newsserver/sconfig"
	"news_ai_server/internal/newsserver/soptions"
	"news_ai_server/pkg/app"
	"news_ai_server/pkg/log"
)

// NewApp 构造 newsserver 应用
func NewApp(name string) *app.App {
	opts := soptions.NewOptions()
	return app.NewApp(
		name,
		app.WithOptions(opts),
		app.WithDescription("新闻 AI Agent: 检索->抓取->GLM生成->公众号草稿->人工确认发布"),
		app.WithRunFunc(run(opts)),
	)
}

// run 应用运行函数: 初始化日志后进入服务 Run
func run(opts *soptions.Options) app.RunFunc {
	return func(basename string) error {
		log.Init(opts.LogOptions.Level, opts.LogOptions.Director)
		defer log.Sync()

		cfg, err := sconfig.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}
		return Run(cfg)
	}
}
