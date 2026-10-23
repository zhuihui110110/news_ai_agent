// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: newsserver 是新闻 AI Agent 服务入口
// (新闻检索 -> 正文抓取 -> GLM 生成 -> 公众号草稿 -> 人工确认发布)

package main

import "news_ai_server/internal/newsserver"

func main() {
	newsserver.NewApp("newsserver").Run()
}
