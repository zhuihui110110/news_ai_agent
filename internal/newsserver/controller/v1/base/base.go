// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: Controller 公共基类与统一应答封装(独立包避免与资源子包形成 import 环)

package base

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
)

// Response 统一应答结构
type Response struct {
	Code    int         `json:"code"` // 0=成功
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BaseController 控制器基类(各资源 Controller 内嵌)
type BaseController struct{}

// Success 成功应答
func Success(resp *restful.Response, data interface{}) {
	writeJSON(resp, http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Failed 失败应答
func Failed(resp *restful.Response, err error) {
	writeJSON(resp, http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
}

// writeJSON 写 JSON
func writeJSON(resp *restful.Response, status int, body interface{}) {
	resp.WriteHeader(status)
	resp.WriteAsJson(body)
}
