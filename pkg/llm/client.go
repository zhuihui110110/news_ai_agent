// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 大模型直连客户端(智谱 GLM /chat/completions, 参照 stock_ai pkg/dify 客户端风格:
// Go 编排层直接调用大模型 API, 不经过任何中间平台)

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"news_ai_server/internal/innerpkg/options"
)

// Client GLM 客户端
type Client struct {
	baseURL     string
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	httpClient  *http.Client
}

// NewClient 由配置构造客户端
func NewClient(opts *options.LLMOptions) *Client {
	return &Client{
		baseURL:     strings.TrimRight(opts.BaseURL, "/"),
		apiKey:      opts.APIKey,
		model:       opts.Model,
		temperature: opts.Temperature,
		maxTokens:   opts.MaxTokens,
		httpClient:  &http.Client{Timeout: opts.Timeout},
	}
}

// Model 当前使用的模型名
func (c *Client) Model() string { return c.model }

// Ready 是否具备调用条件(api-key 已配置)
func (c *Client) Ready() bool { return c.apiKey != "" }

// Chat 发起一次对话补全(阻塞), 返回抽取后的正文与 token 用量
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}
	if req.Temperature == 0 {
		req.Temperature = c.temperature
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM API 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API HTTP %d: %s", resp.StatusCode, truncate(string(data), 512))
	}
	var apiResp apiChatResponse
	if err = json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, truncate(string(data), 512))
	}
	if apiResp.Error != nil {
		return nil, apiResp.Error
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空 choices: %s", truncate(string(data), 512))
	}
	choice := apiResp.Choices[0]
	content := CleanThinkingContent(choice.Message.Content)
	// 思考型模型(如 glm-5.3): 思维链计入 completion tokens,
	// 若正文为空需给出可定位的错误而不是把空串交给下游;
	// 携带 web_search 工具的搜索调用以搜索结果为准, 正文为空不视为错误
	if content == "" && len(apiResp.WebSearch) == 0 {
		if choice.FinishReason == "length" {
			return nil, fmt.Errorf("LLM 输出被截断(finish_reason=length), 思考链耗尽 max_tokens=%d, 请增大 llm.max-tokens", req.MaxTokens)
		}
		return nil, fmt.Errorf("LLM 未返回正文内容(仅思考链), finish_reason=%s", choice.FinishReason)
	}
	return &ChatResponse{
		Content:          content,
		FinishReason:     choice.FinishReason,
		PromptTokens:     apiResp.Usage.PromptTokens,
		CompletionTokens: apiResp.Usage.CompletionTokens,
		TotalTokens:      apiResp.Usage.TotalTokens,
		Model:            req.Model,
		WebSearch:        apiResp.WebSearch,
	}, nil
}

// thinkingRe 匹配 GLM 思考链输出
var thinkingRe = regexp.MustCompile(`(?s)<think>.*?</think>`)

// CleanThinkingContent 去除 <think>...</think> 思考内容(参照 dify client 实现)
func CleanThinkingContent(s string) string {
	return strings.TrimSpace(thinkingRe.ReplaceAllString(s, ""))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
