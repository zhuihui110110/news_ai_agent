// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻任务 Monitor(参照 stock_ai taskserver 的 Monitor 模式):
// 周期性执行 [自动建任务 -> 失败重试 -> 认领并执行待办任务],
// 每个任务按状态机流转: pending -> searching -> fetching -> generating
//                    -> draft_creating -> awaiting_confirm(等人工确认发布)
//
// 频率: task.scan-interval-second(默认30s)
// 存储: news_task / news_source_article / news_article / news_draft
// 去重: news_source_article.url_md5 全局唯一
// 重试: failed 状态且 retry_count < max-retries, 间隔 retry-interval-second 后重回 pending

package application

import (
	"context"
	"fmt"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/internal/domain/infrastructure/mysql/repository"
	"news_ai_server/pkg/log"
)

// StartNewsTaskMonitor 启动新闻任务 Monitor(阻塞直至 ctx 取消)
func StartNewsTaskMonitor(ctx context.Context, d *Deps) {
	interval := time.Duration(d.TaskOpts.ScanIntervalSecond) * time.Second
	job := func() {
		autoCreateTasks(ctx, d)
		recoverInflightTasks(ctx, d)
		retryFailedTasks(ctx, d)
		processPendingTasks(ctx, d)
	}
	job() // 冷启动立即执行一轮
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Infof("news task monitor 退出")
			return
		case <-ticker.C:
			job()
		}
	}
}

// autoCreateTasks 按领域冷却自动创建任务(auto-create-interval-hour 内已有任务则跳过)
func autoCreateTasks(ctx context.Context, d *Deps) {
	if !d.TaskOpts.AutoCreate {
		return
	}
	since := time.Now().Add(-time.Duration(d.TaskOpts.AutoCreateIntervalHour) * time.Hour)
	for topic := range d.NewsOpts.Topics {
		n, err := d.Factory.NewsTaskRepository().CountTopicSince(ctx, topic, since)
		if err != nil {
			log.Errorf("统计任务数失败: %v", err)
			continue
		}
		if n > 0 {
			continue
		}
		m := &model.NewsTaskModel{
			TaskNo:     genTaskNo(topic),
			Topic:      topic,
			Keywords:   joinKeywords(d.NewsOpts.Topics[topic]),
			TaskStatus: constants.TaskStatusPending,
			StageMsg:   "任务已创建, 等待调度",
		}
		if err = d.Factory.NewsTaskRepository().Create(ctx, m); err != nil {
			log.Errorf("自动创建任务失败: %v", err)
			continue
		}
		log.Infof("自动创建任务: id=%d topic=%s taskNo=%s", m.ID, topic, m.TaskNo)
	}
}

// recoverInflightTasks 崩溃自愈: 进程重启会遗留在执行态(searching/fetching/generating/draft_creating)
// 且无 worker 接手的任务, 超过 staleInflight 阈值后 CAS 拉回 pending 重新执行
func recoverInflightTasks(ctx context.Context, d *Deps) {
	const staleInflight = 5 * time.Minute // 大于任一单阶段最长耗时(检索/抓取/生成/建草稿)
	for _, st := range []int32{
		constants.TaskStatusSearching,
		constants.TaskStatusFetching,
		constants.TaskStatusGenerating,
		constants.TaskStatusDraftCreating,
	} {
		tasks, err := d.Factory.NewsTaskRepository().ListByStatus(ctx, st, 20)
		if err != nil {
			log.Errorf("查询执行态任务出错(status=%d): %v", st, err)
			continue
		}
		for i := range tasks {
			t := &tasks[i]
			if time.Since(t.UpdatedAt) < staleInflight {
				continue // 正在正常执行中
			}
			ok, err := d.Factory.NewsTaskRepository().TransStatus(ctx, t.ID,
				st, constants.TaskStatusPending, "执行态超时, 重新入队", "")
			if err != nil {
				log.Errorf("执行态任务重新入队失败: id=%d err=%v", t.ID, err)
				continue
			}
			if ok {
				log.Warnf("执行态任务超时重新入队(疑似进程重启遗留): id=%d status=%d idle=%v",
					t.ID, st, time.Since(t.UpdatedAt).Round(time.Second))
			}
		}
	}
}

// retryFailedTasks 失败任务重试: 达到冷却时间且未超最大重试次数的 failed 任务重回 pending
func retryFailedTasks(ctx context.Context, d *Deps) {
	tasks, err := d.Factory.NewsTaskRepository().ListByStatus(ctx, constants.TaskStatusFailed, 20)
	if err != nil {
		log.Errorf("查询失败任务出错: %v", err)
		return
	}
	coolDown := time.Duration(d.TaskOpts.RetryIntervalSecond) * time.Second
	for i := range tasks {
		t := &tasks[i]
		if t.RetryCount >= int32(d.TaskOpts.MaxRetries) {
			continue
		}
		if time.Since(t.UpdatedAt) < coolDown {
			continue
		}
		ok, err := d.Factory.NewsTaskRepository().TransStatus(ctx, t.ID,
			constants.TaskStatusFailed, constants.TaskStatusPending, "重试入队", "")
		if err != nil {
			log.Errorf("任务重试入队失败: id=%d err=%v", t.ID, err)
			continue
		}
		if ok {
			log.Infof("任务重试入队: id=%d retryCount=%d", t.ID, t.RetryCount)
		}
	}
}

