// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 智谱联网搜索 Provider: 复用 llm 客户端在 /chat/completions 上启用
// 原生 web_search 工具(search_result=true 返回原始搜索结果), 取响应顶层 web_search
// 数组作为新闻源。大陆网络下 Google/Bing RSS 均不可达, 该渠道经智谱服务端搜索, 无墙限制。

package news_client

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"news_ai_server/pkg/llm"
)

// zhipuSearch 智谱 web_search 检索渠道
type zhipuSearch struct {
	llm   *llm.Client
	model string // 检索用的轻量模型(如 glm-4-flash), 与文章生成模型相互独立
}

// recentDays 搜索结果保留的最近天数(超出且结果充足时淘汰旧闻)
const recentDays = 14

// Name 渠道名
func (z *zhipuSearch) Name() string { return "zhipu" }

// Search 调用 web_search 工具检索新闻, 返回原始搜索结果列表
func (z *zhipuSearch) Search(ctx context.Context, query string, limit int) ([]NewsItem, error) {
	if z.llm == nil || !z.llm.Ready() {
		return nil, fmt.Errorf("智谱检索渠道需要配置 llm.api-key")
	}
	req := &llm.ChatRequest{
		Model: z.model,
		Messages: []llm.Message{{
			Role: "user",
			// 引导模型优先检索近期新闻, 日期由模型结合当前时间自行把握
			Content: fmt.Sprintf("帮我检索最近一周内关于「%s」的新闻资讯", query),
		}},
		MaxTokens: 1024,
		Temperature: 0.3,
		Tools: []llm.Tool{{
			Type: "web_search",
			WebSearch: &llm.WebSearchTool{
				Enable:       true,
				SearchQuery:  query,
				SearchResult: true,
			},
		}},
	}
	resp, err := z.llm.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("智谱 web_search 调用失败: %w", err)
	}
	if len(resp.WebSearch) == 0 {
		return nil, fmt.Errorf("智谱 web_search 无搜索结果(query=%s)", query)
	}

	seen := make(map[string]struct{}, len(resp.WebSearch))
	var items []NewsItem
	for i := range resp.WebSearch {
		r := resp.WebSearch[i]
		link := strings.TrimSpace(r.Link)
		title := strings.TrimSpace(r.Title)
		if link == "" || title == "" {
			continue
		}
		if _, dup := seen[link]; dup {
			continue
		}
		seen[link] = struct{}{}
		items = append(items, NewsItem{
			Provider:    z.Name(),
			URL:         link,
			Title:       title,
			Snippet:     truncateRunes(StripTags(r.Content), 500),
			SourceName:  strings.TrimSpace(r.Media),
			PublishedAt: parseSimpleDate(r.PublishDate),
		})
	}

	// 近期优先: 有发布时间且在 recentDays 内的排前, 时间倒序; 无时间/过期的排后
	fresh, stale := splitRecent(items, recentDays)
	if len(fresh) < limit && len(stale) > 0 {
		fresh = append(fresh, stale...)
	}
	if len(fresh) > limit {
		fresh = fresh[:limit]
	}
	return fresh, nil
}

// splitRecent 按发布时间切分为(近期, 其余), 各自组内时间倒序
func splitRecent(items []NewsItem, days int) (recent, others []NewsItem) {
	cut := time.Now().AddDate(0, 0, -days)
	recent, others = make([]NewsItem, 0, len(items)), make([]NewsItem, 0, len(items))
	for i := range items {
		if items[i].PublishedAt != nil && items[i].PublishedAt.After(cut) {
			recent = append(recent, items[i])
		} else {
			others = append(others, items[i])
		}
	}
	byDateDesc := func(s []NewsItem) {
		sort.SliceStable(s, func(i, j int) bool {
			ti, tj := timeOfPtr(s[i].PublishedAt), timeOfPtr(s[j].PublishedAt)
			return ti.After(tj)
		})
	}
	byDateDesc(recent)
	byDateDesc(others)
	return recent, others
}

// parseSimpleDate 解析 2006-01-02 日期
func parseSimpleDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return &t
	}
	return nil
}

// timeOfPtr 空安全取值
func timeOfPtr(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// truncateRunes 按 rune 截断
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
