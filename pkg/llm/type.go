// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 大模型(智谱 GLM, OpenAI 兼容协议)类型定义

package llm

// Message 对话消息
type Message struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"`
}

// ChatRequest /chat/completions 请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"` // 智谱原生工具(如 web_search)
}

// Tool 工具声明(当前仅支持智谱 web_search)
type Tool struct {
	Type      string          `json:"type"` // web_search
	WebSearch *WebSearchTool  `json:"web_search,omitempty"`
}

// WebSearchTool 联网搜索工具参数
type WebSearchTool struct {
	Enable       bool   `json:"enable"`
	SearchQuery  string `json:"search_query,omitempty"`  // 指定搜索词(不填则由模型根据问题生成)
	SearchResult bool   `json:"search_result,omitempty"` // 响应中返回原始搜索结果列表
}

// WebSearchResult 原始搜索结果条目(响应顶层 web_search 数组)
type WebSearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Content     string `json:"content"`
	Media       string `json:"media"`
	PublishDate string `json:"publish_date"` // 如 2026-10-22
	Refer       string `json:"refer"`
}

// ChatRequest 的 HTTP 响应
type apiChatResponse struct {
	Choices []struct {
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"` // 思考型模型(如 glm-5.3)的思维链
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	WebSearch []WebSearchResult `json:"web_search"` // 联网搜索原始结果(search_result=true 时返回)
	Error     *APIError         `json:"error,omitempty"`
}

// APIError 服务端错误体
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return "LLM API错误[" + e.Code + "]: " + e.Message }

// ChatResponse 业务层响应(已抽取正文与用量)
type ChatResponse struct {
	Content          string
	FinishReason     string // stop=正常结束 length=被 max_tokens 截断
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Model            string
	WebSearch        []WebSearchResult // 联网搜索原始结果
}