// processPendingTasks 认领并执行待办任务(CAS 认领, 每轮最多 maxPerRound 个, 串行执行)
func processPendingTasks(ctx context.Context, d *Deps) {
	const maxPerRound = 2
	tasks, err := d.Factory.NewsTaskRepository().ListByStatus(ctx, constants.TaskStatusPending, maxPerRound)
	if err != nil {
		log.Errorf("查询待办任务出错: %v", err)
		return
	}
	for i := range tasks {
		t := &tasks[i]
		ok, err := d.Factory.NewsTaskRepository().TransStatus(ctx, t.ID,
			constants.TaskStatusPending, constants.TaskStatusSearching, "开始检索新闻", "")
		if err != nil {
			log.Errorf("认领任务失败: id=%d err=%v", t.ID, err)
			continue
		}
		if !ok {
			continue // 已被其他实例认领或状态变化
		}
		runPipeline(ctx, d, t.ID)
	}
}

// runPipeline 执行任务流水线, 逐阶段 CAS 流转状态机; 任一阶段失败进入 failed 并累加重试计数
func runPipeline(ctx context.Context, d *Deps, taskID int64) {
	repo := d.Factory.NewsTaskRepository()
	t, err := repo.GetByID(ctx, taskID)
	if t == nil || err != nil {
		log.Errorf("加载任务失败: id=%d err=%v", taskID, err)
		return
	}
	log.Infof("开始执行任务: id=%d topic=%s taskNo=%s", t.ID, t.Topic, t.TaskNo)

	stages := []struct {
		from int32
		to   int32
		name string
		fn   func() error
	}{
		{constants.TaskStatusSearching, constants.TaskStatusFetching, "搜索新闻", func() error {
			return stageSearch(ctx, d, t)
		}},
		{constants.TaskStatusFetching, constants.TaskStatusGenerating, "抓取正文", func() error {
			return stageFetch(ctx, d, t)
		}},
		{constants.TaskStatusGenerating, constants.TaskStatusDraftCreating, "AI生成文章", func() error {
			return stageGenerate(ctx, d, t)
		}},
		{constants.TaskStatusDraftCreating, constants.TaskStatusAwaitingConfirm, "创建公众号草稿", func() error {
			return stageCreateDraft(ctx, d, t)
		}},
	}
	for _, st := range stages {
		if err = ctx.Err(); err != nil {
			return
		}
		if err = st.fn(); err != nil {
			failTask(ctx, repo, t, st.from, err)
			return
		}
		ok, err := repo.TransStatus(ctx, t.ID, st.from, st.to, st.name+"完成", "")
		if err != nil {
			log.Errorf("状态流转失败: id=%d %d->%d err=%v", t.ID, st.from, st.to, err)
			return
		}
		if !ok {
			log.Warnf("状态流转被跳过(状态已被外部修改): id=%d %d->%d", t.ID, st.from, st.to)
			return
		}
		t.TaskStatus = st.to
	}
	log.Infof("任务执行完成, 待人工确认发布: id=%d taskNo=%s", t.ID, t.TaskNo)
}

// failTask 任务失败处理: CAS 进入 failed + 累加重试计数
func failTask(ctx context.Context, repo repository.NewsTaskRepository, t *model.NewsTaskModel, from int32, cause error) {
	msg := cause.Error()
	if len(msg) > 1000 {
		msg = msg[:1000]
	}
	ok, err := repo.TransStatus(ctx, t.ID, from, constants.TaskStatusFailed, "", msg)
	if err != nil {
		log.Errorf("任务标记失败状态出错: id=%d err=%v", t.ID, err)
		return
	}
	if !ok {
		log.Warnf("任务标记失败被跳过(状态已变化): id=%d from=%d", t.ID, from)
		return
	}
	if err = repo.UpdateRetryCount(ctx, t.ID, t.RetryCount+1); err != nil {
		log.Errorf("更新重试计数失败: id=%d err=%v", t.ID, err)
	}
	log.Errorf("任务阶段失败: id=%d from=%d err=%v", t.ID, from, cause)
}

// genTaskNo 生成任务编号: NT + 领域 + 时间 + 随机段(唯一索引兜底)
func genTaskNo(topic string) string {
	return fmt.Sprintf("NT%s%s%04d", topic, time.Now().Format("0102150405"), time.Now().UnixNano()%10000)
}

// joinKeywords 关键词拼接
func joinKeywords(ws []string) string {
	out := ""
	for i, w := range ws {
		if i > 0 {
			out += ","
		}
		out += w
	}
	return out
}
