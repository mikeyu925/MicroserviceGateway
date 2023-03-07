package timerate

import (
	"MSG/middleware/router"
	"MSG/proxy"
	"context"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)
import "golang.org/x/time/rate"

func TestRateLimiter(t *testing.T) {
	// 构建限速器
	// r l.Limit()获得
	// b l.Burst()获得
	l := rate.NewLimiter(1, 5)

	// 获取token
	for i := 0; i < 100; i++ {
		log.Println("before Wait", i)
		// 方式1：
		// 阻塞等待直到获取到token。最多等待2s
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)

		if err := l.Wait(ctx); err != nil {
			log.Println("Waiting Token Timeout!")
		}
		log.Println("After Wait")

		// 返回预计需要等待多久才有新的 token 「可以通过等待指定时间再执行任务
		r := l.Reserve()
		if !r.OK() {
			return
		}
		log.Println("Delay:", r.Delay())
		time.Sleep(r.Delay())

		log.Println("Allow:", l.Allow())
		time.Sleep(200 * time.Millisecond)
	}
}

func TestRateLimiter2(t *testing.T) {
	customHandler := func(c *router.SliceRouteContext) http.Handler {
		rs1 := "http://127.0.0.1:8001/"
		url1, err1 := url.Parse(rs1)
		if err1 != nil {
			log.Println(err1)
		}

		rs2 := "http://127.0.0.1:8002/haha"
		url2, err2 := url.Parse(rs2)
		if err2 != nil {
			log.Println(err2)
		}

		urls := []*url.URL{url1, url2}
		return proxy.NewMultipleHostsReverseProxy(c.Ctx, urls)
	}

	var addr = "127.0.0.1:8006"
	log.Println("Starting http server at:" + addr)

	sliceRouter := router.NewSliceRouter()
	sliceRouter.Group("/").Use(RateLimiter())
	var routerHandler http.Handler = router.NewSliceRouterHandler(customHandler, sliceRouter)
	http.ListenAndServe(addr, routerHandler)
}
