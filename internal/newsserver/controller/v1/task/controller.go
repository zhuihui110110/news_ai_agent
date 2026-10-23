// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 任务 Controller: 创建/列表/详情/重试/取消

package task

import (
	"fmt"
	"strconv"

	restful "github.com/emicklei/go-restful/v3"

	msg "news_ai_server/internal/newsserver/app/message"
	"news_ai_server/internal/newsserver/app/service"
	"news_ai_server/internal/newsserver/controller/v1/base"
)

// Controller 任务控制器
type Controller struct {
	base.BaseController
	svc *service.Service
}

// NewController 构造
func NewController(svc *service.Service) *Controller {
	return &Controller{svc: svc}
}

// Create 创建任务
// POST /v1/task/create
func (c *Controller) Create(req *restful.Request, resp *restful.Response) {
	in := &msg.CreateTaskReq{}
	if err := req.ReadEntity(in); err != nil {
		 base.Failed(resp, err)
		return
	}
	out, err := c.svc.CreateTask(req.Request.Context(), in)
	if err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, out)
}

// List 任务列表
// GET /v1/task/list
func (c *Controller) List(req *restful.Request, resp *restful.Response) {
	in := &msg.TaskListReq{
		Topic:    req.QueryParameter("topic"),
		Page:     queryInt(req, "page"),
		PageSize: queryInt(req, "page_size"),
	}
	if s := req.QueryParameter("status"); s != "" {
		st, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			 base.Failed(resp, fmt.Errorf("status 非法: %w", err))
			return
		}
		v := int32(st)
		in.Status = &v
	}
	out, err := c.svc.ListTasks(req.Request.Context(), in)
	if err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, out)
}

// Detail 任务详情
// GET /v1/task/detail?id=1
func (c *Controller) Detail(req *restful.Request, resp *restful.Response) {
	id, err := strconv.ParseInt(req.QueryParameter("id"), 10, 64)
	if err != nil || id <= 0 {
		 base.Failed(resp, fmt.Errorf("id 非法"))
		return
	}
	out, err := c.svc.TaskDetail(req.Request.Context(), id)
	if err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, out)
}

// Retry 人工重试
// POST /v1/task/retry {"id":1}
func (c *Controller) Retry(req *restful.Request, resp *restful.Response) {
	in := &msg.TaskIDReq{}
	if err := req.ReadEntity(in); err != nil {
		 base.Failed(resp, err)
		return
	}
	if err := c.svc.RetryTask(req.Request.Context(), in.ID); err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, nil)
}

// Cancel 人工取消
// POST /v1/task/cancel {"id":1}
func (c *Controller) Cancel(req *restful.Request, resp *restful.Response) {
	in := &msg.TaskIDReq{}
	if err := req.ReadEntity(in); err != nil {
		 base.Failed(resp, err)
		return
	}
	if err := c.svc.CancelTask(req.Request.Context(), in.ID); err != nil {
		 base.Failed(resp, err)
		return
	}
	 base.Success(resp, nil)
}
