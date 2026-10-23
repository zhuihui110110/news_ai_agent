// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻检索客户端封装: 统一 Provider 接口 + RSS 通用解析,
// 当前实现 google(Google News RSS) 与 bing(Bing News RSS), 均免 API Key

package news_client

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"news_ai_server/pkg/llm"
)

// NewsItem 检索到的一条新闻
type NewsItem struct {
	Provider    string     `json:"provider"`     // 检索来源 google/bing
	URL         string     `json:"url"`          // 原文链接
	Title       string     `json:"title"`        // 标题
	Snippet     string     `json:"snippet"`      // 摘要
	Author      string     `json:"author"`       // 作者
	SourceName  string     `json:"source_name"`  // 媒体名
	PublishedAt *time.Time `json:"published_at"` // 发布时间
}

// Provider 新闻检索接口, 每个渠道一个实现(参照 stock_ai futu_client 单例风格)
type Provider interface {
	Name() string
	// Search 按关键词检索, limit 为最大返回条数
	Search(ctx context.Context, query string, limit int) ([]NewsItem, error)
}

// NewProvider 按名称构造检索渠道(zhipu 渠道需要 llm 客户端与检索模型)
func NewProvider(name string, httpClient *http.Client, userAgent string, llmClient *llm.Client, searchModel string) (Provider, error) {
	switch strings.ToLower(name) {
	case "zhipu":
		return &zhipuSearch{llm: llmClient, model: searchModel}, nil
	case "google":
		return &googleNews{base: baseHTTP{httpClient: httpClient, userAgent: userAgent}}, nil
	case "bing":
		return &bingNews{base: baseHTTP{httpClient: httpClient, userAgent: userAgent}}, nil
	default:
		return nil, fmt.Errorf("未知的新闻检索渠道: %s", name)
	}
}

// ---------------------------------------------------------------------------
// RSS 通用结构与解析
// ---------------------------------------------------------------------------

// rssFeed RSS 文档
type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

// rssChannel RSS channel
type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

// rssItem RSS item
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	Source      string `xml:"source"`
	PubDate     string `xml:"pubDate"`
}

// baseHTTP 各 Provider 共享的 HTTP 基础能力
type baseHTTP struct {
	httpClient *http.Client
	userAgent  string
}

// fetchRSS 拉取并解析 RSS
func (b *baseHTTP) fetchRSS(ctx context.Context, rawURL string) (*rssFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 RSS 请求失败: %w", err)
	}
	req.Header.Set("User-Agent", b.userAgent)
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 RSS 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS HTTP %d: %s", resp.StatusCode, rawURL)
	}
	var feed rssFeed
	if err = xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("解析 RSS 失败: %w", err)
	}
	return &feed, nil
}

// toItems RSS item 转统一结构
func toItems(provider string, items []rssItem, limit int) []NewsItem {
	out := make([]NewsItem, 0, len(items))
	for _, it := range items {
		if len(out) >= limit {
			break
		}
		if it.Link == "" || it.Title == "" {
			continue
		}
		out = append(out, NewsItem{
			Provider:    provider,
			URL:         strings.TrimSpace(it.Link),
			Title:       strings.TrimSpace(it.Title),
			Snippet:     strings.TrimSpace(StripTags(it.Description)),
			Author:      strings.TrimSpace(it.Author),
			SourceName:  strings.TrimSpace(it.Source),
			PublishedAt: parsePubDate(it.PubDate),
		})
	}
	return out
}

// parsePubDate 兼容常见 RSS 时间格式
func parsePubDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// escapeQuery URL 编码检索词
func escapeQuery(q string) string { return url.QueryEscape(q) }
