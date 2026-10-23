// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 草稿管理与人工确认发布:
// 流程 = 系统自动创建草稿 -> 人工在公众号后台/本系统审核 -> 确认后提交 freepublish

package service

import (
	"context"
	"fmt"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
	msg "news_ai_server/internal/newsserver/app/message"
)

// ListDrafts 草稿列表(带任务与文章摘要, 供人工审核)
func (s *Service) ListDrafts(ctx context.Context, req *msg.DraftListReq) (*msg.DraftListResp, error) {
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	drafts, err := s.factory.NewsDraftRepository().ListByPublishStatus(ctx, derefStatus(req.PublishStatus), req.Limit)
	if err != nil {
		return nil, err
	}
	items := make([]msg.DraftItem, 0, len(drafts))
	for i := range drafts {
		d := drafts[i]
		item := msg.DraftItem{Draft: d}
		if t, e := s.factory.NewsTaskRepository().GetByID(ctx, d.TaskID); e == nil && t != nil {
			item.Task = *t
		}
		if a, e := s.factory.NewsArticleRepository().GetByID(ctx, d.ArticleID); e == nil && a != nil {
			item.Title, item.Digest = a.Title, a.Digest
		}
		items = append(items, item)
	}
	return &msg.DraftListResp{List: items}, nil
}

// PublishDraft 人工确认发布: awaiting_confirm 任务 -> 提交 freepublish -> publishing(由 Monitor 收尾)
func (s *Service) PublishDraft(ctx context.Context, draftID int64) (*msg.PublishResp, error) {
	draftRepo := s.factory.NewsDraftRepository()
	taskRepo := s.factory.NewsTaskRepository()

	draft, err := draftRepo.GetByID(ctx, draftID)
	if draft == nil {
		return nil, fmt.Errorf("草稿不存在: %d", draftID)
	}
	if err != nil {
		return nil, err
	}
	if draft.PublishStatus != constants.DraftPublishStatusCreated {
		return nil, fmt.Errorf("草稿状态不允许发布: %d", draft.PublishStatus)
	}
	task, err := taskRepo.GetByID(ctx, draft.TaskID)
	if task == nil {
		return nil, fmt.Errorf("任务不存在: %d", draft.TaskID)
	}
	if err != nil {
		return nil, err
	}
	if task.TaskStatus != constants.TaskStatusAwaitingConfirm {
		return nil, fmt.Errorf("任务状态[%s]不允许发布", constants.TaskStatusName[task.TaskStatus])
	}

	// 状态机: awaiting_confirm -> publishing
	ok, err := taskRepo.TransStatus(ctx, task.ID, constants.TaskStatusAwaitingConfirm,
		constants.TaskStatusPublishing, "人工确认, 提交发布", "")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("任务状态已变化, 请刷新后重试")
	}

	// mock 草稿: 直接标记发布中, 由发布 Monitor 置为成功
	if draft.IsMock == 1 || !s.wechat.Enabled() {
		draft.PublishStatus = constants.DraftPublishStatusDoing
		if err = draftRepo.Update(ctx, draft); err != nil {
			return nil, err
		}
		return &msg.PublishResp{PublishID: "", Mock: true}, nil
	}

	// 真实发布: 提交 freepublish
	publishID, err := s.wechat.SubmitPublish(ctx, draft.MediaID)
	if err != nil {
		// 提交失败回滚任务状态
		_, _ = taskRepo.TransStatus(ctx, task.ID, constants.TaskStatusPublishing,
			constants.TaskStatusFailed, "", "提交发布失败: "+err.Error())
		_ = taskRepo.UpdateRetryCount(ctx, task.ID, task.RetryCount+1)
		draft.PublishStatus = constants.DraftPublishStatusFailed
		draft.ErrMsg = err.Error()
		_ = draftRepo.Update(ctx, draft)
		return nil, fmt.Errorf("提交发布失败: %w", err)
	}
	draft.PublishID = publishID
	draft.PublishStatus = constants.DraftPublishStatusDoing
	if err = draftRepo.Update(ctx, draft); err != nil {
		return nil, err
	}
	return &msg.PublishResp{PublishID: publishID}, nil
}

// derefStatus 状态指针解引用(0 表示不过滤)
func derefStatus(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
