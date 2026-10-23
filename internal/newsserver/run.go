// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 服务 Run(参照 stock_ai run.go: 信号驱动优雅退出,
// 并行启动 REST API 与后台任务 Monitor)

package newsserver

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"

	"news_ai_server/internal/innerpkg/server"
	"news_ai_server/internal/innerpkg/wechat_client"
	"news_ai_server/internal/newsserver/app/service"
	v1 "news_ai_server/internal/newsserver/controller/v1"
	"news_ai_server/internal/newsserver/sconfig"
	"news_ai_server/pkg/log"
)

// Run 启动服务直到收到退出信号
func Run(cfg *sconfig.Config) error {
	svr, err := GetServerOr(cfg)
	if err != nil {
		return fmt.Errorf("初始化服务失败: %w", err)
	}

	// 信号驱动的 ctx, 收到 SIGINT/SIGTERM 时取消
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	apiServer := server.NewGenericAPIServer(cfg.ServingOptions.Address())
	svc := service.NewService(svr.Factory(), wechat_client.GetClient(),
		cfg.WechatOptions, cfg.NewsOptions)
	apiServer.AddWebService(v1.NewAPIWebService(svc))

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// REST API
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiServer.Start(ctx); err != nil {
			errCh <- fmt.Errorf("API 服务异常: %w", err)
		}
	}()

	// 后台任务 Monitor(检索/生成/草稿流水线 + 发布轮询)
	wg.Add(1)
	go func() {
		defer wg.Done()
		StartTasks(ctx, svr.Deps())
	}()

	log.Infof("newsserver started: %s", cfg.Options.String())
	<-ctx.Done()
	log.Infof("收到退出信号, 等待任务收尾...")
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}
	return nil
}
