// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 新闻正文抓取器: GET 页面 -> 自动识别编码(GBK/UTF-8等) -> 解析 HTML
// 提取 <article> 主体文本(无 article 时退化为 body), 同时抽取 og:image 作为封面图;
// 仅做只读抓取, 遵守目标站点 robots 语义与 UA 标识

package news_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	maxBodyBytes = 5 << 20 // 单页最大 5MB
	maxTextLen   = 12000   // 正文截断长度
)

// ArticleContent 抓取到的正文
type ArticleContent struct {
	Title    string // 页面标题
	Text     string // 主体纯文本
	CoverURL string // og:image 封面图
}

// FetchArticle 抓取并解析一篇文章
func FetchArticle(ctx context.Context, httpClient *http.Client, userAgent, rawURL string) (*ArticleContent, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("抓取页面失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("页面 HTTP %d: %s", resp.StatusCode, rawURL)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("读取页面失败: %w", err)
	}
	// 编码识别与统一转 UTF-8
	utf8Data, err := decodeHTML(data, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(bytes.NewReader(utf8Data))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	ac := &ArticleContent{}
	ac.Title = findTitle(root)
	ac.CoverURL = findMetaContent(root, "og:image")

	main := findArticleNode(root)
	if main == nil {
		main = root
	}
	ac.Text = truncateText(extractText(main, ""), maxTextLen)
	if strings.TrimSpace(ac.Text) == "" {
		return nil, fmt.Errorf("未提取到正文: %s", rawURL)
	}
	return ac, nil
}

// decodeHTML 自动识别编码并转 UTF-8(支持 GBK/GB18030 等)
func decodeHTML(data []byte, contentType string) ([]byte, error) {
	enc, name, _ := charset.DetermineEncoding(data, contentType)
	if enc == nil {
		return data, nil
	}
	// charset 包对 gb2312 的识别按 gb18030 处理更稳
	if name == "gb2312" || name == "gbk" {
		enc = simplifiedchinese.GB18030
	}
	out, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), enc.NewDecoder()))
	if err != nil {
		return nil, fmt.Errorf("编码转换失败: %w", err)
	}
	return out, nil
}

// findTitle 提取 <title>
func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
		return strings.TrimSpace(n.FirstChild.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := findTitle(c); t != "" {
			return t
		}
	}
	return ""
}

// findMetaContent 提取 meta[property=xxx] 的 content(如 og:image)
func findMetaContent(n *html.Node, property string) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		isTarget, content := false, ""
		for _, a := range n.Attr {
			switch a.Key {
			case "property", "name":
				if a.Val == property {
					isTarget = true
				}
			case "content":
				content = a.Val
			}
		}
		if isTarget && content != "" {
			return content
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if v := findMetaContent(c, property); v != "" {
			return v
		}
	}
	return ""
}

// findArticleNode 优先取 <article>, 否则 <body>
func findArticleNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "article" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findArticleNode(c); r != nil {
			return r
		}
	}
	return nil
}

// skipTags 提取正文时跳过的标签
var skipTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "nav": true,
	"header": true, "footer": true, "aside": true, "form": true,
	"button": true, "iframe": true, "svg": true,
}

// blockTags 块级标签前换行
var blockTags = map[string]bool{
	"p": true, "div": true, "br": true, "li": true, "ul": true, "ol": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"tr": true, "table": true, "blockquote": true, "section": true,
}

// extractText 递归抽取可见文本
func extractText(n *html.Node, cur string) string {
	if n.Type == html.ElementNode && skipTags[n.Data] {
		return cur
	}
	if n.Type == html.TextNode {
		cur += n.Data
	}
	sep := ""
	if n.Type == html.ElementNode && blockTags[n.Data] {
		sep = "\n"
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cur = extractText(c, cur)
	}
	return cur + sep
}

// truncateText 截断正文
func truncateText(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ") // 规整空白
	lines := strings.Split(s, "\n")
	var b strings.Builder
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if b.Len()+len(l) > n {
			break
		}
		b.WriteString(l)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

// StripTags 去除 HTML 标签得到纯文本(RSS description 清洗用)
func StripTags(s string) string {
	if s == "" {
		return ""
	}
	tokenizer := html.NewTokenizer(strings.NewReader(s))
	var b strings.Builder
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			b.Write(tokenizer.Text())
		}
	}
	return strings.TrimSpace(b.String())
}
