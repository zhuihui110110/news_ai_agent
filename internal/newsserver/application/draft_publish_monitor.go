// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 草稿发布 Monitor: 轮询"发布中"(freepublish 已提交)的草稿结果,
// 成功则任务流转 published, 失败则任务流转 failed(等待重试)
//
// 说明: 提交发布动作由人工确认接口(service.PublishDraft)触发, 本 Monitor 只负责收尾

package application

import (
	"context"
	"fmt"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
	"news_ai_server/internal/innerpkg/wechat_client"
	"news_ai_server/pkg/log"
)

// StartDraftPublishMonitor 启动发布轮询 Monitor(阻塞直至 ctx 取消)
func StartDraftPublishMonitor(ctx context.Context, d *Deps) {
	interval := time.Duration(d.TaskOpts.ScanIntervalSecond) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Infof("draft publish monitor 退出")
			return
		case <-ticker.C:
			pollPublishing(ctx, d)
		}
	}
}

// pollPublishing 轮询发布中的草稿
func pollPublishing(ctx context.Context, d *Deps) {
	drafts, err := d.Factory.NewsDraftRepository().ListByPublishStatus(ctx, constants.DraftPublishStatusDoing, 20)
	if err != nil {
		log.Errorf("查询发布中草稿失败: %v", err)
		return
	}
	for i := range drafts {
		dr := &drafts[i]
		if dr.IsMock == 1 {
			// mock 草稿直接置为成功
			dr.PublishStatus = constants.DraftPublishStatusSuccess
			dr.ArticleURL = fmt.Sprintf("mock://article/%d", dr.ID)
			if err = d.Factory.NewsDraftRepository().Update(ctx, dr); err != nil {
				log.Errorf("更新 mock 草稿失败: id=%d err=%v", dr.ID, err)
			}
			finishTaskPublish(ctx, d, dr.TaskID, true, "")
			continue
		}
		res, err := wechat_client.GetClient().GetPublishResult(ctx, dr.PublishID)
		if err != nil {
			log.Warnf("查询发布结果失败(下轮重试): draftId=%d err=%v", dr.ID, err)
			continue
		}
		switch res.PublishStatus {
		case wechat_client.PublishResultSuccess:
			dr.PublishStatus = constants.DraftPublishStatusSuccess
			dr.ArticleURL = res.ArticleURL
			if err = d.Factory.NewsDraftRepository().Update(ctx, dr); err != nil {
				log.Errorf("更新草稿发布结果失败: id=%d err=%v", dr.ID, err)
				continue
			}
			finishTaskPublish(ctx, d, dr.TaskID, true, "")
			log.Infof("发布成功: draftId=%d url=%s", dr.ID, res.ArticleURL)
		case wechat_client.PublishResultDoing:
			// 仍在发布中, 下轮继续
		default:
			dr.PublishStatus = constants.DraftPublishStatusFailed
			dr.ErrMsg = fmt.Sprintf("发布失败: status=%d %s", res.PublishStatus, res.FailMsg)
			if err = d.Factory.NewsDraftRepository().Update(ctx, dr); err != nil {
				log.Errorf("更新草稿失败状态失败: id=%d err=%v", dr.ID, err)
			}
			finishTaskPublish(ctx, d, dr.TaskID, false, dr.ErrMsg)
			log.Errorf("发布失败: draftId=%d %s", dr.ID, dr.ErrMsg)
		}
	}
}

// finishTaskPublish 发布收尾: 任务 publishing -> published / failed
func finishTaskPublish(ctx context.Context, d *Deps, taskID int64, success bool, errMsg string) {
	repo := d.Factory.NewsTaskRepository()
	t, err := repo.GetByID(ctx, taskID)
	if t == nil || err != nil {
		return
	}
	if success {
		if _, err = repo.TransStatus(ctx, taskID,
			constants.TaskStatusPublishing, constants.TaskStatusPublished, "发布成功", ""); err != nil {
			log.Errorf("任务标记发布成功失败: id=%d err=%v", taskID, err)
		}
		return
	}
	ok, err := repo.TransStatus(ctx, taskID,
		constants.TaskStatusPublishing, constants.TaskStatusFailed, "", errMsg)
	if err != nil {
		log.Errorf("任务标记发布失败状态出错: id=%d err=%v", taskID, err)
		return
	}
	if ok {
		_ = repo.UpdateRetryCount(ctx, taskID, t.RetryCount+1)
	}
}
