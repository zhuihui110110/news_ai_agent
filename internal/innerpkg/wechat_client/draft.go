// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 草稿箱接口: 自动创建草稿(/cgi-bin/draft/add), 之后由人工在公众号后台
// 或本系统确认发布, 全流程不触碰"群发直接推送"能力

package wechat_client

import (
	"context"
	"fmt"
	"net/url"
)

// DraftArticle 草稿文章字段
type DraftArticle struct {
	Title            string `json:"title"`                       // 标题
	Author           string `json:"author,omitempty"`            // 作者
	Digest           string `json:"digest,omitempty"`            // 摘要
	Content          string `json:"content"`                     // 正文 HTML
	ContentSourceURL string `json:"content_source_url,omitempty"` // 原文链接
	ThumbMediaID     string `json:"thumb_media_id"`              // 封面图永久素材id
	NeedOpenComment  int32  `json:"need_open_comment,omitempty"` // 是否打开评论
	OnlyFansCanComment int32 `json:"only_fans_can_comment,omitempty"`
}

// addDraftResp 创建草稿响应
type addDraftResp struct {
	baseResp
	MediaID string `json:"media_id"`
}

// AddDraft 在草稿箱新建一篇文章, 返回草稿 media_id
func (c *Client) AddDraft(ctx context.Context, article *DraftArticle) (string, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("access_token", token)

	body := map[string]interface{}{"articles": []*DraftArticle{article}}
	var out addDraftResp
	if err = c.postJSON(ctx, "/cgi-bin/draft/add", q, body, &out); err != nil {
		return "", err
	}
	if err = checkErr("创建草稿", out.baseResp); err != nil {
		return "", err
	}
	if out.MediaID == "" {
		return "", fmt.Errorf("创建草稿未返回 media_id")
	}
	return out.MediaID, nil
}

// deleteDraftResp 删除草稿响应
type deleteDraftResp struct {
	baseResp
}

// DeleteDraft 删除草稿(重试前清理旧草稿用)
func (c *Client) DeleteDraft(ctx context.Context, mediaID string) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("access_token", token)
	var out deleteDraftResp
	if err = c.postJSON(ctx, "/cgi-bin/draft/delete", q, map[string]string{"media_id": mediaID}, &out); err != nil {
		return err
	}
	return checkErr("删除草稿", out.baseResp)
}
