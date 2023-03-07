package router

import (
	"context"
	"net/http"
	"strings"
)

const abortIndex int8 = 63

type HandlerFunc func(*SliceRouteContext)

// 路由器结构体 Router 「路由数组结构体」
type SliceRouter struct {
	groups []*sliceRoute
}

type sliceRoute struct {
	// 反向指针
	*SliceRouter
	// 请求路径
	path string
	// 请求处理器列表
	handlers []HandlerFunc
}

// SliceRouteContext 路由上下文
// 每个路由对应一个上下实例，同时维护请求和响应对象
type SliceRouteContext struct {
	*sliceRoute
	index int8 // 当前处理器执行到哪个位置
	Ctx   context.Context
	Req   *http.Request
	Rw    http.ResponseWriter
}

// NewSliceRouter 构造路由器实例
func NewSliceRouter() *SliceRouter {
	return &SliceRouter{}
}

// Group 根据指定路径构造路由
// 每个路由维护一个路由器指针
func (g *SliceRouter) Group(path string) *sliceRoute {
	// init sliceRoute
	return &sliceRoute{
		SliceRouter: g, // this指针
		path:        path,
	}
}

func (route *sliceRoute) Use(middlewares ...HandlerFunc) *sliceRoute {
	route.handlers = append(route.handlers, middlewares...)
	// 当前路由在路由器中是否存在
	flag := false
	for _, r := range route.SliceRouter.groups {
		if route == r {
			flag = true
			break
		}
	}
	if !flag {
		// 不存在，则添加
		route.SliceRouter.groups = append(route.SliceRouter.groups, route)
	}
	return route
}

// 定义处理器类型函数
// 接收 *SliceRouteContext 类型作为参数
// 返回 http.Handler 结果
type handler func(*SliceRouteContext) http.Handler

//	 SliceRouterHandler 方法数组路由器的核心处理器
//		维护一个方法数组路由器的指针：*SliceRouter
//		支持用户自定义处理器
type SliceRouterHandler struct {
	h      handler
	router *SliceRouter
}

// ServeHTTP 实现了 http.Handler 接口的方法
//
//	作为当前路由器的 http 服务的处理器入口
//	依次执行路由的处理函数
func (rh *SliceRouterHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// 初始化路由Context实例
	c := newSliceRouterContext(rw, req, rh.router)
	// 检查路由是否绑定用户自定义处理函数
	if rh.h != nil {
		c.handlers = append(c.handlers, func(routeContext *SliceRouteContext) {
			rh.h(c).ServeHTTP(c.Rw, c.Req)
		})
	}
	// 依次执行路由的处理函数
	c.Reset()
	c.Next()

}

// NewSliceRouterHandler 创建 http 服务的处理器
// 将实现了 http.Handler 接口的实例返回
func NewSliceRouterHandler(h handler, router *SliceRouter) *SliceRouterHandler {
	//  build http.handler instance with SliceRouter
	return &SliceRouterHandler{
		h:      h,
		router: router,
	}
}

// 初始化路由上下文实例
func newSliceRouterContext(rw http.ResponseWriter, req *http.Request, r *SliceRouter) *SliceRouteContext {
	// 初始化最长url匹配路由
	sr := &sliceRoute{}
	// 最长url前缀匹配
	matchUrlLen := 0
	for _, route := range r.groups {
		// uri匹配成功：前缀匹配
		if strings.HasPrefix(req.RequestURI, route.path) {
			// 记录最长匹配 uri
			pathLen := len(route.path)
			if pathLen > matchUrlLen {
				matchUrlLen = pathLen
				// 浅拷贝：拷贝数组指针
				*sr = *route
			}
		}
	}

	c := &SliceRouteContext{
		Rw:         rw,
		Req:        req,
		Ctx:        req.Context(),
		sliceRoute: sr}
	c.Reset() // 重置一下
	return c
}

func (c *SliceRouteContext) Get(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

func (c *SliceRouteContext) Set(key, val interface{}) {
	c.Ctx = context.WithValue(c.Ctx, key, val)
}

// Next 从最先加入中间件开始回调
func (c *SliceRouteContext) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		// 循环调用每一个handler
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 跳出中间件方法
func (c *SliceRouteContext) Abort() {
	c.index = abortIndex
}

// IsAborted 是否跳过了回调
func (c *SliceRouteContext) IsAborted() bool {
	return c.index >= abortIndex
}

// Reset 重置回调
func (c *SliceRouteContext) Reset() {
	c.index = -1
}
