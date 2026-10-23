// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: Google News RSS 检索实现(免 Key):
// https://news.google.com/rss/search?q=<query>&hl=zh-CN&gl=CN&ceid=CN:zh-Hans

package news_client

import (
	"context"
	"fmt"
)

// googleNews Google News RSS 渠道
type googleNews struct{ base baseHTTP }

// Name 渠道名
func (g *googleNews) Name() string { return "google" }

// Search 检索新闻
func (g *googleNews) Search(ctx context.Context, query string, limit int) ([]NewsItem, error) {
	rawURL := fmt.Sprintf(
		"https://news.google.com/rss/search?q=%s&hl=zh-CN&gl=CN&ceid=CN:zh-Hans",
		escapeQuery(query),
	)
	feed, err := g.base.fetchRSS(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("google 检索[%s]失败: %w", query, err)
	}
	return toItems(g.Name(), feed.Channel.Items, limit), nil
}
