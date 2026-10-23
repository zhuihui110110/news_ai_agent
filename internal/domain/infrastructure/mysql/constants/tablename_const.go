// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: MySQL 表名与状态机常量(参照 stock_ai constants 风格)

package constants

// 表名
const (
	NewsTaskTableName         = "news_task"         // 新闻任务表
	NewsSourceArticleTableName = "news_source_article" // 新闻源文章表
	NewsArticleTableName      = "news_article"      // AI生成文章表
	NewsDraftTableName        = "news_draft"        // 公众号草稿表
)

// 新闻领域(新闻领域=民生、经济、人工智能、科技)
const (
	TopicLivelihood = "livelihood" // 民生
	TopicEconomy    = "economy"    // 经济
	TopicAI         = "ai"         // 人工智能
	TopicTech       = "tech"       // 科技
)

// AllTopics 全部领域
var AllTopics = []string{TopicLivelihood, TopicEconomy, TopicAI, TopicTech}

// TopicName 领域中文名
var TopicName = map[string]string{
	TopicLivelihood: "民生",
	TopicEconomy:    "经济",
	TopicAI:         "人工智能",
	TopicTech:       "科技",
}

// 任务状态机:
//
//	pending(0) -> searching(10) -> fetching(20) -> generating(30) -> draft_creating(40)
//	          -> awaiting_confirm(50) -> publishing(60) -> published(70)
//	任意执行态 -> failed(80) --重试--> pending(0)
//	任意未终态 -> cancelled(90)
const (
	TaskStatusPending         int32 = 0  // 待执行
	TaskStatusSearching       int32 = 10 // 搜索新闻中
	TaskStatusFetching        int32 = 20 // 抓取正文中
	TaskStatusGenerating      int32 = 30 // AI生成中
	TaskStatusDraftCreating   int32 = 40 // 创建草稿中
	TaskStatusAwaitingConfirm int32 = 50 // 待人工确认
	TaskStatusPublishing      int32 = 60 // 发布中
	TaskStatusPublished       int32 = 70 // 已发布(终态)
	TaskStatusFailed          int32 = 80 // 失败(可重试)
	TaskStatusCancelled       int32 = 90 // 已取消(终态)
)

// TaskStatusName 状态中文名
var TaskStatusName = map[int32]string{
	TaskStatusPending:         "待执行",
	TaskStatusSearching:       "搜索新闻中",
	TaskStatusFetching:        "抓取正文中",
	TaskStatusGenerating:      "AI生成中",
	TaskStatusDraftCreating:   "创建草稿中",
	TaskStatusAwaitingConfirm: "待人工确认",
	TaskStatusPublishing:      "发布中",
	TaskStatusPublished:       "已发布",
	TaskStatusFailed:          "失败",
	TaskStatusCancelled:       "已取消",
}

// TaskStatusTransitions 状态机流转表(from -> 允许的 to 列表), 用于 CAS 更新合法性校验
var TaskStatusTransitions = map[int32][]int32{
	TaskStatusPending:         {TaskStatusSearching, TaskStatusCancelled},
	TaskStatusSearching:       {TaskStatusFetching, TaskStatusFailed, TaskStatusCancelled},
	TaskStatusFetching:        {TaskStatusGenerating, TaskStatusFailed, TaskStatusCancelled},
	TaskStatusGenerating:      {TaskStatusDraftCreating, TaskStatusFailed, TaskStatusCancelled},
	TaskStatusDraftCreating:   {TaskStatusAwaitingConfirm, TaskStatusFailed, TaskStatusCancelled},
	TaskStatusAwaitingConfirm: {TaskStatusPublishing, TaskStatusCancelled},
	TaskStatusPublishing:      {TaskStatusPublished, TaskStatusFailed},
	TaskStatusFailed:          {TaskStatusPending, TaskStatusCancelled},
	TaskStatusPublished:       {},
	TaskStatusCancelled:       {},
}

// 源文章正文状态
const (
	ContentStatusNone   int32 = 0 // 未抓取
	ContentStatusOK     int32 = 1 // 抓取成功
	ContentStatusFailed int32 = 2 // 抓取失败
)

// 草稿发布状态
const (
	DraftPublishStatusNone     int32 = 0 // 未发布
	DraftPublishStatusCreated  int32 = 1 // 草稿已创建
	DraftPublishStatusDoing    int32 = 2 // 发布中
	DraftPublishStatusSuccess  int32 = 3 // 已发布
	DraftPublishStatusFailed   int32 = 4 // 发布失败
)
