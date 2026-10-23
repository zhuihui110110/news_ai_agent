// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 公众号草稿模型(草稿创建+发布状态跟踪)

package model

import (
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
)

// NewsDraftModel 公众号草稿表模型
type NewsDraftModel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	TaskID        int64     `gorm:"column:task_id;type:bigint;not null;default:0;uniqueIndex:uk_task_id;comment:所属任务id" json:"task_id"`
	ArticleID     int64     `gorm:"column:article_id;type:bigint;not null;default:0;comment:文章id" json:"article_id"`
	MediaID       string    `gorm:"column:media_id;type:varchar(128);not null;default:'';comment:草稿media_id" json:"media_id"`
	PublishID     string    `gorm:"column:publish_id;type:varchar(128);not null;default:'';comment:发布任务id" json:"publish_id"`
	PublishStatus int32     `gorm:"column:publish_status;type:tinyint;not null;default:0;index:idx_publish_status;comment:发布状态" json:"publish_status"`
	ArticleURL    string    `gorm:"column:article_url;type:varchar(1024);not null;default:'';comment:发布后文章链接" json:"article_url"`
	ErrMsg        string    `gorm:"column:err_msg;type:varchar(1024);not null;default:'';comment:失败原因" json:"err_msg"`
	IsMock        int32     `gorm:"column:is_mock;type:tinyint;not null;default:0;comment:1=mock草稿" json:"is_mock"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (m *NewsDraftModel) TableName() string { return constants.NewsDraftTableName }
