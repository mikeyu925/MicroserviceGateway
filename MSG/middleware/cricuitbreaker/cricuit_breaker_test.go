package cricuitbreaker

import (
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// 配置一个熔断器
func TestHystrix(t *testing.T) {
	var hystrixName = "hystrixName"

	// 实现统计面板
	// 启动一个流服务器，统计熔断、降级、限流的结果，实时发送到：8070 服务器上，就可以通过dashboard查看
	hStreamHandler := hystrix.NewStreamHandler()
	hStreamHandler.Start()
	go http.ListenAndServe(":8070", hStreamHandler)

	hystrix.ConfigureCommand(hystrixName, hystrix.CommandConfig{
		Timeout:                1000,
		MaxConcurrentRequests:  1,
		SleepWindow:            5000,
		RequestVolumeThreshold: 1,
		ErrorPercentThreshold:  1,
	})

	for i := 0; i < 10000; i++ {
		// 异步调用 Go()
		// 同步调用 Do()
		err := hystrix.Do(hystrixName, func() error { // 业务逻辑
			// 错误测试
			if i == 0 {
				return errors.New("service error : " + strconv.Itoa(i))
			}
			log.Println("do service: " + strconv.Itoa(i))
			return nil
		}, func(err error) error { // 降级方法
			fmt.Println("here is plan B!") // 一般可以和其他的服务器进行断开，禁止转发「访问」
			return errors.New("fallback err:" + err.Error())
		})
		if err != nil {
			log.Println("hystrix err: " + err.Error())
			time.Sleep(time.Second)
			log.Println("sleep 1 second", i)
		}
	}
}
