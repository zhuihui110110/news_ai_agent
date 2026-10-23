// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: REST API 消息定义(参照 stock_ai app/message 的 XxxReq/XxxResp 风格)

package message

import (
	"news_ai_server/internal/domain/infrastructure/mysql/model"
)

// CreateTaskReq 创建任务请求
type CreateTaskReq struct {
	Topics   []string `json:"topics"`   // 领域列表: livelihood/economy/ai/tech, 空则默认全部
	Keywords []string `json:"keywords"` // 可选: 覆盖默认关键词
}

// CreateTaskResp 创建任务响应
type CreateTaskResp struct {
	Tasks []*model.NewsTaskModel `json:"tasks"`
}

// TaskListReq 任务列表请求
type TaskListReq struct {
	Topic    string `json:"topic"`     // 领域筛选
	Status   *int32 `json:"status"`    // 状态筛选
	Page     int    `json:"page"`      // 从1开始
	PageSize int    `json:"page_size"` // 默认20
}

// TaskListResp 任务列表响应
type TaskListResp struct {
	Total int64                  `json:"total"`
	List  []model.NewsTaskModel  `json:"list"`
}

// TaskDetailResp 任务详情响应(任务 + 源新闻 + 文章 + 草稿)
type TaskDetailResp struct {
	Task     *model.NewsTaskModel          `json:"task"`
	Articles []model.NewsSourceArticleModel `json:"articles"`
	Article  *model.NewsArticleModel       `json:"generated"`
	Draft    *model.NewsDraftModel         `json:"draft"`
}

// TaskIDReq 任务ID请求(重试/取消/详情)
type TaskIDReq struct {
	ID int64 `json:"id"`
}

// DraftListReq 草稿列表请求
type DraftListReq struct {
	PublishStatus *int32 `json:"publish_status"`
	Limit         int    `json:"limit"`
}

// DraftListResp 草稿列表响应(带任务与文章摘要便于人工审核)
type DraftListResp struct {
	List []DraftItem `json:"list"`
}

// DraftItem 草稿条目
type DraftItem struct {
	Draft  model.NewsDraftModel   `json:"draft"`
	Task   model.NewsTaskModel    `json:"task"`
	Title  string                 `json:"title"`  // 文章标题
	Digest string                 `json:"digest"` // 文章摘要
}

// DraftIDReq 草稿ID请求(确认发布)
type DraftIDReq struct {
	ID int64 `json:"id"`
}

// PublishResp 确认发布响应
type PublishResp struct {
	PublishID string `json:"publish_id"`
	Mock      bool   `json:"mock"`
}
