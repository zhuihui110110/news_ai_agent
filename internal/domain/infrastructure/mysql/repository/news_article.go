// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: AI 生成文章 Repository 接口

package repository

import (
	"context"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
)

// NewsArticleRepository AI 生成文章仓储
type NewsArticleRepository interface {
	Upsert(ctx context.Context, m *model.NewsArticleModel) error
	GetByTaskID(ctx context.Context, taskID int64) (*model.NewsArticleModel, error)
	GetByID(ctx context.Context, id int64) (*model.NewsArticleModel, error)
}
