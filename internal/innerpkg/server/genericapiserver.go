// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 通用 HTTP API 服务器(参照 stock_ai GenericAPIServer 精简):
// go-restful 容器 + panic恢复/请求日志/健康检查中间件, 支持优雅退出

package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	restful "github.com/emicklei/go-restful/v3"

	"news_ai_server/pkg/log"
)

// RouteRegistrar 路由注册函数(由各资源 controller 提供)
type RouteRegistrar func(ws *restful.WebService)

// GenericAPIServer 通用 API 服务器
type GenericAPIServer struct {
	addr      string
	server    *http.Server
	container *restful.Container
}

// NewGenericAPIServer 构造服务器
func NewGenericAPIServer(addr string) *GenericAPIServer {
	container := restful.NewContainer()
	container.RecoverHandler(recoverHandler)
	container.DoNotRecover(false)
	srv := &GenericAPIServer{addr: addr, container: container, server: &http.Server{
		Addr:    addr,
		Handler: container,
	}}
	srv.container.Add(healthWebService())
	return srv
}

// Container restful 容器
func (s *GenericAPIServer) Container() *restful.Container { return s.container }

// AddWebService 注册业务 WebService
func (s *GenericAPIServer) AddWebService(ws *restful.WebService) {
	s.container.Add(ws)
}

// healthWebService 健康检查
func healthWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/healthz")
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		_ = resp.WriteEntity("ok")
	}))
	return ws
}

// recoverHandler panic 恢复与日志
func recoverHandler(err interface{}, w http.ResponseWriter) {
	log.Errorw("http panic", "err", fmt.Sprintf("%v", err), "stack", string(debug.Stack()))
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"code":500,"message":"internal error"}`))
}

// LoggingMiddleware 请求日志中间件(容器 filter 形式)
func LoggingMiddleware(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)
	log.Infof("%s %s -> %d (%s)", req.Request.Method, req.Request.URL.Path, resp.StatusCode(), time.Since(start))
}

// Start 启动 HTTP 服务(阻塞)
func (s *GenericAPIServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("http server listening on %s\n", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}
