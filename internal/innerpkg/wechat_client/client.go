// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 微信公众号 API 客户端(参照 stock_ai futu_client 的 Init/GetClient 单例风格):
// 覆盖 access_token 管理(stable_token, 进程内缓存)、永久素材上传、草稿箱、发布能力

package wechat_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"news_ai_server/internal/innerpkg/options"
)

var (
	client     *Client
	clientOnce sync.Once
)

// GetClient 获取客户端单例(需先 Init)
func GetClient() *Client {
	if client == nil {
		panic("wechat client 未初始化, 请先调用 Init")
	}
	return client
}

// Init 初始化客户端(幂等)
func Init(opts *options.WechatOptions) {
	clientOnce.Do(func() {
		client = newClient(opts)
	})
}

// newClient 构造客户端
func newClient(opts *options.WechatOptions) *Client {
	return &Client{
		baseURL:    opts.BaseURL,
		appID:      opts.AppID,
		appSecret:  opts.AppSecret,
		enabled:    opts.Enabled,
		httpClient: &http.Client{Timeout: opts.Timeout},
	}
}

// Client 微信公众号客户端
type Client struct {
	baseURL    string
	appID      string
	appSecret  string
	enabled    bool
	httpClient *http.Client

	mu       sync.Mutex
	token    string
	tokenExp int64 // token 过期时间(unix 秒)
}

// Enabled 是否启用真实公众号调用(未启用时业务层走 mock)
func (c *Client) Enabled() bool { return c.enabled }

// baseResp 微信通用错误码结构
type baseResp struct {
	Errcode int32  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// checkErr 统一 errcode 检查
func checkErr(op string, resp baseResp) error {
	if resp.Errcode != 0 {
		return fmt.Errorf("%s失败[%d]: %s", op, resp.Errcode, resp.Errmsg)
	}
	return nil
}

// postJSON JSON POST 统一入口
func (c *Client) postJSON(ctx context.Context, path string, query url.Values, body interface{}, out interface{}) error {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求失败: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 %s 失败: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s HTTP %d: %s", path, resp.StatusCode, string(data))
	}
	if err = json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("解析 %s 响应失败: %w, body: %s", path, err, string(data))
	}
	return nil
}
