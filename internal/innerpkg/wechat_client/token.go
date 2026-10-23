// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: access_token 管理: 使用 stable_token 接口获取, 进程内缓存并提前 5 分钟刷新;
// 线上注意在公众号后台配置服务器 IP 白名单

package wechat_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// tokenRefreshLead 提前刷新时间
const tokenRefreshLead = 5 * time.Minute

// tokenRefreshing 单飞标记, 避免并发重复刷新
var tokenRefreshing atomic.Bool

// GetAccessToken 获取可用 access_token(带缓存)
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.token != "" && time.Now().Unix() < c.tokenExp-int64(tokenRefreshLead.Seconds()) {
		tok := c.token
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()

	if !tokenRefreshing.CompareAndSwap(false, true) {
		// 已有并发刷新, 简单等待后读缓存
		time.Sleep(500 * time.Millisecond)
		c.mu.Lock()
		tok := c.token
		c.mu.Unlock()
		if tok != "" {
			return tok, nil
		}
		return "", fmt.Errorf("access_token 刷新中, 请稍后重试")
	}
	defer tokenRefreshing.Store(false)
	return c.forceRefreshToken(ctx, false)
}

// forceRefreshToken 强制刷新(fail 时可带 force_refresh 重试一次)
func (c *Client) forceRefreshToken(ctx context.Context, force bool) (string, error) {
	body := map[string]interface{}{
		"grant_type":    "client_credential",
		"appid":         c.appID,
		"secret":        c.appSecret,
		"force_refresh": force,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/cgi-bin/stable_token", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("构造 stable_token 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 stable_token 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("读取 stable_token 响应失败: %w", err)
	}
	var out struct {
		baseResp
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err = json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("解析 stable_token 响应失败: %w, body: %s", err, string(raw))
	}
	if out.Errcode != 0 {
		return "", fmt.Errorf("获取 access_token 失败[%d]: %s", out.Errcode, out.Errmsg)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("获取 access_token 失败: %s", string(raw))
	}
	c.mu.Lock()
	c.token = out.AccessToken
	c.tokenExp = time.Now().Unix() + out.ExpiresIn
	c.mu.Unlock()
	return out.AccessToken, nil
}
