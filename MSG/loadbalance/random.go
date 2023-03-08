package loadbalance

import (
	"errors"
	"math/rand"
	"strings"
)

// 轮询算法
type RandomBalance struct {
	// 服务器主机地址
	servAddrs []string

	conf LoadBalanceConf
}

func (r *RandomBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("The param's length is 0! It should at least 1!s")
	}
	for i := 0; i < len(params); i++ {
		r.servAddrs = append(r.servAddrs, params[i])
	}
	return nil
}

func (r *RandomBalance) Next() string {
	lens := len(r.servAddrs)
	if lens == 0 {
		return ""
	}
	return r.servAddrs[rand.Intn(len(r.servAddrs))]
}

func (c *RandomBalance) Update() {
	if conf, ok := c.conf.(*LoadBalanceZkConf); ok {
		c.servAddrs = []string{}
		for _, ip := range conf.GetConf() {
			c.Add(strings.Split(ip, ",")...)
		}
	}
}

func (c *RandomBalance) SetConf(conf LoadBalanceConf) {

}

func (c *RandomBalance) Get(s string) (string, error) {
	return c.Next(), nil
}
