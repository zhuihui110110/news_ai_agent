// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻任务 Repository 接口(实现见 mysql/factory)

package repository

import (
	"context"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
)

// NewsTaskRepository 新闻任务仓储
type NewsTaskRepository interface {
	Create(ctx context.Context, m *model.NewsTaskModel) error
	GetByID(ctx context.Context, id int64) (*model.NewsTaskModel, error)
	ListByStatus(ctx context.Context, status int32, limit int) ([]model.NewsTaskModel, error)
	ListPage(ctx context.Context, topic string, status *int32, offset, limit int) ([]model.NewsTaskModel, int64, error)
	// TransStatus 状态机 CAS 流转: UPDATE ... WHERE id=? AND task_status=?, 返回是否成功
	TransStatus(ctx context.Context, id int64, from, to int32, stageMsg, errMsg string) (bool, error)
	Update(ctx context.Context, m *model.NewsTaskModel) error
	UpdateRetryCount(ctx context.Context, id int64, retryCount int32) error
	UpdateLLMUsage(ctx context.Context, id int64, llmModel string, promptTokens, completionTokens int32) error
	// CountTopicSince 统计某领域自给定时间以来的任务数(自动建任务冷却用)
	CountTopicSince(ctx context.Context, topic string, since time.Time) (int64, error)
}
