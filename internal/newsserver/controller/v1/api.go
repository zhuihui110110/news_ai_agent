// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: API WebService 汇总注册(参照 stock_ai controller/v1/api.go):
// 路径前缀 /newsserver/api, 挂载 task 与 draft 两组路由

package v1

import (
	restful "github.com/emicklei/go-restful/v3"

	"news_ai_server/internal/newsserver/app/service"
	"news_ai_server/internal/newsserver/controller/v1/draft"
	"news_ai_server/internal/newsserver/controller/v1/task"
)

// apiPath API 路径前缀
const apiPath = "/newsserver/api"

// NewAPIWebService 构造业务 WebService
func NewAPIWebService(svc *service.Service) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(apiPath).
		Consumes(restful.MIME_JSON, "application/x-www-form-urlencoded").
		Produces(restful.MIME_JSON)

	task.Registry(ws, task.NewController(svc))
	draft.Registry(ws, draft.NewController(svc))
	return ws
}
