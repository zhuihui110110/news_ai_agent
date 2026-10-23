// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 公众号草稿 Repository 接口

package repository

import (
	"context"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
)

// NewsDraftRepository 公众号草稿仓储
type NewsDraftRepository interface {
	Upsert(ctx context.Context, m *model.NewsDraftModel) error
	GetByID(ctx context.Context, id int64) (*model.NewsDraftModel, error)
	GetByTaskID(ctx context.Context, taskID int64) (*model.NewsDraftModel, error)
	ListByPublishStatus(ctx context.Context, status int32, limit int) ([]model.NewsDraftModel, error)
	Update(ctx context.Context, m *model.NewsDraftModel) error
}
