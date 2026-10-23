// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 公众号草稿 Repository 实现(按 task_id upsert)

package factory

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/internal/domain/infrastructure/mysql/repository"
)

// 编译期校验
var _ repository.NewsDraftRepository = (*newsDraftRepository)(nil)

type newsDraftRepository struct {
	db *gorm.DB
}

var onConflictNewsDraft = clause.OnConflict{
	Columns:   []clause.Column{{Name: "task_id"}},
	DoUpdates: clause.AssignmentColumns([]string{"article_id", "media_id", "publish_id", "publish_status", "article_url", "err_msg", "is_mock", "updated_at"}),
}

func (r *newsDraftRepository) Upsert(ctx context.Context, m *model.NewsDraftModel) error {
	return dbWithCtx(ctx, r.db).Clauses(onConflictNewsDraft).Create(m).Error
}

func (r *newsDraftRepository) GetByID(ctx context.Context, id int64) (*model.NewsDraftModel, error) {
	var m model.NewsDraftModel
	if err := dbWithCtx(ctx, r.db).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *newsDraftRepository) GetByTaskID(ctx context.Context, taskID int64) (*model.NewsDraftModel, error) {
	var m model.NewsDraftModel
	if err := dbWithCtx(ctx, r.db).Where("task_id = ?", taskID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// ListByPublishStatus 按发布状态查询(status<=0 表示不过滤)
func (r *newsDraftRepository) ListByPublishStatus(ctx context.Context, status int32, limit int) ([]model.NewsDraftModel, error) {
	ms := make([]model.NewsDraftModel, 0)
	q := dbWithCtx(ctx, r.db)
	if status > 0 {
		q = q.Where("publish_status = ?", status)
	}
	err := q.Order("id ASC").Limit(limit).Find(&ms).Error
	return ms, err
}

func (r *newsDraftRepository) Update(ctx context.Context, m *model.NewsDraftModel) error {
	return dbWithCtx(ctx, r.db).Model(&model.NewsDraftModel{}).
		Where("id = ?", m.ID).Updates(m).Error
}
