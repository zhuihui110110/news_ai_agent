// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻源文章 Repository 接口

package repository

import (
	"context"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
)

// NewsSourceArticleRepository 新闻源文章仓储
type NewsSourceArticleRepository interface {
	// BatchUpsertIgnore 批量插入并忽略 url_md5 重复, 返回实际新增条数(跨任务去重)
	BatchUpsertIgnore(ctx context.Context, ms []model.NewsSourceArticleModel) (int64, error)
	ListByTaskID(ctx context.Context, taskID int64) ([]model.NewsSourceArticleModel, error)
	// UpdateContent 回写抓取正文
	UpdateContent(ctx context.Context, id int64, content string, contentLen, contentStatus int32) error
	// MarkSelected 标记入选生成素材
	MarkSelected(ctx context.Context, taskID int64, ids []int64) error
}
