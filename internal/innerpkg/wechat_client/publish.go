// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 发布接口(freepublish): 仅在人工确认后由本系统提交发布,
// 提交后为异步任务, 通过 publish_id 轮询结果

package wechat_client

import (
	"context"
	"net/url"
)

// PublishStatus 发布状态语义(freepublish/get 返回值):
// 0=成功 1=发布过程中 2>=发布失败(原文: 2 可能为审核不通过/撤销等)
const (
	PublishResultSuccess   = 0
	PublishResultDoing     = 1
	PublishResultFailedMin = 2
)

// submitPublishResp 提交发布响应
type submitPublishResp struct {
	baseResp
	PublishID string `json:"publish_id"`
}

// SubmitPublish 提交草稿发布(人工确认动作触发), 返回 publish_id
func (c *Client) SubmitPublish(ctx context.Context, mediaID string) (string, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("access_token", token)

	var out submitPublishResp
	if err = c.postJSON(ctx, "/cgi-bin/freepublish/submit", q, map[string]string{"media_id": mediaID}, &out); err != nil {
		return "", err
	}
	if err = checkErr("提交发布", out.baseResp); err != nil {
		return "", err
	}
	return out.PublishID, nil
}

// PublishResult 发布结果
type PublishResult struct {
	PublishStatus int    // 0=成功 1=发布中 2+=失败
	ArticleURL    string // 成功后的文章链接
	FailMsg       string // 失败/审核信息
}

// getPublishResp 查询发布响应
type getPublishResp struct {
	baseResp
	PublishStatus int `json:"publish_status"`
	ArticleID     string `json:"article_id"`
	ArticleDetail struct {
		Count int `json:"count"`
		Item  []struct {
			ArticleURL string `json:"article_url"`
			Idx        int    `json:"idx"`
		} `json:"item"`
	} `json:"article_detail"`
	FailIdx []int `json:"fail_idx"`
}

// GetPublishResult 查询发布结果
func (c *Client) GetPublishResult(ctx context.Context, publishID string) (*PublishResult, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("access_token", token)

	var out getPublishResp
	if err = c.postJSON(ctx, "/cgi-bin/freepublish/get", q, map[string]string{"publish_id": publishID}, &out); err != nil {
		return nil, err
	}
	if err = checkErr("查询发布结果", out.baseResp); err != nil {
		return nil, err
	}
	res := &PublishResult{PublishStatus: out.PublishStatus}
	if len(out.ArticleDetail.Item) > 0 {
		res.ArticleURL = out.ArticleDetail.Item[0].ArticleURL
	}
	if len(out.FailIdx) > 0 {
		res.FailMsg = "存在审核未通过的文章"
	}
	return res, nil
}
