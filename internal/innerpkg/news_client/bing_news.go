// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: Bing News RSS 检索实现(免 Key):
// https://www.bing.com/news/search?q=<query>&format=RSS

package news_client

import (
	"context"
	"fmt"
	"strings"
)

// bingNews Bing News RSS 渠道
type bingNews struct{ base baseHTTP }

// Name 渠道名
func (b *bingNews) Name() string { return "bing" }

// Search 检索新闻
func (b *bingNews) Search(ctx context.Context, query string, limit int) ([]NewsItem, error) {
	rawURL := fmt.Sprintf(
		"https://www.bing.com/news/search?q=%s&format=RSS&setmkt=zh-CN&setlang=zh-CN",
		escapeQuery(query),
	)
	feed, err := b.base.fetchRSS(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("bing 检索[%s]失败: %w", query, err)
	}
	// Bing RSS 的标题常为 "标题 - 媒体名", 做一次拆分
	items := make([]rssItem, 0, len(feed.Channel.Items))
	for _, it := range feed.Channel.Items {
		if idx := strings.LastIndex(it.Title, " - "); idx > 0 && it.Source == "" {
			it.Source = strings.TrimSpace(it.Title[idx+3:])
			it.Title = strings.TrimSpace(it.Title[:idx])
		}
		items = append(items, it)
	}
	return toItems(b.Name(), items, limit), nil
}
