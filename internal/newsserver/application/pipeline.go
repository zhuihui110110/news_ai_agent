// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 任务流水线四阶段实现:
//
//	stageSearch    多渠道检索新闻, url_md5 全局去重入库
//	stageFetch     并发抓取正文(worker pool), 按正文质量+时间选出素材
//	stageGenerate  组装素材调用 GLM 直连 API 生成公众号文章(标题/摘要/HTML/标签)
//	stageCreateDraft 上传封面素材 -> 创建公众号草稿(未配置公众号时落 mock 草稿)

package application

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/internal/innerpkg/news_client"
	"news_ai_server/internal/innerpkg/wechat_client"
	"news_ai_server/pkg/log"
)

// minArticles 任务至少需要的源新闻数
const minArticles = 3

// stageSearch 阶段1: 检索新闻并去重入库
func stageSearch(ctx context.Context, d *Deps, t *model.NewsTaskModel) error {
	keywords := d.NewsOpts.Topics[t.Topic]
	if len(keywords) == 0 {
		return fmt.Errorf("领域[%s]未配置检索关键词", t.Topic)
	}
	repo := d.Factory.NewsSourceArticleRepository()
	total := int64(0)
	for _, kw := range keywords {
		for _, p := range d.Providers {
			items, err := p.Search(ctx, kw, 15)
			if err != nil {
				log.Warnf("检索失败(继续): provider=%s kw=%s err=%v", p.Name(), kw, err)
				continue
			}
			ms := make([]model.NewsSourceArticleModel, 0, len(items))
			for i := range items {
				it := items[i]
				ms = append(ms, model.NewsSourceArticleModel{
					TaskID:      t.ID,
					Provider:    it.Provider,
					URL:         it.URL,
					URLMd5:      md5Hex(it.URL),
					Title:       truncateStr(it.Title, 500),
					Snippet:     truncateStr(it.Snippet, 2000),
					Author:      truncateStr(it.Author, 120),
					SourceName:  truncateStr(it.SourceName, 120),
					PublishedAt: it.PublishedAt,
				})
			}
			n, err := repo.BatchUpsertIgnore(ctx, ms)
			if err != nil {
				return fmt.Errorf("源新闻入库失败: %w", err)
			}
			total += n
		}
	}
	all, err := repo.ListByTaskID(ctx, t.ID)
	if err != nil {
		return err
	}
	if len(all) < minArticles {
		return fmt.Errorf("有效新闻不足: 需要%d条, 实际%d条(新增%d)", minArticles, len(all), total)
	}
	return nil
}

// stageFetch 阶段2: 并发抓取正文并选出素材
func stageFetch(ctx context.Context, d *Deps, t *model.NewsTaskModel) error {
	repo := d.Factory.NewsSourceArticleRepository()
	arts, err := repo.ListByTaskID(ctx, t.ID)
	if err != nil {
		return err
	}
	fetchAll(ctx, d, arts)

	// 选材: 有正文的优先, 其次按发布时间倒序
	sort.SliceStable(arts, func(i, j int) bool {
		bi, bj := arts[i].ContentStatus == constants.ContentStatusOK, arts[j].ContentStatus == constants.ContentStatusOK
		if bi != bj {
			return bi
		}
		ti, tj := timeOf(arts[i].PublishedAt), timeOf(arts[j].PublishedAt)
		return ti.After(tj)
	})
	limit := d.NewsOpts.MaxArticlesPerTask
	if limit > len(arts) {
		limit = len(arts)
	}
	ids := make([]int64, 0, limit)
	for i := 0; i < limit; i++ {
		ids = append(ids, arts[i].ID)
	}
	if err = repo.MarkSelected(ctx, t.ID, ids); err != nil {
		return err
	}
	okCount := 0
	for i := 0; i < limit; i++ {
		if arts[i].ContentStatus == constants.ContentStatusOK {
			okCount++
		}
	}
	if okCount == 0 {
		log.Warnf("任务无正文可用, 将使用摘要生成: id=%d", t.ID)
	}
	return nil
}

// fetchAll worker pool 并发抓取正文, 单条失败不阻断
func fetchAll(ctx context.Context, d *Deps, arts []model.NewsSourceArticleModel) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, d.NewsOpts.FetchWorkers)
	for i := range arts {
		a := &arts[i]
		if a.ContentStatus != constants.ContentStatusNone {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			cctx, cancel := context.WithTimeout(ctx, d.NewsOpts.FetchTimeout)
			defer cancel()
			ac, err := news_client.FetchArticle(cctx, d.ScrapeHTTP, d.NewsOpts.UserAgent, a.URL)
			status, content := constants.ContentStatusOK, ""
			if err != nil || ac == nil {
				status = constants.ContentStatusFailed
				log.Debugf("正文抓取失败(使用摘要兜底): id=%d err=%v", a.ID, err)
			} else {
				content = ac.Text
			}
			_ = d.Factory.NewsSourceArticleRepository().
				UpdateContent(ctx, a.ID, truncateStr(content, 12000), int32(len([]rune(content))), status)
		}()
	}
	wg.Wait()
}

