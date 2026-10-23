// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻源文章模型(检索结果+正文抓取结果, url_md5 全局去重)

package model

import (
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
)

// NewsSourceArticleModel 新闻源文章表模型
type NewsSourceArticleModel struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	TaskID        int64      `gorm:"column:task_id;type:bigint;not null;default:0;index:idx_task_id;comment:所属任务id" json:"task_id"`
	Provider      string     `gorm:"column:provider;type:varchar(32);not null;default:'';comment:检索来源(google/bing)" json:"provider"`
	URL           string     `gorm:"column:url;type:varchar(1024);not null;default:'';comment:原文链接" json:"url"`
	URLMd5        string     `gorm:"column:url_md5;type:char(32);not null;default:'';uniqueIndex:uk_url_md5;comment:链接md5去重键" json:"url_md5"`
	Title         string     `gorm:"column:title;type:varchar(512);not null;default:'';comment:标题" json:"title"`
	Snippet       string     `gorm:"column:snippet;type:varchar(2048);not null;default:'';comment:摘要" json:"snippet"`
	Author        string     `gorm:"column:author;type:varchar(128);not null;default:'';comment:作者" json:"author"`
	SourceName    string     `gorm:"column:source_name;type:varchar(128);not null;default:'';comment:媒体名" json:"source_name"`
	PublishedAt   *time.Time `gorm:"column:published_at;type:datetime;comment:发布时间" json:"published_at"`
	Content       string     `gorm:"column:content;type:mediumtext;comment:抓取的正文纯文本" json:"content"`
	ContentStatus int32      `gorm:"column:content_status;type:tinyint;not null;default:0;comment:正文状态" json:"content_status"`
	ContentLen    int32      `gorm:"column:content_len;type:int;not null;default:0;comment:正文字数" json:"content_len"`
	Selected      int32      `gorm:"column:selected;type:tinyint;not null;default:0;comment:是否入选生成素材" json:"selected"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (m *NewsSourceArticleModel) TableName() string { return constants.NewsSourceArticleTableName }
