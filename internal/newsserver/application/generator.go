// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 文章生成器: 组装素材 Prompt -> 直连调用 GLM(OpenAI 兼容 /chat/completions)
// -> 解析 JSON 产出(标题/摘要/正文HTML/标签)

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"news_ai_server/internal/domain/infrastructure/mysql/constants"
	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/pkg/llm"
)

// GeneratedArticle 生成结果(JSON 协议)
type GeneratedArticle struct {
	Title       string   `json:"title"`
	Digest      string   `json:"digest"`
	Tags        []string `json:"tags"`
	ContentHTML string   `json:"content_html"`
}

// GenerateArticle 调用大模型生成公众号文章
func GenerateArticle(ctx context.Context, client *llm.Client, topic string, articles []model.NewsSourceArticleModel) (*GeneratedArticle, *llm.ChatResponse, error) {
	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt()},
			{Role: "user", Content: buildUserPrompt(topic, articles)},
		},
	}
	resp, err := client.Chat(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("调用 GLM 生成失败: %w", err)
	}
	g, err := parseGeneratedJSON(resp.Content)
	if err != nil {
		return nil, nil, err
	}
	return g, resp, nil
}

// systemPrompt 系统提示词
func systemPrompt() string {
	return `你是资深科技媒体编辑, 为微信公众号撰写"每日热点速览"类文章。
要求:
1. 基于给定新闻素材综合成文, 不编造素材之外的事实;
2. 输出严格 JSON(不要输出多余文字/代码块标记), 结构:
   {"title":"标题(25字内)","digest":"摘要(120字内)","tags":["标签1","标签2"],"content_html":"正文HTML"}
3. content_html 仅使用 section/h3/p/strong/ul/li/blockquote 标签, 不使用图片;
4. 每条新闻一个小节: h3 小节标题 -> 一段事件概述 -> ul 列出 2~3 个关键点(strong 强调关键词);
5. 结尾用一个"编辑观察"小节(blockquote)给出整体趋势点评;
6. 语气专业克制, 面向科技从业者。`
}

// buildUserPrompt 用户提示词: 领域 + 素材列表
func buildUserPrompt(topic string, articles []model.NewsSourceArticleModel) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "新闻领域: %s(%s)\n今天的热点素材如下, 共%d条:\n\n",
		constants.TopicName[topic], topic, len(articles))
	for i := range articles {
		a := &articles[i]
		fmt.Fprintf(sb, "【素材%d】标题: %s\n来源: %s %s\n", i+1, a.Title, a.Provider, a.SourceName)
		if a.PublishedAt != nil {
			fmt.Fprintf(sb, "时间: %s\n", a.PublishedAt.Format("2006-01-02 15:04"))
		}
		body := a.Content
		if a.ContentStatus != constants.ContentStatusOK || body == "" {
			body = a.Snippet // 正文抓取失败时退化为摘要
		}
		if r := []rune(body); len(r) > 2000 {
			body = string(r[:2000])
		}
		fmt.Fprintf(sb, "内容: %s\n\n", body)
	}
	sb.WriteString("请综合以上素材, 按系统要求输出 JSON。")
	return sb.String()
}

// parseGeneratedJSON 解析模型输出(兼容 ```json 围栏)
func parseGeneratedJSON(content string) (*GeneratedArticle, error) {
	s := strings.TrimSpace(content)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// 截取首个 { 到最后一个 } 之间的内容, 容错模型前后缀说明
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			s = s[i : j+1]
		}
	}
	var g GeneratedArticle
	if err := json.Unmarshal([]byte(s), &g); err != nil {
		return nil, fmt.Errorf("解析生成结果 JSON 失败: %w, 原文前500字: %s", err, truncateStr(content, 500))
	}
	if g.Title == "" || g.ContentHTML == "" {
		return nil, fmt.Errorf("生成结果缺少 title/content_html")
	}
	return &g, nil
}
