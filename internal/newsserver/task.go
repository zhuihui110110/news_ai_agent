// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 后台任务启动入口(参照 stock_ai taskserver/tasks.Start())

package newsserver

import (
	"context"

	appmw "news_ai_server/internal/newsserver/application"
	"news_ai_server/pkg/log"
)

// StartTasks 启动全部后台 Monitor(每个 Monitor 一个 goroutine)
func StartTasks(ctx context.Context, deps *appmw.Deps) {
	log.Infof("启动后台任务: 新闻任务流水线 + 草稿发布轮询")
	go appmw.StartNewsTaskMonitor(ctx, deps)
	go appmw.StartDraftPublishMonitor(ctx, deps)
}
