// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: AI 生成文章 Repository 实现(按 task_id upsert, 一条任务一条文章)

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
var _ repository.NewsArticleRepository = (*newsArticleRepository)(nil)

type newsArticleRepository struct {
	db *gorm.DB
}

var onConflictNewsArticle = clause.OnConflict{
	Columns:   []clause.Column{{Name: "task_id"}},
	DoUpdates: clause.AssignmentColumns([]string{"title", "digest", "content", "tags", "cover_url", "llm_model", "updated_at"}),
}

func (r *newsArticleRepository) Upsert(ctx context.Context, m *model.NewsArticleModel) error {
	return dbWithCtx(ctx, r.db).Clauses(onConflictNewsArticle).Create(m).Error
}

func (r *newsArticleRepository) GetByTaskID(ctx context.Context, taskID int64) (*model.NewsArticleModel, error) {
	var m model.NewsArticleModel
	if err := dbWithCtx(ctx, r.db).Where("task_id = ?", taskID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *newsArticleRepository) GetByID(ctx context.Context, id int64) (*model.NewsArticleModel, error) {
	var m model.NewsArticleModel
	if err := dbWithCtx(ctx, r.db).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}
