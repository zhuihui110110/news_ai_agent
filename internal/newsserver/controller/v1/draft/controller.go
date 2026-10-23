// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 草稿 Controller: 列表(人工审核) + 确认发布

package draft

import (
	"fmt"
	"strconv"

	restful "github.com/emicklei/go-restful/v3"

	msg "news_ai_server/internal/newsserver/app/message"
	"news_ai_server/internal/newsserver/app/service"
	"news_ai_server/internal/newsserver/controller/v1/base"
)

// Controller 草稿控制器
type Controller struct {
	base.BaseController
	svc *service.Service
}

// NewController 构造
func NewController(svc *service.Service) *Controller {
	return &Controller{svc: svc}
}

// List 草稿列表
// GET /v1/draft/list?publish_status=1&limit=20
func (c *Controller) List(req *restful.Request, resp *restful.Response) {
	in := &msg.DraftListReq{Limit: queryInt(req, "limit")}
	if s := req.QueryParameter("publish_status"); s != "" {
		st, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			 base.Failed(resp, err)
			return
		}
		v := int32(st)
		in.PublishStatus = &v
	}
	out, err := c.svc.ListDrafts(req.Request.Context(), in)
	if err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, out)
}

// Publish 人工确认发布
// POST /v1/draft/publish {"id":1}
func (c *Controller) Publish(req *restful.Request, resp *restful.Response) {
	in := &msg.DraftIDReq{}
	if err := req.ReadEntity(in); err != nil {
		 base.Failed(resp, err)
		return
	}
	if in.ID <= 0 {
		 base.Failed(resp, fmt.Errorf("id 非法"))
		return
	}
	out, err := c.svc.PublishDraft(req.Request.Context(), in.ID)
	if err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, out)
}

// queryInt 查询参数转 int
func queryInt(req *restful.Request, key string) int {
	n, _ := strconv.Atoi(req.QueryParameter(key))
	return n
}
