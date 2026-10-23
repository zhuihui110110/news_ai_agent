// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: Server 组装(参照 stock_ai server.go PrepareRun 模式):
// 初始化 MySQL 工厂 / GLM 客户端 / 新闻检索渠道 / 公众号客户端, 注入依赖

package newsserver

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/factory"
	"news_ai_server/pkg/llm"
	"news_ai_server/internal/innerpkg/news_client"
	"news_ai_server/internal/innerpkg/wechat_client"
	appmw "news_ai_server/internal/newsserver/application"
	"news_ai_server/internal/newsserver/sconfig"
)

var (
	serverInstance *Server
	serverOnce     sync.Once
)

// Server 服务实例
type Server struct {
	cfg        *sconfig.Config
	factory    factory.Factory
	llmClient  *llm.Client
	providers  []news_client.Provider
	scrapeHTTP *http.Client
}

// GetServerOr 准备并获取服务单例
func GetServerOr(cfg *sconfig.Config) (*Server, error) {
	var err error
	serverOnce.Do(func() {
		s := &Server{cfg: cfg}
		if err = s.prepare(); err != nil {
			return
		}
		serverInstance = s
	})
	if err != nil {
		return nil, err
	}
	if serverInstance == nil {
		return nil, fmt.Errorf("server 未初始化")
	}
	return serverInstance, nil
}

// prepare 初始化各组件(顺序: DB -> LLM -> 检索渠道 -> 公众号)
func (s *Server) prepare() error {
	// 1. MySQL 工厂(含自动建表)
	f, err := factory.GetMySQLFactoryOr(s.cfg.MySQLOptions)
	if err != nil {
		return err
	}
	s.factory = f

	// 2. GLM 大模型客户端(直连)
	s.llmClient = llm.NewClient(s.cfg.LLMOptions)

	// 3. 新闻检索渠道(zhipu 渠道复用 GLM 客户端走智谱服务端联网搜索)
	for _, name := range s.cfg.NewsOptions.SearchProviders {
		p, e := news_client.NewProvider(name,
			&http.Client{Timeout: 15 * time.Second},
			s.cfg.NewsOptions.UserAgent,
			s.llmClient,
			s.cfg.NewsOptions.SearchModel)
		if e != nil {
			return e
		}
		s.providers = append(s.providers, p)
	}

	// 4. 正文抓取 HTTP 客户端
	s.scrapeHTTP = &http.Client{Timeout: s.cfg.NewsOptions.FetchTimeout}

	// 5. 微信公众号客户端(enabled=false 时业务层走 mock)
	wechat_client.Init(s.cfg.WechatOptions)
	return nil
}

// Deps 输出后台任务依赖
func (s *Server) Deps() *appmw.Deps {
	return &appmw.Deps{
		Factory:     s.factory,
		LLM:         s.llmClient,
		Providers:   s.providers,
		ScrapeHTTP:  s.scrapeHTTP,
		NewsOpts:    s.cfg.NewsOptions,
		TaskOpts:    s.cfg.TaskOptions,
		WechatOpts:  s.cfg.WechatOptions,
		Wechat:      wechat_client.GetClient(),
	}
}

// Factory 暴露工厂(给 service 层)
func (s *Server) Factory() factory.Factory { return s.factory }
