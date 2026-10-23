// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 任务管理: 创建/列表/详情/重试/取消

package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
	"news_ai_server/internal/domain/infrastructure/mysql/model"
	msg "news_ai_server/internal/newsserver/app/message"
)

// CreateTask 创建任务(领域合法性校验后落库, 等待 Monitor 调度)
func (s *Service) CreateTask(ctx context.Context, req *msg.CreateTaskReq) (*msg.CreateTaskResp, error) {
	topics := req.Topics
	if len(topics) == 0 {
		topics = constants.AllTopics
	}
	for _, tp := range topics {
		if _, ok := constants.TopicName[tp]; !ok {
			return nil, fmt.Errorf("非法领域: %s(支持 livelihood/economy/ai/tech)", tp)
		}
	}
	tasks := make([]*model.NewsTaskModel, 0, len(topics))
	for _, tp := range topics {
		keywords := s.newsOpts.Topics[tp]
		if len(req.Keywords) > 0 {
			keywords = req.Keywords // 允许人工指定关键词覆盖
		}
		m := &model.NewsTaskModel{
			TaskNo:     fmt.Sprintf("NT%s%s%04d", tp, time.Now().Format("0102150405"), time.Now().UnixNano()%10000),
			Topic:      tp,
			Keywords:   strings.Join(keywords, ","),
			TaskStatus: constants.TaskStatusPending,
			StageMsg:   "任务已创建, 等待调度",
		}
		if err := s.factory.NewsTaskRepository().Create(ctx, m); err != nil {
			return nil, fmt.Errorf("创建任务失败: %w", err)
		}
		tasks = append(tasks, m)
	}
	return &msg.CreateTaskResp{Tasks: tasks}, nil
}

// ListTasks 任务分页列表
func (s *Service) ListTasks(ctx context.Context, req *msg.TaskListReq) (*msg.TaskListResp, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	list, total, err := s.factory.NewsTaskRepository().ListPage(ctx, req.Topic, req.Status,
		(req.Page-1)*req.PageSize, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &msg.TaskListResp{Total: total, List: list}, nil
}

// TaskDetail 任务详情(任务 + 源新闻 + 生成文章 + 草稿)
func (s *Service) TaskDetail(ctx context.Context, id int64) (*msg.TaskDetailResp, error) {
	t, err := s.factory.NewsTaskRepository().GetByID(ctx, id)
	if t == nil {
		return nil, fmt.Errorf("任务不存在: %d", id)
	}
	if err != nil {
		return nil, err
	}
	sources, err := s.factory.NewsSourceArticleRepository().ListByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}
	article, err := s.factory.NewsArticleRepository().GetByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}
	draft, err := s.factory.NewsDraftRepository().GetByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &msg.TaskDetailResp{Task: t, Articles: sources, Article: article, Draft: draft}, nil
}

// RetryTask 人工重试: failed(或待确认)任务重回 pending
func (s *Service) RetryTask(ctx context.Context, id int64) error {
	t, err := s.factory.NewsTaskRepository().GetByID(ctx, id)
	if t == nil {
		return fmt.Errorf("任务不存在: %d", id)
	}
	if err != nil {
		return err
	}
	if !t.CanTransTo(constants.TaskStatusPending) {
		return fmt.Errorf("当前状态[%s]不允许重试", constants.TaskStatusName[t.TaskStatus])
	}
	ok, err := s.factory.NewsTaskRepository().TransStatus(ctx, id, t.TaskStatus,
		constants.TaskStatusPending, "人工重试入队", "")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("任务状态已变化, 请刷新后重试")
	}
	return nil
}

// CancelTask 人工取消(终态: published/cancelled 除外)
func (s *Service) CancelTask(ctx context.Context, id int64) error {
	t, err := s.factory.NewsTaskRepository().GetByID(ctx, id)
	if t == nil {
		return fmt.Errorf("任务不存在: %d", id)
	}
	if err != nil {
		return err
	}
	if !t.CanTransTo(constants.TaskStatusCancelled) {
		return fmt.Errorf("当前状态[%s]不允许取消", constants.TaskStatusName[t.TaskStatus])
	}
	ok, err := s.factory.NewsTaskRepository().TransStatus(ctx, id, t.TaskStatus,
		constants.TaskStatusCancelled, "人工取消", "")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("任务状态已变化, 请刷新后重试")
	}
	return nil
}
