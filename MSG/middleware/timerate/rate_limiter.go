package timerate

import (
	"MSG/middleware/router"
	"fmt"
	"golang.org/x/time/rate"
)

// 网关集成限流功能
func RateLimiter(params ...int) func(c *router.SliceRouteContext) {
	var r rate.Limit = 1
	var b = 2
	if len(params) == 2 {
		r = rate.Limit(params[0])
		b = params[1]
	}
	l := rate.NewLimiter(r, b)

	return func(c *router.SliceRouteContext) {
		// 如果无法获取到token，则跳出中间件，直接返回
		if !l.Allow() {
			c.Rw.Write([]byte(fmt.Sprintf("rate limie : %v,%v", l.Limit(), l.Burst())))
			c.Abort() // 设置标志，表明被限流了
			return
		}
		// 可以获取到token，执行中间件
		c.Next()
	}
}
