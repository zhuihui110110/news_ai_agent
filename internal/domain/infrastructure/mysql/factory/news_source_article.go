// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻源文章 Repository 实现(BatchUpsertIgnore 基于 url_md5 冲突忽略实现跨任务去重)

package factory

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/internal/domain/infrastructure/mysql/repository"
)

// 编译期校验
var _ repository.NewsSourceArticleRepository = (*newsSourceArticleRepository)(nil)

type newsSourceArticleRepository struct {
	db *gorm.DB
}

// BatchUpsertIgnore 批量插入, url_md5 冲突时忽略(不更新), 返回实际新增条数
func (r *newsSourceArticleRepository) BatchUpsertIgnore(ctx context.Context, ms []model.NewsSourceArticleModel) (int64, error) {
	if len(ms) == 0 {
		return 0, nil
	}
	res := dbWithCtx(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(ms, 100)
	return res.RowsAffected, res.Error
}

func (r *newsSourceArticleRepository) ListByTaskID(ctx context.Context, taskID int64) ([]model.NewsSourceArticleModel, error) {
	ms := make([]model.NewsSourceArticleModel, 0)
	err := dbWithCtx(ctx, r.db).
		Where("task_id = ?", taskID).Order("id ASC").Find(&ms).Error
	return ms, err
}

func (r *newsSourceArticleRepository) UpdateContent(ctx context.Context, id int64, content string, contentLen, contentStatus int32) error {
	return dbWithCtx(ctx, r.db).Model(&model.NewsSourceArticleModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"content":        content,
			"content_len":    contentLen,
			"content_status": contentStatus,
		}).Error
}

func (r *newsSourceArticleRepository) MarkSelected(ctx context.Context, taskID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return dbWithCtx(ctx, r.db).Model(&model.NewsSourceArticleModel{}).
		Where("task_id = ? AND id IN ?", taskID, ids).
		Update("selected", 1).Error
}
