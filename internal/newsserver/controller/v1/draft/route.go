// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 草稿路由注册

package draft

import (
	restful "github.com/emicklei/go-restful/v3"
)

// Registry 注册草稿路由
func Registry(ws *restful.WebService, c *Controller) {
	ws.Route(ws.GET("/v1/draft/list").To(c.List).
		Doc("草稿列表(人工审核用)"))
	ws.Route(ws.POST("/v1/draft/publish").To(c.Publish).
		Doc("人工确认发布草稿到公众号"))
}
