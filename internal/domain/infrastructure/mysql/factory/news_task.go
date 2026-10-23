// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻任务 Repository 实现(核心: TransStatus 状态机 CAS 流转)

package factory

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/internal/domain/infrastructure/mysql/repository"
)

// 编译期校验: 确保 newsTaskRepository 实现了 NewsTaskRepository 接口
var _ repository.NewsTaskRepository = (*newsTaskRepository)(nil)

type newsTaskRepository struct {
	db *gorm.DB
}

func (r *newsTaskRepository) Create(ctx context.Context, m *model.NewsTaskModel) error {
	return dbWithCtx(ctx, r.db).Create(m).Error
}

func (r *newsTaskRepository) GetByID(ctx context.Context, id int64) (*model.NewsTaskModel, error) {
	var m model.NewsTaskModel
	if err := dbWithCtx(ctx, r.db).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *newsTaskRepository) ListByStatus(ctx context.Context, status int32, limit int) ([]model.NewsTaskModel, error) {
	ms := make([]model.NewsTaskModel, 0)
	err := dbWithCtx(ctx, r.db).
		Where("task_status = ?", status).
		Order("id ASC").Limit(limit).Find(&ms).Error
	return ms, err
}

func (r *newsTaskRepository) ListPage(ctx context.Context, topic string, status *int32, offset, limit int) ([]model.NewsTaskModel, int64, error) {
	ms := make([]model.NewsTaskModel, 0)
	q := dbWithCtx(ctx, r.db).Model(&model.NewsTaskModel{})
	if topic != "" {
		q = q.Where("topic = ?", topic)
	}
	if status != nil {
		q = q.Where("task_status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&ms).Error
	return ms, total, err
}

// TransStatus 状态机 CAS 流转: 仅当当前状态==from 时更新为 to, 返回是否流转成功
func (r *newsTaskRepository) TransStatus(ctx context.Context, id int64, from, to int32, stageMsg, errMsg string) (bool, error) {
	updates := map[string]interface{}{
		"task_status": to,
		"stage_msg":   stageMsg,
		"err_msg":     errMsg,
	}
	res := dbWithCtx(ctx, r.db).Model(&model.NewsTaskModel{}).
		Where("id = ? AND task_status = ?", id, from).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *newsTaskRepository) Update(ctx context.Context, m *model.NewsTaskModel) error {
	return dbWithCtx(ctx, r.db).Model(&model.NewsTaskModel{}).
		Where("id = ?", m.ID).Updates(m).Error
}

func (r *newsTaskRepository) UpdateLLMUsage(ctx context.Context, id int64, llmModel string, promptTokens, completionTokens int32) error {
	return dbWithCtx(ctx, r.db).Model(&model.NewsTaskModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"llm_model":             llmModel,
			"llm_prompt_tokens":     promptTokens,
			"llm_completion_tokens": completionTokens,
		}).Error
}

func (r *newsTaskRepository) UpdateRetryCount(ctx context.Context, id int64, retryCount int32) error {
	return dbWithCtx(ctx, r.db).Model(&model.NewsTaskModel{}).
		Where("id = ?", id).
		Update("retry_count", retryCount).Error
}

func (r *newsTaskRepository) CountTopicSince(ctx context.Context, topic string, since time.Time) (int64, error) {
	var n int64
	err := dbWithCtx(ctx, r.db).Model(&model.NewsTaskModel{}).
		Where("topic = ? AND created_at >= ?", topic, since).Count(&n).Error
	return n, err
}
