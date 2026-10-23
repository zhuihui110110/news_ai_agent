// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻任务模型(任务状态机载体)

package model

import (
	"slices"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
)

// NewsTaskModel 新闻任务表模型, 记录一次"检索->生成->草稿->发布"全流程
type NewsTaskModel struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	TaskNo              string    `gorm:"column:task_no;type:varchar(64);not null;default:'';uniqueIndex:uk_task_no;comment:任务编号" json:"task_no"`
	Topic               string    `gorm:"column:topic;type:varchar(32);not null;default:'';index:idx_topic_status,priority:1;comment:新闻领域" json:"topic"`
	Keywords            string    `gorm:"column:keywords;type:varchar(512);not null;default:'';comment:检索关键词,逗号分隔" json:"keywords"`
	TaskStatus          int32     `gorm:"column:task_status;type:tinyint;not null;default:0;index:idx_status;index:idx_topic_status,priority:2;comment:任务状态" json:"task_status"`
	StageMsg            string    `gorm:"column:stage_msg;type:varchar(1024);not null;default:'';comment:当前阶段说明" json:"stage_msg"`
	ErrMsg              string    `gorm:"column:err_msg;type:varchar(1024);not null;default:'';comment:失败原因" json:"err_msg"`
	RetryCount          int32     `gorm:"column:retry_count;type:int;not null;default:0;comment:已重试次数" json:"retry_count"`
	LLMModel            string    `gorm:"column:llm_model;type:varchar(64);not null;default:'';comment:生成使用的模型" json:"llm_model"`
	LLMPromptTokens     int32     `gorm:"column:llm_prompt_tokens;type:int;not null;default:0;comment:提示词token" json:"llm_prompt_tokens"`
	LLMCompletionTokens int32     `gorm:"column:llm_completion_tokens;type:int;not null;default:0;comment:补全token" json:"llm_completion_tokens"`
	CreatedAt           time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName 表名
func (m *NewsTaskModel) TableName() string { return constants.NewsTaskTableName }

// CanTransTo 校验状态机流转合法性
func (m *NewsTaskModel) CanTransTo(to int32) bool {
	return slices.Contains(constants.TaskStatusTransitions[m.TaskStatus], to)
}
