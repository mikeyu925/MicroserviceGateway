package loadbalance

import (
	"errors"
	"strings"
)

// 轮询算法
type RoundRobinBalance struct {
	// 服务器主机地址
	servAddrs []string
	// 当前轮询的结点索引
	curIndex int

	conf LoadBalanceConf
}

func (r *RoundRobinBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("The param's length is 0! It should at least 1!s")
	}
	for i := 0; i < len(params); i++ {
		r.servAddrs = append(r.servAddrs, params[i])
	}
	return nil
}

func (r *RoundRobinBalance) Next() string {
	lens := len(r.servAddrs)
	if lens == 0 {
		return ""
	}
	addr := r.servAddrs[r.curIndex]
	r.curIndex = (r.curIndex + 1) % lens
	return addr
}

func (c *RoundRobinBalance) Update() {
	if conf, ok := c.conf.(*LoadBalanceZkConf); ok {
		c.servAddrs = []string{}
		for _, ip := range conf.GetConf() {
			c.Add(strings.Split(ip, ",")...)
		}
	}
}

func (c *RoundRobinBalance) SetConf(conf LoadBalanceConf) {

}

func (c *RoundRobinBalance) Get(s string) (string, error) {
	return c.Next(), nil
}
