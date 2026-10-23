// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: AI 生成文章模型(GLM 产出, 一条任务一条)

package model

import (
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
)

// NewsArticleModel AI 生成文章表模型
type NewsArticleModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	TaskID    int64     `gorm:"column:task_id;type:bigint;not null;default:0;uniqueIndex:uk_task_id;comment:所属任务id" json:"task_id"`
	Title     string    `gorm:"column:title;type:varchar(256);not null;default:'';comment:文章标题" json:"title"`
	Digest    string    `gorm:"column:digest;type:varchar(512);not null;default:'';comment:摘要" json:"digest"`
	Content   string    `gorm:"column:content;type:mediumtext;comment:公众号正文HTML" json:"content"`
	Tags      string    `gorm:"column:tags;type:varchar(256);not null;default:'';comment:标签逗号分隔" json:"tags"`
	CoverURL  string    `gorm:"column:cover_url;type:varchar(1024);not null;default:'';comment:封面图URL" json:"cover_url"`
	LLMModel  string    `gorm:"column:llm_model;type:varchar(64);not null;default:'';comment:生成模型" json:"llm_model"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (m *NewsArticleModel) TableName() string { return constants.NewsArticleTableName }
