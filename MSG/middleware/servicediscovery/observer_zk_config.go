package servicediscovery

import (
	"MSG/middleware/servicediscovery/zookeeper"
	"fmt"
)

// 负载均衡配置「抽象主体」
type LoadBalanceConf interface {
	Attach(o Observer)
	GetConf() []string
	WatchConf()
	UpdateConf(conf []string)
}

// 负载均衡配置「具体主体」
type LoadBalanceZkConf struct {
	observers    []Observer        // 观察者列表
	path         string            // zk的path地址
	zhHosts      []string          // zk的集群列表
	confIpWeight map[string]string // IP与权重的映射表 Ip -> Weight
	activeList   []string          // 可用主机列表
	format       string            // 格式化
}

func NewLoadBalanceZkConf(format, path string, zhHosts []string, conf map[string]string) (*LoadBalanceZkConf, error) {
	zkManager := zookeeper.NewZkManager(zhHosts)
	zkManager.GetConnect()
	defer zkManager.Close()
	zList, err := zkManager.GetServerListByPath(path) // path作为父结点
	if err != nil {
		return nil, err
	}
	mConf := &LoadBalanceZkConf{format: format, activeList: zList, confIpWeight: conf, zhHosts: zhHosts, path: path}
	//启动监听
	mConf.WatchConf()
	return mConf, nil
}

func (s *LoadBalanceZkConf) Attach(o Observer) {
	s.observers = append(s.observers, o)
}
func (s *LoadBalanceZkConf) GetConf() []string {
	return s.activeList
}

// 监听当前结点的所有下级结点的变化
func (s *LoadBalanceZkConf) WatchConf() {
	zkManager := zookeeper.NewZkManager(s.zhHosts)
	zkManager.GetConnect()
	// 等待当前结点的下游结点发生变化，里面开启了一个协程
	chanList, chanErr := zkManager.WatchServerListByPath(s.path)
	go func() {
		defer zkManager.Close()
		for {
			select {
			case changeErr := <-chanErr:
				fmt.Println("changeErr", changeErr)
			case changedList := <-chanList:
				fmt.Println("watch node changed")
				s.UpdateConf(changedList)
			}
		}
	}()
}

func (s *LoadBalanceZkConf) UpdateConf(conf []string) {
	// 所有的观察者进行更新
	s.activeList = conf
	for _, obs := range s.observers {
		obs.Update()
	}
}

// 观察者接口
type Observer interface {
	Update()
}

// 具体观察者
type LoadBalanceObserver struct {
	zkConf *LoadBalanceZkConf
}

// 根据指定名称返回一个具体观察者实例
func NewLoadBalanceObserver(conf *LoadBalanceZkConf) *LoadBalanceObserver {
	return &LoadBalanceObserver{
		zkConf: conf,
	}
}