// stageGenerate 阶段3: 调用 GLM 生成文章
func stageGenerate(ctx context.Context, d *Deps, t *model.NewsTaskModel) error {
	if !d.LLM.Ready() {
		return fmt.Errorf("LLM api-key 未配置(llm.api-key 或环境变量 LLM_API_KEY)")
	}
	repo := d.Factory.NewsSourceArticleRepository()
	all, err := repo.ListByTaskID(ctx, t.ID)
	if err != nil {
		return err
	}
	selected := make([]model.NewsSourceArticleModel, 0, len(all))
	for i := range all {
		if all[i].Selected == 1 {
			selected = append(selected, all[i])
		}
	}
	if len(selected) == 0 {
		return fmt.Errorf("无选中素材")
	}
	result, usage, err := GenerateArticle(ctx, d.LLM, t.Topic, selected)
	if err != nil {
		return err
	}
	am := &model.NewsArticleModel{
		TaskID:   t.ID,
		Title:    truncateStr(result.Title, 250),
		Digest:   truncateStr(result.Digest, 500),
		Content:  result.ContentHTML,
		Tags:     truncateStr(strings.Join(result.Tags, ","), 250),
		LLMModel: usage.Model,
	}
	if err = d.Factory.NewsArticleRepository().Upsert(ctx, am); err != nil {
		return fmt.Errorf("文章入库失败: %w", err)
	}
	_ = d.Factory.NewsTaskRepository().UpdateLLMUsage(ctx, t.ID, usage.Model,
		int32(usage.PromptTokens), int32(usage.CompletionTokens))
	log.Infof("文章生成完成: id=%d title=%s tokens=%d", t.ID, am.Title, usage.TotalTokens)
	return nil
}

// stageCreateDraft 阶段4: 创建公众号草稿(自动草稿, 人工确认后发布)
func stageCreateDraft(ctx context.Context, d *Deps, t *model.NewsTaskModel) error {
	article, err := d.Factory.NewsArticleRepository().GetByTaskID(ctx, t.ID)
	if article == nil || err != nil {
		return fmt.Errorf("生成文章不存在: %w", err)
	}
	draft := &model.NewsDraftModel{TaskID: t.ID, ArticleID: article.ID}

	if d.Wechat.Enabled() {
		mediaID, err := createWechatDraft(ctx, d, article)
		if err != nil {
			return err
		}
		draft.MediaID = mediaID
		draft.PublishStatus = constants.DraftPublishStatusCreated
	} else {
		// 未配置公众号: 落 mock 草稿, 全流程可本地跑通
		draft.MediaID = fmt.Sprintf("MOCK-%s", t.TaskNo)
		draft.PublishStatus = constants.DraftPublishStatusCreated
		draft.IsMock = 1
	}
	if err = d.Factory.NewsDraftRepository().Upsert(ctx, draft); err != nil {
		return fmt.Errorf("草稿记录入库失败: %w", err)
	}
	return nil
}

// createWechatDraft 真实公众号草稿: 封面素材上传 + draft/add
func createWechatDraft(ctx context.Context, d *Deps, article *model.NewsArticleModel) (string, error) {
	wc := wechat_client.GetClient()
	// 封面: 使用配置的本地默认封面图(永久素材)
	coverData, err := os.ReadFile(d.WechatOpts.DefaultCoverPath)
	if err != nil {
		return "", fmt.Errorf("读取封面图失败(%s), 请检查 wechat.default-cover-path: %w", d.WechatOpts.DefaultCoverPath, err)
	}
	thumbMediaID, err := wc.UploadMaterialImageBytes(ctx, "cover.jpg", coverData)
	if err != nil {
		return "", fmt.Errorf("上传封面素材失败: %w", err)
	}
	mediaID, err := wc.AddDraft(ctx, &wechat_client.DraftArticle{
		Title:            article.Title,
		Author:           "AI新闻编辑",
		Digest:           article.Digest,
		Content:          article.Content,
		ThumbMediaID:     thumbMediaID,
		NeedOpenComment:  0,
		OnlyFansCanComment: 0,
	})
	if err != nil {
		return "", fmt.Errorf("创建公众号草稿失败: %w", err)
	}
	return mediaID, nil
}

// md5Hex 字符串 md5
func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// truncateStr 字符串按 rune 截断
func truncateStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// timeOf 空安全时间
func timeOf(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
