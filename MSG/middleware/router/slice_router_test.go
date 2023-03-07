package router

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

// 1.创建路由器：
//
//	一个路由器，包含多个路由
//	每个路由都可以有多个处理器「回调函数」
//
// 2.构建URI路由中间件：使用路由器对每个请求URI构建路由中间件
// 3.构建方法数组「一些列的毁掉函数」，并整合到URI路由中间件
// 4.将路由器作为http服务的处理器
func TestSliceRouter(t *testing.T) {
	var addr = "127.0.0.1:8006"
	log.Println("Starting httpserver at " + addr)

	// 创建路由器
	sliceRouter := NewSliceRouter()
	routeRoot := sliceRouter.Group("/")
	routeRoot.Use(middleware_hello, func(context *SliceRouteContext) {
		fmt.Println("11111111")
	})

	var routerHandler http.Handler = NewSliceRouterHandler(uerDefineMiddleware, sliceRouter)
	http.ListenAndServe(addr, routerHandler)
}

func middleware_hello(sc *SliceRouteContext) {
	fmt.Println("Now is running hello middleware!")
	sc.Next()
	fmt.Println("GoodBye!")
}

func uerDefineMiddleware(sc *SliceRouteContext) http.Handler {
	return &userDefineHandler{}
}

type userDefineHandler struct {
}

func (uh *userDefineHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("heooooooooo"))
}
