// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 后台任务(application 层)依赖集合, 参照 stock_ai application Monitor 的注入方式

package application

import (
	"net/http"

	"news_ai_server/internal/domain/infrastructure/mysql/factory"
	"news_ai_server/pkg/llm"
	"news_ai_server/internal/innerpkg/news_client"
	"news_ai_server/internal/innerpkg/options"
	"news_ai_server/internal/innerpkg/wechat_client"
)

// Deps 后台任务依赖
type Deps struct {
	Factory     factory.Factory            // MySQL 仓储工厂
	LLM         *llm.Client                // GLM 大模型客户端(直连)
	Providers   []news_client.Provider     // 新闻检索渠道
	ScrapeHTTP  *http.Client               // 正文抓取客户端
	Wechat      *wechat_client.Client      // 公众号客户端
	NewsOpts    *options.NewsOptions       // 新闻配置
	TaskOpts    *options.TaskOptions       // 任务配置
	WechatOpts  *options.WechatOptions     // 公众号配置
}
