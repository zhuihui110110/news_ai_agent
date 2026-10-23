// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 任务路由注册(参照 stock_ai controller/v1/<resource>/route.go 的 Registry 模式)

package task

import (
	"strconv"

	restful "github.com/emicklei/go-restful/v3"
)

// Registry 注册任务路由
func Registry(ws *restful.WebService, c *Controller) {
	ws.Route(ws.POST("/v1/task/create").To(c.Create).
		Doc("创建新闻任务(领域: livelihood/economy/ai/tech)"))
	ws.Route(ws.GET("/v1/task/list").To(c.List).
		Doc("任务分页列表"))
	ws.Route(ws.GET("/v1/task/detail").To(c.Detail).
		Doc("任务详情(含源新闻/生成文章/草稿)"))
	ws.Route(ws.POST("/v1/task/retry").To(c.Retry).
		Doc("人工重试任务"))
	ws.Route(ws.POST("/v1/task/cancel").To(c.Cancel).
		Doc("人工取消任务"))
}

// queryInt 查询参数转 int
func queryInt(req *restful.Request, key string) int {
	n, _ := strconv.Atoi(req.QueryParameter(key))
	return n
}
