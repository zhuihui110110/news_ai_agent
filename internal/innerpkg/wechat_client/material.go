// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 永久素材接口: 新建草稿必须使用永久素材的 thumb_media_id 作为封面,
// 因此封面图需先经 /cgi-bin/material/add_material 上传

package wechat_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// materialResp 上传素材响应
type materialResp struct {
	baseResp
	MediaID string `json:"media_id"`
	URL     string `json:"url"`
}

// UploadMaterialImageBytes 上传永久图片素材(封面图), 返回 media_id
func (c *Client) UploadMaterialImageBytes(ctx context.Context, filename string, data []byte) (string, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("access_token", token)
	q.Set("type", "image")

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("media", filename)
	if err != nil {
		return "", fmt.Errorf("构造 multipart 失败: %w", err)
	}
	if _, err = fw.Write(data); err != nil {
		return "", err
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/cgi-bin/material/add_material?"+q.Encode(), &buf)
	if err != nil {
		return "", fmt.Errorf("构造上传请求失败: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传素材失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	var out materialResp
	if err = json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("解析上传素材响应失败: %w, body: %s", err, string(raw))
	}
	if err = checkErr("上传素材", out.baseResp); err != nil {
		return "", err
	}
	if out.MediaID == "" {
		return "", fmt.Errorf("上传素材未返回 media_id: %s", string(raw))
	}
	return out.MediaID, nil
}

// DownloadImage 下载图片字节(用于将新闻封面转为公众号素材)
func (c *Client) DownloadImage(ctx context.Context, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载封面图失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载封面图 HTTP %d: %s", resp.StatusCode, imageURL)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 10<<20))
}
