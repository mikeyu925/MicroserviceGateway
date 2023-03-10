package loadbalance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 加权轮询算法
type WeightRoundRobinBalance struct {
	// 服务器主机地址
	servAddrs []*node
	// 当前轮询的结点索引
	curIndex int

	conf LoadBalanceConf
}

const (
	MaxFails    int           = 3
	FailTimeout time.Duration = time.Second * 3
)

/*
node 每个服务器结点有不同的权重，并且在每一轮访问后可能会发生变化
*/
type node struct {
	// 主机地址 host:port
	addr string
	// 初始化权重
	weight int
	// 结点当前的临时权重，每一轮都可能变化  currentWeight = currentWeight + effectiveWeight
	// 每一轮都选择权重最大的结点
	currentWeight int
	// 有效权重，默认与weight相同，每当发生故障时，有效权重-1
	effectiveWeight int
	// failTimeout内最大失败次数，如果到达，则在failTimeout内不能再被选择
	maxFails int
	// 指定超时到时见「用于衡量最大失败次数，也用于超时」
	failTimeout time.Duration
	// 失败的时间结点
	failTimes []time.Time // 类似一个滑动窗口
}

/*
添加带权重的服务器主机
格式： "host:port","weight","host:port","weight", ... ,"host:port","weight"
*/
func (r *WeightRoundRobinBalance) Add(params ...string) error {
	length := len(params)
	if length == 0 || length%2 == 1 {
		return errors.New("The param's length is error!")
	}
	for i := 0; i < length; i += 2 {
		addr := params[i]
		weight, err := strconv.ParseInt(params[i+1], 10, 32)
		if err != nil {
			return err
		}
		// 默认权重为1
		if weight <= 0 {
			weight = 1
		}
		n := node{
			addr:            addr,
			weight:          int(weight),
			currentWeight:   0,
			effectiveWeight: int(weight),
			maxFails:        MaxFails,
			failTimeout:     FailTimeout,
		}
		r.servAddrs = append(r.servAddrs, &n)
	}
	return nil
}

// 找到权重最大的下一个服务器
/*

为了避免每次都访问同一个服务器，每一轮选中之后，需要对其进行降权

*/

func (r *WeightRoundRobinBalance) Next() (string, error) {
	lens := len(r.servAddrs)
	if lens == 0 {
		return "", errors.New("Ne server address!")
	}
	var index int = 0
	var maxNode *node = nil
	effectiveWeightSum := 0
	// 循环计算每个服务器的权值，选择最大的返回；
	// 对选中的服务器进行降权 ：currentWeight - sum(effectiveWeight)
	for i, servNode := range r.servAddrs {
		// 计算每个服务器的权重: 临时权重 + 有效权重  「选中最大的临时权重结点」
		servNode.currentWeight += servNode.effectiveWeight
		if servNode.maxFails <= 0 { // 不能被选中
			refreshErrRecords(servNode)
			servNode.maxFails = MaxFails - len(servNode.failTimes)
			if servNode.maxFails <= 0 {
				fmt.Println(servNode.addr, " 进入小黑屋！")
				continue
			}
		}
		if maxNode == nil || servNode.currentWeight > maxNode.currentWeight {
			maxNode = servNode
			index = i
		}
		effectiveWeightSum += servNode.effectiveWeight
	}
	// 对选中结点进行降权
	maxNode.currentWeight -= effectiveWeightSum
	r.curIndex = index
	return maxNode.addr, nil
}

// 奖励与惩罚策略」
func (r *WeightRoundRobinBalance) CallBack(addr string, flag bool) {
	for i := 0; i < len(r.servAddrs); i++ {
		w := r.servAddrs[i]
		if w.addr == addr {
			if flag {
				// 防止有效权重超过初始权重
				if w.effectiveWeight < w.weight {
					w.effectiveWeight++
				}
			} else { // 访问服务器失败
				w.effectiveWeight--
				// 刷新错误时间表
				refreshErrRecords(w)
				// 添加当前错误时间点
				w.failTimes = append(w.failTimes, time.Now())
				// 更新时间段内错误次数
				w.maxFails = MaxFails - len(w.failTimes)
			}
			break
		}
	}
}

func refreshErrRecords(w *node) {
	now := time.Now()
	var i = 0
	for len(w.failTimes) > i && now.Sub(w.failTimes[i]) >= w.failTimeout {
		i += 1
	}
	w.failTimes = w.failTimes[i:]
}

func print(rb *WeightRoundRobinBalance, addr string) {
	fmt.Println(" 主机地址 \t\t\t当前权重\t有效权重")
	total := 0
	for j := 0; j < len(rb.servAddrs); j++ {
		w := rb.servAddrs[j]
		total += w.effectiveWeight
		cw := strconv.Itoa(w.currentWeight)
		ew := strconv.Itoa(w.effectiveWeight)
		if w.addr == addr {
			// 被选中的服务器高亮显示
			fmt.Printf("%c[1;0;31m%s%c[0m", 0x1B, addr, 0x1B)
		} else {
			fmt.Print(w.addr)
		}
		var str = "\t\t" + cw + "\t\t" + ew + "\t\t"
		fmt.Println(str)
	}
	fmt.Println("有效权重之和:\t\t\t\t" + strconv.Itoa(total))
}

func (c *WeightRoundRobinBalance) Update() {
	if conf, ok := c.conf.(*LoadBalanceZkConf); ok {
		c.servAddrs = nil
		for _, ip := range conf.GetConf() {
			c.Add(strings.Split(ip, ",")...)
		}
	}
}

func (c *WeightRoundRobinBalance) SetConf(conf LoadBalanceConf) {

}

func (c *WeightRoundRobinBalance) Get(s string) (string, error) {
	return c.Next()
}
